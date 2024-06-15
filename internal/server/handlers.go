package server

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	log "metrics/internal/logger"
	m "metrics/internal/models"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tidwall/gjson"
	"go.uber.org/zap"
)

type (
	helper func([]byte, []byte) ([]byte, error) // для доп операций перед сохраненим в Store

	Repository interface {
		Put(key string, data []byte, helps ...helper) (int, error)
		Get(key string) ([]byte, error)
		List() ([][]byte, error)
	}

	MetricManager struct {
		Serv            *http.Server
		Store           Repository
		Address         string `env:"ADDRESS" envDefault:"none"`
		StoreInterval   int    `env:"STORE_INTERVAL" envDefault:"-1"`
		Restore         bool   `env:"RESTORE" envDefault:"true"`
		FileStoragePath string
		DBAddress       string `env:"DATABASE_DSN" envDefault:"none"`
	}
)

// инициализация хранилища и запуск
func (mm *MetricManager) Run() error {
	switch store := mm.Store.(type) {
	case *DataBase:
		log.Info("DB storage")
		config, err := pgxpool.ParseConfig(mm.DBAddress)
		if err != nil {
			return fmt.Errorf("unable to parse connection string: %w", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		if store.Pool, err = pgxpool.NewWithConfig(ctx, config); err != nil {
			return fmt.Errorf("unable to create connection pool: %w", err)
		}

		if err = store.createTables(ctx); err != nil {
			return err
		}

		if err = store.prepareQueries(ctx); err != nil {
			return err
		}

	case *FileStorage:
		log.Info("File storage")
		if mm.Restore {
			store.restoreFromFile()
		}
		if mm.StoreInterval > 0 {
			store.dumpWithInterval()
		}
	}

	log.Info("Metric Manger configuration",
		zap.String("addr", mm.Address),
		zap.Int("store interval", mm.StoreInterval),
		zap.Bool("restore", mm.Restore),
		zap.String("file store path", mm.FileStoragePath),
		zap.String("data base config", mm.DBAddress))

	return mm.Serv.ListenAndServe()
}

func (mm *MetricManager) UpdateHandler(rw http.ResponseWriter, req *http.Request) {
	mtype := chi.URLParam(req, m.Mtype)
	name := chi.URLParam(req, m.ID)
	value := chi.URLParam(req, m.Value)

	metric, err := m.NewMetric(name, mtype, value)
	if err != nil {
		log.Warn("UpdateHandler(): Invalid metric value or type")
		http.Error(rw, m.BadRequestMessage, http.StatusBadRequest)
		return
	}
	bytes, err := metric.MarshalJSON()
	if err != nil {
		log.Warn("UpdateHandler(): marshal error", zap.Error(err))
		http.Error(rw, m.InternalErrorMsg, http.StatusInternalServerError)
		return
	}

	if _, err = mm.Store.Put(name, bytes, getHelper(mtype)); err != nil {
		log.Warn("UpdateHandler(): storage error", zap.Error(err))
		http.Error(rw, m.InternalErrorMsg, http.StatusInternalServerError)
		return
	}

	rw.WriteHeader(http.StatusOK)
}

func (mm *MetricManager) GetHandler(rw http.ResponseWriter, req *http.Request) {
	mtype := chi.URLParam(req, m.Mtype)
	name := chi.URLParam(req, m.ID)

	newBytes, err := mm.Store.Get(name)
	if err != nil {
		log.Warn("GetHandler(): Coundn't fetch the metric from store")
		http.Error(rw, m.NotFoundMessage, http.StatusNotFound)
		return
	}
	// получение значений полей Delta или Value
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
	metrics, _ := mm.Store.List()
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

	if _, err = mm.Store.Put(name, bytes, getHelper(mtype)); err != nil {
		log.Warn("UpdateJSON(): couldn't write to store", zap.Error(err))
		http.Error(rw, m.InternalErrorMsg, http.StatusInternalServerError)
		return
	}

	newBytes := bytes
	if mtype == m.Counter {
		name := gjson.GetBytes(bytes, m.Delta).String()
		newBytes, _ = mm.Store.Get(name)
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

	log.Info("bytes", zap.String("bytes", string(bytes)))

	name := gjson.GetBytes(bytes, m.ID).String()
	log.Info("needed metric name", zap.String("name", name))
	if bytes, err = mm.Store.Get(name); err != nil {
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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.Ping(ctx); err != nil {
		log.Warn("PingHandler(): There is no connection to data base")
		http.Error(rw, m.InternalErrorMsg, http.StatusInternalServerError)
		return
	}
	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte("The connection is established!"))
}

func (mm *MetricManager) UpdatesJSON(rw http.ResponseWriter, req *http.Request) {
	b, err := io.ReadAll(req.Body)
	if err != nil {
		log.Warn("UpdatesJSON(): Couldn't read request body")
		http.Error(rw, m.BadRequestMessage, http.StatusBadRequest)
		return
	}
	defer req.Body.Close()

	if db, ok := mm.Store.(*DataBase); ok {
		if err = db.insertBatch(context.TODO(), b); err != nil {
			log.Warn("UpdatesJSON(): couldn't send a batch", zap.Error(err))
			http.Error(rw, m.InternalErrorMsg, http.StatusBadRequest)
		}
	} else {
		gjson.ParseBytes(b).ForEach(func(key, value gjson.Result) bool {
			name := value.Get(m.ID).String()
			mtype := value.Get(m.Mtype).String()
			_, _ = mm.Store.Put(name, []byte(value.Raw), getHelper(mtype))
			return true
		})
	}
	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte("Batch is accepted"))
}
