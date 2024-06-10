package server

import (
	"io"
	"net/http"
	"strconv"

	l "metrics/internal/logger"
	m "metrics/internal/models"

	"github.com/go-chi/chi/v5"
	"github.com/tidwall/gjson"
	"go.uber.org/zap"
)

type (
	helper func([]byte, []byte) ([]byte, error) // для доп операций перед сохраненим в Store

	Repository interface {
		Put(key string, data []byte, help helper) (int, error)
		Get(key string) ([]byte, bool)
		List() [][]byte
	}

	MetricsManager struct {
		Serv  *http.Server
		Store Repository
	}
)

// инициализация хранилища и запуск
func (mm *MetricsManager) Run() error {
	if store, ok := mm.Store.(*FileStorage); ok {
		if store.Restore {
			store.restoreFromFile()
		}
		if store.Interval > 0 {
			store.startTicker()
		}
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

	_, err = mm.Store.Put(name, bytes, getHelper(mtype))
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
	for _, bytes := range mm.Store.List() {
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
		l.Warn("Couldn't read request body")
	}
	defer req.Body.Close()

	name := gjson.GetBytes(bytes, m.ID).String()
	mtype := gjson.GetBytes(bytes, m.Mtype).String()

	_, err = mm.Store.Put(name, bytes, getHelper(mtype))
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
