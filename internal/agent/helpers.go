package agent

import (
	"fmt"

	"metrics/internal/logger"
	"metrics/internal/models"

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
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

func send(kind string, name string, val any, client *resty.Client) {
	if sendFormat == "json" {
		var metric models.Metrics
		metric.MType = kind
		metric.ID = name
		switch val := val.(type) {
		case int64:
			metric.Delta = &val
		case float64:
			metric.Value = &val
		}
		resp, err := client.R().
			SetHeader("Content-Type", "application/json").
			SetBody(metric).
			Post("update/")
		if err != nil {
			logger.Warn("No connection")
			successSend = false
		}
		logger.Debug("RESPONSE:", zap.String("resp", string(resp.Body())))
	} else {
		var URL string
		if kind == gauge {
			URL = fmt.Sprintf("update/%s/%s/%f",
				kind, name, val)
		} else {
			URL = fmt.Sprintf("update/%s/%s/%d", kind, name, val)
		}
		if _, err := client.R().Post(URL); err != nil {
			logger.Warn("No connection")
			successSend = false
		}
	}
}
