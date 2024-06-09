package server

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	l "metrics/internal/logger"
	m "metrics/internal/models"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tidwall/gjson"
	"go.uber.org/zap"
)

type (
	helper func([]byte, []byte) ([]byte, error) // для доп операций перед сохраненим в Store

	Repository interface {
		Put(key string, data []byte, helpers ...helper) (int, error)
		Get(key string) ([]byte, bool)
		Pop(key string) ([]byte, error)
	}

	MetricsManager struct {
		Serv  *http.Server
		Store Repository
	}
)

// инициализация хранилища и запуск
func (mm *MetricsManager) Run() error {
	switch s := mm.Store.(type) {
	case *DataBase:
		config, err := pgxpool.ParseConfig(s.Addr)
		if err != nil {
			return fmt.Errorf("unable to parse connection string: %w", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		if s.Pool, err = pgxpool.NewWithConfig(ctx, config); err != nil {
			return fmt.Errorf("unable to create connection pool: %w", err)
		}
	case *FileStorage:
		if s.Restore {
			s.restoreFromFile()
		}
		s.startTicker()
	}

	return mm.Serv.ListenAndServe()
}

func (mm *MetricsManager) UpdateHandler(rw http.ResponseWriter, req *http.Request) {
	mtype := chi.URLParam(req, m.Mtype)
	name := chi.URLParam(req, m.ID)
	value := chi.URLParam(req, m.Value)

	metric, err := m.NewMetric(name, mtype, value)
	if err != nil {
		l.Warn("UpdateHandler(): Invalid metric value or type")
		http.Error(rw, m.BadRequestMessage, http.StatusBadRequest)
		return
	}
	bytes, err := metric.MarshalJSON()
	if err != nil {
		l.Warn("UpdateHandler(): marshal error", zap.Error(err))
		http.Error(rw, m.InternalErrorMsg, http.StatusInternalServerError)
		return
	}

	if mtype == m.Counter {
		_, err = mm.Store.Put(name, bytes, addCounter)
	} else {
		_, err = mm.Store.Put(name, bytes)
	}
	if err != nil {
		l.Warn("UpdateHandler(): storage error", zap.Error(err))
		http.Error(rw, m.InternalErrorMsg, http.StatusInternalServerError)
		return
	}

	rw.WriteHeader(http.StatusOK)
}

func (mm *MetricsManager) GetHandler(rw http.ResponseWriter, req *http.Request) {
	mtype := chi.URLParam(req, m.Mtype)
	name := chi.URLParam(req, m.ID)

	newBytes, ok := mm.Store.Get(name)
	if !ok {
		l.Warn("GetHandler(): Coundn't fetch the metric from store")
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

func (mm *MetricsManager) GetAllHandler(rw http.ResponseWriter, req *http.Request) {
	list := make([]Item, 0, m.MetricsNumber)

	var metric m.Metrics
	for _, bytes := range getList(mm.Store) {
		_ = metric.UnmarshalJSON(bytes)
		list = append(list, Item{Met: metric.String()})
	}

	html, err := renderGetAll(list)
	if err != nil {
		l.Warn("GetAllHandler(): An error occured during html rendering")
		http.Error(rw, m.InternalErrorMsg, http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "text/html")
	rw.WriteHeader(http.StatusOK)
	_, _ = rw.Write(html.Bytes())
}

func (mm *MetricsManager) UpdateJSON(rw http.ResponseWriter, req *http.Request) {
	bytes, err := io.ReadAll(req.Body)
	if err != nil {
		l.Warn("Couldn't read with decompress")
	}
	defer req.Body.Close()

	name := gjson.GetBytes(bytes, m.ID).String()
	mtype := gjson.GetBytes(bytes, m.Mtype).String()

	if mtype == m.Counter {
		_, err = mm.Store.Put(name, bytes, addCounter)
	} else {
		_, err = mm.Store.Put(name, bytes)
	}
	if err != nil {
		l.Warn("UpdateJSON(): couldn't write to store", zap.Error(err))
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

func (mm *MetricsManager) GetJSON(rw http.ResponseWriter, req *http.Request) {
	bytes, err := io.ReadAll(req.Body)
	if err != nil {
		l.Warn("GetJSON(): Couldn't read request body")
		http.Error(rw, m.BadRequestMessage, http.StatusBadRequest)
		return
	}
	defer req.Body.Close()

	name := gjson.GetBytes(bytes, m.ID).String()
	bytes, ok := mm.Store.Get(name)
	if !ok {
		l.Warn("GetJSON(): No such metric in store")
		http.Error(rw, m.NotFoundMessage, http.StatusNotFound)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	_, _ = rw.Write(bytes)
}

func (mm *MetricsManager) PingHandler(rw http.ResponseWriter, req *http.Request) {
	db, ok := mm.Store.(*DataBase)
	if !ok {
		l.Warn("PingHandler(): Invalid storage type for ping")
		http.Error(rw, m.InternalErrorMsg, http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.Ping(ctx); err != nil {
		l.Warn("PingHandler(): There is no connection to data base")
		http.Error(rw, m.InternalErrorMsg, http.StatusInternalServerError)
		return
	}
	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte("The connection is established!"))
}
