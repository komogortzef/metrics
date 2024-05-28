package server

import (
	"encoding/binary"
	"io"
	"math"
	"net/http"

	"metrics/internal/logger"
	"metrics/internal/models"

	"go.uber.org/zap"
)

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

	bytes, err = toBytes(metricData.MType, metricData)
	if err != nil {
		http.Error(rw, badRequestMessage, http.StatusBadRequest)
		return
	}

	if metricData.MType == counter {
		err = storage.Update(metricData.ID, bytes)
		bytes, _ := storage.Get(metricData.ID)
		intVal := int64(binary.LittleEndian.Uint64(bytes))
		metricData.Delta = &intVal
	} else {
		err = storage.Update(metricData.ID, bytes)
	}
	if err != nil {
		logger.Warn("saving error")
		http.Error(rw, internalErrorMsg, http.StatusInternalServerError)
		return
	}
	models.Accounter.Put(metricData.MType, metricData.ID)

	jsonBytes, err := metricData.MarshalJSON()
	if err != nil {
		logger.Warn("Couldn't serealize", zap.String("error", err.Error()))
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	_, _ = rw.Write(jsonBytes)
}

func GetJSON(rw http.ResponseWriter, req *http.Request) {
	logger.Info("GetJSON starts...")
	var metricData models.Metrics

	var bytes []byte
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

	val, ok := storage.Get(metricData.ID)
	if !ok {
		http.Error(rw, notFoundMessage, http.StatusNotFound)
		return
	}

	if metricData.MType == gauge {
		num := math.Float64frombits(binary.LittleEndian.Uint64(val))
		metricData.Value = &num
	} else {
		num := int64(binary.LittleEndian.Uint64(val))
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
