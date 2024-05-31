package server

import (
	"io"
	"strconv"

	"metrics/internal/logger"
	"metrics/internal/models"
	"net/http"

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

func UpdateHandler(rw http.ResponseWriter, req *http.Request) {
	logger.Debug("UPDATE HANDLER starts ...")
	kind := chi.URLParam(req, "kind")
	name := chi.URLParam(req, "name")
	val := chi.URLParam(req, "val")

	if err := accounter.put(kind, name, val); err != nil {
		logger.Warn("Invalid metric value or type")
		http.Error(rw, badRequestMessage, http.StatusBadRequest)
		return
	}

	rw.WriteHeader(http.StatusOK)
}

func GetHandler(rw http.ResponseWriter, req *http.Request) {
	logger.Debug("GET HANDLER starts ...")
	name := chi.URLParam(req, "name")

	numStr, ok := storage.Get(name)
	if !ok {
		logger.Warn("there is no such metric")
		http.Error(rw, notFoundMessage, http.StatusNotFound)
		return
	}

	rw.WriteHeader(http.StatusOK)
	_, _ = rw.Write([]byte(numStr))
}

func GetAllHandler(rw http.ResponseWriter, req *http.Request) {
	logger.Debug("GET ALL HANDLER starts ...")
	list := make([]Item, 0, metricsNumber)

	for _, key := range accounter.list() {
		val, ok := storage.Get(key)
		if !ok {
			continue
		}
		list = append(list, Item{Name: key, Value: val})
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

func UpdateJSON(rw http.ResponseWriter, req *http.Request) {
	logger.Info("UpdateJSON starts...")

	var metricData models.Metrics
	bytes, err := io.ReadAll(req.Body)
	if err != nil {
		logger.Warn("Couldn't read with decompress")
	}

	if err := metricData.UnmarshalJSON(bytes); err != nil {
		logger.Warn("Unmarshall JSON error")
		http.Error(rw, badRequestMessage, http.StatusBadRequest)
		return
	}
	defer req.Body.Close()

	if err = accounter.putJSON(metricData); err != nil {
		logger.Warn("Invalid metric value or type")
		http.Error(rw, badRequestMessage, http.StatusBadRequest)
		return
	}

	jsonBytes, err := metricData.MarshalJSON()
	if err != nil {
		logger.Warn("Couldn't serealize", zap.String("error", err.Error()))
	}

	rw.WriteHeader(http.StatusOK)
	_, _ = rw.Write(jsonBytes)
}

func GetJSON(rw http.ResponseWriter, req *http.Request) {
	logger.Info("GetJSON starts...")

	var metricData models.Metrics
	bytes, err := io.ReadAll(req.Body)
	if err != nil {
		logger.Warn("Couldn't read with decompress")
	}
	defer req.Body.Close()

	if err := metricData.UnmarshalJSON(bytes); err != nil {
		logger.Warn("Unmarshall JSON error")
		http.Error(rw, badRequestMessage, http.StatusBadRequest)
		return
	}

	val, ok := storage.Get(metricData.ID)
	if !ok {
		http.Error(rw, notFoundMessage, http.StatusNotFound)
		return
	}

	if metricData.MType == gauge {
		num, _ := strconv.ParseFloat(val, 64)
		metricData.Value = &num
	} else {
		num, _ := strconv.ParseInt(val, 10, 64)
		metricData.Delta = &num
	}

	jsonBytes, err := metricData.MarshalJSON()
	if err != nil {
		logger.Warn("Couldn't serealize", zap.String("error", err.Error()))
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	_, _ = rw.Write(jsonBytes)
}
