package server

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	"metrics/internal/logger"
	"metrics/internal/models"

	"go.uber.org/zap"
)

func UpdateJSON(rw http.ResponseWriter, req *http.Request) {
	logger.Info("UpdateJSON starts...")

	var metricData models.Metrics
	json, _ := io.ReadAll(req.Body)
	logger.Info("Unmarshal JSON...")
	if err := metricData.UnmarshalJSON(json); err != nil {
		http.Error(rw, badRequestMessage, http.StatusBadRequest)
		return
	}
	defer req.Body.Close()

	logger.Info("Saving in memory...")
	switch metricData.MType {
	case gauge:
		val := fmt.Sprintf("%f", *metricData.Value)
		_ = storage.Save(metricData.ID, []byte(val))
	case counter:
		val := strconv.FormatInt(*metricData.Delta, 10)
		_ = storage.Save(metricData.ID, []byte(val), withAccInt64)
		bytes, _ := storage.Get(metricData.ID)
		intVal, _ := strconv.ParseInt(string(bytes), 10, 64)
		metricData.Delta = &intVal
	default:
		http.Error(rw, badRequestMessage, http.StatusBadRequest)
	}

	logger.Info("Marshal JSON...")
	jsonBytes, err := metricData.MarshalJSON()
	if err != nil {
		logger.Warn("Couldn't serealize", zap.String("error", err.Error()))
	}

	logger.Info("Sending response...")
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	_, _ = rw.Write(jsonBytes)
}

func GetJSON(rw http.ResponseWriter, req *http.Request) {
	logger.Info("GetJSON starts...")
	var metricData models.Metrics

	logger.Info("Unmarshal JSON...")
	jsonBytes, _ := io.ReadAll(req.Body)
	if err := metricData.UnmarshalJSON(jsonBytes); err != nil {
		http.Error(rw, badRequestMessage, http.StatusBadRequest)
		return
	}

	logger.Info("fetching starts...")
	val, ok := storage.Get(metricData.ID)
	if !ok {
		http.Error(rw, notFoundMessage, http.StatusNotFound)
		return
	}

	if metricData.MType == gauge {
		num, _ := strconv.ParseFloat(string(val), 64)
		metricData.Value = &num
	} else {
		num, _ := strconv.ParseInt(string(val), 10, 64)
		metricData.Delta = &num
	}

	logger.Info("Marshal JSON...")
	jsonBytes, err := metricData.MarshalJSON()
	if err != nil {
		logger.Warn("Couldn't serealize", zap.String("error", err.Error()))
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	_, _ = rw.Write(jsonBytes)
}
