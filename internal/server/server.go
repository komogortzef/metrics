package server

import (
	ctx "context"
	"io"
	"net/http"
	"strconv"

	log "metrics/internal/logger"
	s "metrics/internal/service"

	"github.com/pquerna/ffjson/ffjson"
	"go.uber.org/zap"
)

type MetricManager struct {
	Serv            *http.Server
	Store           Storage
	Address         string `env:"ADDRESS" envDefault:"none"`
	StoreInterval   int    `env:"STORE_INTERVAL" envDefault:"-1"`
	Restore         bool   `env:"RESTORE" envDefault:"true"`
	FileStoragePath string
	DBAddress       string `env:"DATABASE_DSN" envDefault:"none"`
}

func (mm *MetricManager) Run(ctx ctx.Context) {
	errChan := make(chan error, 1)
	go func() {
		err := mm.Serv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
		close(errChan)
	}()
	if mm.StoreInterval > 0 && mm.FileStoragePath != s.NoStorage {
		dumpWait(ctx, mm.Store, mm.FileStoragePath, mm.StoreInterval)
	}

	select {
	case <-ctx.Done():
		mm.Store.Close()
		if err := mm.Serv.Shutdown(ctx); err != nil {
			log.Fatal("server shutdown err", zap.Error(err))
		}
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
		http.Error(rw, s.BadRequestMessage, http.StatusBadRequest)
		return
	}
	if _, err = mm.Store.Put(req.Context(), metric); err != nil {
		log.Warn("UpdateHandler(): storage error", zap.Error(err))
		http.Error(rw, s.InternalErrorMsg, http.StatusInternalServerError)
		return
	}
	rw.WriteHeader(http.StatusOK)
}

func (mm *MetricManager) GetHandler(rw http.ResponseWriter, req *http.Request) {
	mtype, name, _ := processURL(req.URL.Path)
	met := &s.Metrics{ID: name, MType: mtype}
	metric, err := mm.Store.Get(req.Context(), met)
	if err != nil {
		log.Warn("GetHandler(): Coundn't fetch the metric from store", zap.Error(err))
		http.Error(rw, s.NotFoundMessage, http.StatusNotFound)
		return
	}
	var numStr string
	if mtype == s.Counter {
		numStr = strconv.FormatInt(*metric.Delta, 10)
	} else {
		numStr = strconv.FormatFloat(*metric.Value, 'f', -1, 64)
	}

	rw.WriteHeader(http.StatusOK)
	_, _ = rw.Write([]byte(numStr))
}

func (mm *MetricManager) GetAllHandler(rw http.ResponseWriter, req *http.Request) {
	list := make([]Item, 0, metricsNumber)
	metrics, _ := mm.Store.List(req.Context())
	for _, m := range metrics {
		list = append(list, Item{Met: m.String()})
	}
	html, err := renderGetAll(list)
	if err != nil {
		log.Warn("GetAllHandler(): An error occured during html rendering")
		http.Error(rw, s.InternalErrorMsg, http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "text/html")
	rw.WriteHeader(http.StatusOK)
	_, _ = rw.Write(html.Bytes())
}

func (mm *MetricManager) UpdateJSON(rw http.ResponseWriter, req *http.Request) {
	bytes, err := io.ReadAll(req.Body)
	if err != nil {
		log.Warn("Couldn't read with decompress")
	}
	defer req.Body.Close()

	metric := &s.Metrics{}
	_ = metric.UnmarshalJSON(bytes)
	if metric, err = mm.Store.Put(req.Context(), metric); err != nil {
		log.Warn("UpdateJSON(): couldn't write to store", zap.Error(err))
		http.Error(rw, s.InternalErrorMsg, http.StatusInternalServerError)
		return
	}

	bytes, _ = metric.MarshalJSON()
	rw.WriteHeader(http.StatusOK)
	_, _ = rw.Write(bytes)
}

func (mm *MetricManager) GetJSON(rw http.ResponseWriter, req *http.Request) {
	bytes, err := io.ReadAll(req.Body)
	if err != nil {
		log.Warn("GetJSON(): Couldn't read request body")
		http.Error(rw, s.BadRequestMessage, http.StatusBadRequest)
		return
	}
	defer req.Body.Close()

	metric := &s.Metrics{}
	_ = metric.UnmarshalJSON(bytes)
	if metric, err = mm.Store.Get(req.Context(), metric); err != nil {
		log.Warn("GetJSON(): No such metric in store", zap.Error(err))
		http.Error(rw, s.NotFoundMessage, http.StatusNotFound)
		return
	}

	bytes, _ = metric.MarshalJSON()
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	_, _ = rw.Write(bytes)
}

func (mm *MetricManager) PingHandler(rw http.ResponseWriter, req *http.Request) {
	if err := mm.Store.Ping(req.Context()); err != nil {
		log.Warn("ping error", zap.Error(err))
		http.Error(rw, s.InternalErrorMsg, http.StatusInternalServerError)
		return
	}
	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte("The connection is established!"))
}

func (mm *MetricManager) BatchHandler(rw http.ResponseWriter, req *http.Request) {
	b, err := io.ReadAll(req.Body)
	if err != nil {
		log.Warn("UpdatesJSON(): Couldn't read request body")
		http.Error(rw, s.BadRequestMessage, http.StatusBadRequest)
		return
	}
	defer req.Body.Close()

	var metrics []*s.Metrics
	if err = ffjson.Unmarshal(b, &metrics); err != nil {
		log.Warn("batchHandler(): unmarshal error", zap.Error(err))
		http.Error(rw, s.InternalErrorMsg, http.StatusInternalServerError)
		return
	}

	if err = mm.Store.PutBatch(req.Context(), metrics); err != nil {
		log.Warn("UpdatesJSON(): couldn't send the batch", zap.Error(err))
		http.Error(rw, s.InternalErrorMsg, http.StatusBadRequest)
		return
	}

	rw.WriteHeader(http.StatusOK)
}
