package server

import (
	"net/http"

	"metrics/internal/logger"
	"metrics/internal/models"

	"github.com/go-chi/chi/v5"
)

const (
	internalErrorMsg  = "internal server error"
	notFoundMessage   = "not found"
	badRequestMessage = "bad request"
	gauge             = "gauge"
	counter           = "counter"
)

type Repository interface {
	Update(key string, value []byte) error
	Get(key string) ([]byte, bool)
}

var storage Repository

func UpdateHandler(rw http.ResponseWriter, req *http.Request) {
	logger.Debug("UPDATE HANDLER starts ...")

	kind := chi.URLParam(req, "kind")
	name := chi.URLParam(req, "name")
	val := chi.URLParam(req, "val")

	bytes, err := toBytes(kind, val)
	if err != nil {
		logger.Warn("Invalid value or kind")
		http.Error(rw, badRequestMessage, http.StatusBadRequest)
		return
	}

	if err = storage.Update(name, bytes); err != nil {
		logger.Warn("Internal server error")
		http.Error(rw, internalErrorMsg, http.StatusInternalServerError)
	}
	models.Accounter.Put(kind, name)

	rw.WriteHeader(http.StatusOK)
}

func GetHandler(rw http.ResponseWriter, req *http.Request) {
	logger.Debug("GET HANDLER starts ...")
	if kind := chi.URLParam(req, "kind"); !(kind == gauge || kind == counter) {
		logger.Warn("Invalid metric type")
		http.Error(rw, notFoundMessage, http.StatusNotFound)
	}
	name := chi.URLParam(req, "name")
	data, ok := storage.Get(name)
	if !ok {
		logger.Warn("there is no such metric")
		http.Error(rw, notFoundMessage, http.StatusNotFound)
		return
	}

	numStr := bytesToString(name, data)

	rw.WriteHeader(http.StatusOK)
	_, _ = rw.Write([]byte(numStr))
}

func GetAllHandler(rw http.ResponseWriter, req *http.Request) {
	logger.Debug("GET ALL HANDLER starts ...")
	list := make([]Item, 0, models.MetricsNumber)

	for _, key := range models.Accounter.List() {
		bytes, ok := storage.Get(key)
		if !ok {
			continue
		}
		list = append(list, Item{Name: key, Value: bytesToString(key, bytes)})
	}

	html, err := renderGetAll(list)
	if err != nil {
		logger.Warn("An error occured during html rendering")
		http.Error(rw, internalErrorMsg, http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "text/html")
	rw.WriteHeader(http.StatusOK)
	_, _ = rw.Write(html.Bytes())
}
