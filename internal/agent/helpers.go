package agent

import (
	"bytes"
	"fmt"
	"net/http"

	"metrics/internal/compress"
	"metrics/internal/logger"
	"metrics/internal/models"

	"go.uber.org/zap"
)

const (
	gauge   = "gauge"
	counter = "counter"
)

var (
	address        string = "localhost:8080"
	pollInterval   int    = 2
	reportInterval int    = 10
	successSend    bool   = true
	sendFormat     string = ""
)

func SetCond(addr, format string, poll, report int) {
	address = addr
	pollInterval = poll
	reportInterval = report
	sendFormat = format
}

func send(kind, name string, val float64) {
	baseurl := "http://" + address + "/update/"
	switch sendFormat {
	case "json":
		var metric models.Metrics
		metric.MType = kind
		metric.ID = name

		if kind == counter {
			intVal := int64(val)
			metric.Delta = &intVal
		} else {
			metric.Value = &val
		}

		jsonBytes, err := metric.MarshalJSON()
		if err != nil {
			logger.Error("Coulnd't Marshall JSON")
		}

		compJSON, err := compress.Compress(jsonBytes)
		if err != nil {
			logger.Error("compress error!!")
		}

		req, err := http.NewRequest(http.MethodPost, baseurl, bytes.NewReader(compJSON))
		if err != nil {
			logger.Warn("Create request error")
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Content-Encoding", "gzip")
		r, err := http.DefaultClient.Do(req)
		if err != nil {
			logger.Warn("No connection", zap.String("err", err.Error()))
			successSend = false
		}
		defer func() {
			if r != nil && r.Body != nil {
				r.Body.Close()
			}
		}()
	default:
		url := fmt.Sprintf("%s%s/%s/%v", baseurl, kind, name, val)
		req, err := http.NewRequest(http.MethodPost, url, nil)
		if err != nil {
			logger.Warn("Couldn't create a req")
		}
		logger.Info(url)
		r, err := http.DefaultClient.Do(req)
		if err != nil {
			logger.Warn("No connection", zap.String("err", err.Error()))
			successSend = false
		}
		defer func() {
			if r != nil && r.Body != nil {
				r.Body.Close()
			}
		}()
	}
}
