package server

import (
	"net/http"
	"regexp"
	"sync"

	"metrics/internal/logger"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

const (
	internalServerErrorMessage = "internal server error"
	notFoundMessage            = "not found"
	badRequestMessage          = "bad request"
)

// для действий со значениями при сохранении в хранилище
type Operation func([]byte, []byte) ([]byte, error)

type Repository interface {
	Save(key string, value []byte, opers ...Operation) error
	Get(key string) ([]byte, bool)
	GetAll() map[string][]byte
}

var storage Repository

// в будущем пригодится для конфигурирования хранилища
func SetStorage(st string) {
	logger.Info("Set storage ...")
	switch st {
	default:
		storage = &MemStorage{
			Items: make(map[string][]byte, metricsNumber),
			Mtx:   &sync.RWMutex{},
		}
	}
}

func UpdateHandler(rw http.ResponseWriter, req *http.Request) {
	logger.Info("UPDATE HANDLER starts ...")
	name := chi.URLParam(req, "name")
	val := chi.URLParam(req, "val")

	switch chi.URLParam(req, "kind") {
	case "gauge":
		isReal := regexp.MustCompile(`^-?\d*\.?\d+$`).MatchString
		if !isReal(val) {
			logger.Warn("Invalid value for gauge metric")
			http.Error(rw, badRequestMessage, http.StatusBadRequest)
			return
		}
		if err := storage.Save(name, []byte(val)); err != nil {
			http.Error(rw, internalServerErrorMessage, http.StatusInternalServerError)
			return
		}
		rw.WriteHeader(http.StatusOK)

	case "counter":
		isNatural := regexp.MustCompile(`^\d+$`).MatchString
		if !isNatural(val) {
			logger.Warn("Invalid value for counter metric")
			http.Error(rw, badRequestMessage, http.StatusBadRequest)
			return
		}
		if err := storage.Save(name, []byte(val), WithAccInt64); err != nil {
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
	logger.Info("GET HANDLER starts ...")
	name := chi.URLParam(req, "name")

	switch chi.URLParam(req, "kind") {
	case "counter":
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
	case "gauge":
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
	default:
		logger.Warn("Invalid metric type")
		http.Error(rw, notFoundMessage, http.StatusNotFound)
	}
}

func GetAllHandler(rw http.ResponseWriter, req *http.Request) {
	logger.Info("GET ALL HANDLER starts ...")
	list := make([]Item, 0, metricsNumber)

	logger.Info("Collect all metrics...")
	for name, value := range storage.GetAll() {
		list = append(list, Item{Name: name, Value: string(value)})
	}

	logger.Info("Creating an html page...")
	html, err := renderGetAll(list)
	if err != nil {
		http.Error(rw, internalServerErrorMessage, http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "text/html")
	if _, err := html.WriteTo(rw); err != nil {
		logger.Warn("The html page could not be sent")
		http.Error(rw, internalServerErrorMessage, http.StatusInternalServerError)
	}
	rw.WriteHeader(http.StatusOK)
}
