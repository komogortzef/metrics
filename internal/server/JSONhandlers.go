package server

import (
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
		val := strconv.FormatFloat(*metricData.Value, 'g', -1, 64)
		_ = storage.Save(metricData.ID, []byte(val))
	case counter:
		val := strconv.FormatInt(*metricData.Delta, 10)
		_ = storage.Save(metricData.ID, []byte(val), withAccInt64)
		bytes, _ := storage.Get(metricData.ID)
		floatVal, _ := strconv.ParseFloat(string(bytes), 64)
		metricData.Value = &floatVal
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
	rw.Write(jsonBytes)
}

func GetJSON(rw http.ResponseWriter, req *http.Request) {

}
