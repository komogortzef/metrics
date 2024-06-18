package server

import (
	"context"
	"io"
	"net/http"
	"strconv"

	log "metrics/internal/logger"
	m "metrics/internal/models"

	"github.com/tidwall/gjson"
	"go.uber.org/zap"
)

type helper func([]byte, []byte) ([]byte, error)

type Repository interface {
	Put(ctx context.Context, key string, data []byte, helps ...helper) error
	Get(ctx context.Context, key string) ([]byte, error)
	List(ctx context.Context) ([][]byte, error)
}

type MetricManager struct {
	Serv            *http.Server
	Store           Repository
	Address         string `env:"ADDRESS" envDefault:"none"`
	StoreInterval   int    `env:"STORE_INTERVAL" envDefault:"-1"`
	Restore         bool   `env:"RESTORE" envDefault:"true"`
	FileStoragePath string
	DBAddress       string `env:"DATABASE_DSN" envDefault:"none"`
}

func (mm *MetricManager) Run(ctx context.Context) {
	log.Info("Metric Manger configuration",
		zap.String("addr", mm.Address),
		zap.Int("store interval", mm.StoreInterval),
		zap.Bool("restore", mm.Restore),
		zap.String("file store path", mm.FileStoragePath),
		zap.String("data base config", mm.DBAddress))

	errChan := make(chan error, 1)
	go func() {
		err := mm.Serv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
		close(errChan)
	}()

	if mm.StoreInterval > 0 && mm.FileStoragePath != "" {
		dumpWait(ctx, mm.Store, mm.FileStoragePath, mm.StoreInterval)
	}

	select {
	case <-ctx.Done():
		if mm.FileStoragePath != m.NoStorage {
			if err := dump(ctx, mm.FileStoragePath, mm.Store); err != nil {
				log.Fatal("couldn't dump to file", zap.Error(err))
			}
		}
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
	metric, err := m.NewMetric(mtype, name, value)
	if err != nil {
		log.Warn("NewMetric error", zap.Error(err))
		http.Error(rw, m.BadRequestMessage, http.StatusBadRequest)
		return
	}
	bytes, err := metric.MarshalJSON()
	if err != nil {
		log.Warn("UpdateHandler(): marshal error", zap.Error(err))
		http.Error(rw, m.InternalErrorMsg, http.StatusInternalServerError)
		return
	}
	if err = mm.Store.Put(req.Context(), name, bytes, getHelper(mtype)); err != nil {
		log.Warn("UpdateHandler(): storage error", zap.Error(err))
		http.Error(rw, m.InternalErrorMsg, http.StatusInternalServerError)
		return
	}
	rw.WriteHeader(http.StatusOK)
}

func (mm *MetricManager) GetHandler(rw http.ResponseWriter, req *http.Request) {
	mtype, name, _ := processURL(req.URL.Path)
	newBytes, err := mm.Store.Get(req.Context(), name)
	if err != nil {
		log.Warn("GetHandler(): Coundn't fetch the metric from store")
		http.Error(rw, m.NotFoundMessage, http.StatusNotFound)
		return
	}
	var numStr string
	if mtype == m.Counter {
		numBytes := gjson.GetBytes(newBytes, m.Delta)
		numStr = strconv.FormatInt(numBytes.Int(), 10)
	} else {
		numBytes := gjson.GetBytes(newBytes, m.Value)
		numStr = strconv.FormatFloat(numBytes.Float(), 'f', -1, 64)
	}

	rw.WriteHeader(http.StatusOK)
	_, _ = rw.Write([]byte(numStr))
}

func (mm *MetricManager) GetAllHandler(rw http.ResponseWriter, req *http.Request) {
	list := make([]Item, 0, m.MetricsNumber)

	var metric m.Metrics
	metrics, _ := mm.Store.List(req.Context())
	for _, bytes := range metrics {
		_ = metric.UnmarshalJSON(bytes)
		list = append(list, Item{Met: metric.String()})
	}
	html, err := renderGetAll(list)
	if err != nil {
		log.Warn("GetAllHandler(): An error occured during html rendering")
		http.Error(rw, m.InternalErrorMsg, http.StatusInternalServerError)
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

	name := gjson.GetBytes(bytes, m.ID).String()
	mtype := gjson.GetBytes(bytes, m.Mtype).String()
	if err = mm.Store.Put(req.Context(), name, bytes, getHelper(mtype)); err != nil {
		log.Warn("UpdateJSON(): couldn't write to store", zap.Error(err))
		http.Error(rw, m.InternalErrorMsg, http.StatusInternalServerError)
		return
	}
	newBytes := bytes
	if mtype == m.Counter {
		name := gjson.GetBytes(bytes, m.Delta).String()
		newBytes, _ = mm.Store.Get(req.Context(), name)
	}

	rw.WriteHeader(http.StatusOK)
	_, _ = rw.Write(newBytes)
}

func (mm *MetricManager) GetJSON(rw http.ResponseWriter, req *http.Request) {
	bytes, err := io.ReadAll(req.Body)
	if err != nil {
		log.Warn("GetJSON(): Couldn't read request body")
		http.Error(rw, m.BadRequestMessage, http.StatusBadRequest)
		return
	}
	defer req.Body.Close()

	name := gjson.GetBytes(bytes, m.ID).String()
	log.Info("needed metric name", zap.String("name", name))
	if bytes, err = mm.Store.Get(req.Context(), name); err != nil {
		log.Warn("GetJSON(): No such metric in store", zap.Error(err))
		http.Error(rw, m.NotFoundMessage, http.StatusNotFound)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	_, _ = rw.Write(bytes)
}

func (mm *MetricManager) PingHandler(rw http.ResponseWriter, req *http.Request) {
	db, ok := mm.Store.(*DataBase)
	if !ok {
		log.Warn("PingHandler(): Invalid storage type for ping")
		http.Error(rw, m.InternalErrorMsg, http.StatusInternalServerError)
		return
	}
	if err := db.Ping(req.Context()); err != nil {
		log.Warn("PingHandler(): There is no connection to data base")
		http.Error(rw, m.InternalErrorMsg, http.StatusInternalServerError)
		return
	}

	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte("The connection is established!"))
}

func (mm *MetricManager) BatchHandler(rw http.ResponseWriter, req *http.Request) {
	b, err := io.ReadAll(req.Body)
	if err != nil {
		log.Warn("UpdatesJSON(): Couldn't read request body")
		http.Error(rw, m.BadRequestMessage, http.StatusBadRequest)
		return
	}
	defer req.Body.Close()

	if db, ok := mm.Store.(*DataBase); ok {
		if err = db.insertBatch(req.Context(), b); err != nil {
			log.Warn("UpdatesJSON(): couldn't send the batch", zap.Error(err))
			http.Error(rw, m.InternalErrorMsg, http.StatusBadRequest)
		}
	} else {
		gjson.ParseBytes(b).ForEach(func(key, value gjson.Result) bool {
			name := value.Get(m.ID).String()
			mtype := value.Get(m.Mtype).String()
			_ = mm.Store.Put(req.Context(), name, []byte(value.Raw), getHelper(mtype))
			return true
		})
	}
	rw.WriteHeader(http.StatusOK)
}
