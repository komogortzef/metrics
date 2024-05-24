package server

import (
	"net/http"
	"regexp"

	"metrics/internal/logger"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

const (
	internalErrorMsg  = "internal server error"
	notFoundMessage   = "not found"
	badRequestMessage = "bad request"
	gauge             = "gauge"
	counter           = "counter"
)

type Operation func([]byte, []byte) ([]byte, error)

type Repository interface {
	Save(key string, value []byte, opers ...Operation) error
	Get(key string) ([]byte, bool)
	GetAll() map[string][]byte
}

var storage Repository

func UpdateHandler(rw http.ResponseWriter, req *http.Request) {
	logger.Debug("UPDATE HANDLER starts ...")

	kind := chi.URLParam(req, "kind")
	name := chi.URLParam(req, "name")
	val := chi.URLParam(req, "val")

	switch kind {
	case gauge:
		isReal := regexp.MustCompile(`^-?\d*\.?\d+$`).MatchString
		if !isReal(val) {
			logger.Warn("Invalid value for gauge", zap.String("val", val))
			http.Error(rw, badRequestMessage, http.StatusBadRequest)
			return
		}
		if err := storage.Save(name, []byte(val)); err != nil {
			http.Error(rw, internalErrorMsg, http.StatusInternalServerError)
			return
		}
		rw.WriteHeader(http.StatusOK)
	case counter:
		isNatural := regexp.MustCompile(`^\d+$`).MatchString
		if !isNatural(val) {
			logger.Warn("Invalid value for counter")
			http.Error(rw, badRequestMessage, http.StatusBadRequest)
			return
		}
		if err := storage.Save(name, []byte(val), withAccInt64); err != nil {
			http.Error(rw, badRequestMessage, http.StatusBadRequest)
			return
		}
		rw.WriteHeader(http.StatusOK)
	default:
		logger.Warn("Invalid metric type")
		http.Error(rw, badRequestMessage, http.StatusBadRequest)
	}
}

func GetHandler(rw http.ResponseWriter, req *http.Request) {
	logger.Debug("GET HANDLER starts ...")
	name := chi.URLParam(req, "name")

	switch chi.URLParam(req, "kind") {
	case gauge:
		data, ok := storage.Get(name)
		if !ok {
			logger.Warn("There is no such metric")
			http.Error(rw, notFoundMessage, http.StatusNotFound)
			return
		}
		if bytes, err := rw.Write(data); err != nil {
			logger.Warn("failed to send data. size", zap.Int("size", bytes))
		}
		rw.WriteHeader(http.StatusOK)
	case counter:
		data, ok := storage.Get(name)
		if !ok {
			logger.Warn("There is no such metric")
			http.Error(rw, notFoundMessage, http.StatusNotFound)
			return
		}
		rw.WriteHeader(http.StatusOK)
		_, _ = rw.Write(data)
	default:
		logger.Warn("Invalid metric type")
		http.Error(rw, notFoundMessage, http.StatusNotFound)
	}
}

func GetAllHandler(rw http.ResponseWriter, req *http.Request) {
	logger.Debug("GET ALL HANDLER starts ...")
	list := make([]Item, 0, metricsNumber)

	logger.Info("Collect all metrics...")
	for name, value := range storage.GetAll() {
		list = append(list, Item{Name: name, Value: string(value)})
	}

	logger.Info("Creating an html page...")
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
