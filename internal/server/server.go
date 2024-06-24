package server

import (
	ctx "context"
	"errors"
	"io"
	"net/http"
	"strconv"

	log "metrics/internal/logger"
	s "metrics/internal/service"

	"github.com/pquerna/ffjson/ffjson"
	"go.uber.org/zap"
)

const (
	internalErrorMsg  = "internal server error"
	notFoundMessage   = "not found"
	badRequestMessage = "bad request"
)

type Storage interface {
	Put(ctx.Context, *s.Metrics) (*s.Metrics, error)
	Get(ctx.Context, *s.Metrics) (*s.Metrics, error)
	List(ctx.Context) ([]*s.Metrics, error)
	PutBatch(ctx.Context, []*s.Metrics) error
	Close()
}

type MetricManager struct {
	Storage
	http.Server
}

func (mm *MetricManager) Run(cx ctx.Context) {
	errChan := make(chan error, 1)
	go func() {
		err := mm.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChan <- err
		}
		close(errChan)
	}()

	fs, isFs := mm.Storage.(*FileStorage)
	if isFs {
		fs.dumpWait(cx)
	}
	select {
	case <-cx.Done():
		if isFs {
			if err := fs.dump(cx); err != nil {
				log.Warn("couldn't dump to file", zap.Error(err))
			}
		}
		mm.Storage.Close()
		_ = mm.Shutdown(cx)
		log.Debug("Goodbye!")
	case err := <-errChan:
		log.Fatal("server running error", zap.Error(err))
	}
}

func (mm *MetricManager) UpdateHandler(rw http.ResponseWriter, req *http.Request) {
	mtype, name, value := processURL(req.URL.Path)
	metric, err := s.NewMetric(mtype, name, value)
	if err != nil {
		log.Warn("NewMetric error", zap.Error(err))
		http.Error(rw, badRequestMessage, http.StatusBadRequest)
		return
	}
	if _, err = mm.Put(req.Context(), metric); err != nil {
		log.Warn("UpdateHandler(): storage error", zap.Error(err))
		http.Error(rw, internalErrorMsg, http.StatusInternalServerError)
		return
	}
	rw.WriteHeader(http.StatusOK)
}

func (mm *MetricManager) GetHandler(rw http.ResponseWriter, req *http.Request) {
	mtype, name, _ := processURL(req.URL.Path)
	met := &s.Metrics{ID: name, MType: mtype}
	metric, err := mm.Get(req.Context(), met)
	if errors.Is(err, ErrConnDB) {
		log.Warn("GetHandler(): storage error", zap.Error(err))
		http.Error(rw, internalErrorMsg, http.StatusInternalServerError)
		return
	} else if err != nil {
		log.Warn("GetHandler(): Coundn't fetch the metric from store", zap.Error(err))
		http.Error(rw, notFoundMessage, http.StatusNotFound)
		return
	}
	var numStr string
	if metric.IsCounter() {
		numStr = strconv.FormatInt(*metric.Delta, 10)
	} else {
		numStr = strconv.FormatFloat(*metric.Value, 'f', -1, 64)
	}
	rw.WriteHeader(http.StatusOK)
	_, _ = rw.Write([]byte(numStr))
}

func (mm *MetricManager) GetAllHandler(rw http.ResponseWriter, req *http.Request) {
	list := make([]Item, 0, metricsNumber)
	metrics, err := mm.List(req.Context())
	if errors.Is(err, ErrConnDB) {
		log.Warn("GetAllHandler(): storage error", zap.Error(err))
		http.Error(rw, internalErrorMsg, http.StatusInternalServerError)
		return
	}
	for _, m := range metrics {
		list = append(list, Item{Met: m.String()})
	}
	html, err := renderGetAll(list)
	if err != nil {
		log.Warn("GetAllHandler(): An error occured during html rendering")
		http.Error(rw, internalErrorMsg, http.StatusInternalServerError)
		return
	}
	rw.Header().Set("Content-Type", "text/html")
	rw.WriteHeader(http.StatusOK)
	_, _ = rw.Write(html.Bytes())
}

func (mm *MetricManager) UpdateJSON(rw http.ResponseWriter, req *http.Request) {
	log.Debug("UpdateJSON...")
	bytes, err := io.ReadAll(req.Body)
	if err != nil {
		log.Warn("Couldn't read with decompress")
	}
	defer req.Body.Close()

	metric := &s.Metrics{}
	_ = metric.UnmarshalJSON(bytes)
	if metric, err = mm.Put(req.Context(), metric); err != nil {
		log.Warn("UpdateJSON(): couldn't write to store", zap.Error(err))
		http.Error(rw, internalErrorMsg, http.StatusInternalServerError)
		return
	}
	bytes, _ = metric.MarshalJSON()
	rw.WriteHeader(http.StatusOK)
	_, _ = rw.Write(bytes)
}

func (mm *MetricManager) GetJSON(rw http.ResponseWriter, req *http.Request) {
	log.Debug("GetJSON...")
	bytes, err := io.ReadAll(req.Body)
	if err != nil {
		log.Warn("GetJSON(): Couldn't read request body")
		http.Error(rw, badRequestMessage, http.StatusBadRequest)
		return
	}
	defer req.Body.Close()

	metric := &s.Metrics{}
	_ = metric.UnmarshalJSON(bytes)
	if metric, err = mm.Get(req.Context(), metric); errors.Is(err, ErrConnDB) {
		log.Warn("GetJSON(): store error", zap.Error(err))
		http.Error(rw, internalErrorMsg, http.StatusInternalServerError)
		return
	} else if err != nil {
		log.Warn("GetJSON(): No such metric in store", zap.Error(err))
		http.Error(rw, notFoundMessage, http.StatusNotFound)
		return
	}
	bytes, _ = metric.MarshalJSON()
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	_, _ = rw.Write(bytes)
}

func (mm *MetricManager) PingHandler(rw http.ResponseWriter, req *http.Request) {
	if db, ok := mm.Storage.(*DataBase); ok {
		if err := db.Ping(req.Context()); err != nil {
			log.Warn("ping error", zap.Error(err))
			http.Error(rw, internalErrorMsg, http.StatusInternalServerError)
			return
		}
	}
	rw.WriteHeader(http.StatusOK)
	_, _ = rw.Write([]byte("The connection is established!"))
}

func (mm *MetricManager) BatchHandler(rw http.ResponseWriter, req *http.Request) {
	log.Debug("BatchHandler...")
	b, err := io.ReadAll(req.Body)
	if err != nil {
		log.Warn("BatchHandler(): Couldn't read request body")
		http.Error(rw, badRequestMessage, http.StatusBadRequest)
		return
	}
	defer req.Body.Close()

	var metrics []*s.Metrics
	if err = ffjson.Unmarshal(b, &metrics); err != nil {
		log.Warn("batchHandler(): unmarshal error", zap.Error(err))
		http.Error(rw, internalErrorMsg, http.StatusInternalServerError)
		return
	}
	if err = mm.PutBatch(req.Context(), metrics); err != nil {
		log.Warn("UpdatesJSON(): couldn't send the batch", zap.Error(err))
		http.Error(rw, internalErrorMsg, http.StatusBadRequest)
		return
	}
	rw.WriteHeader(http.StatusOK)
}
