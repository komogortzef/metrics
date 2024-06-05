package agent

import (
	"bytes"
	"fmt"
	"net/http"
	"runtime"

	c "metrics/internal/compress"
	l "metrics/internal/logger"
	m "metrics/internal/models"

	"go.uber.org/zap"
)

var memStats = runtime.MemStats{}

func (sm *SelfMonitor) getMetrics() {
	sm.metrics = [m.MetricsNumber]m.Metrics{
		buildMetric("Alloc", float64(memStats.Alloc)),
		buildMetric("BuckHashSys", float64(memStats.BuckHashSys)),
		buildMetric("Frees", float64(memStats.Frees)),
		buildMetric("GCCPUFraction", memStats.GCCPUFraction),
		buildMetric("GCSys", float64(memStats.GCSys)),
		buildMetric("HeapAlloc", float64(memStats.HeapAlloc)),
		buildMetric("HeapIdle", float64(memStats.HeapIdle)),
		buildMetric("HeapInuse", float64(memStats.HeapInuse)),
		buildMetric("HeapObjects", float64(memStats.HeapObjects)),
		buildMetric("HeapReleased", float64(memStats.HeapReleased)),
		buildMetric("HeapSys", float64(memStats.HeapSys)),
		buildMetric("LastGC", float64(memStats.LastGC)),
		buildMetric("Lookups", float64(memStats.Lookups)),
		buildMetric("MCacheInuse", float64(memStats.MCacheInuse)),
		buildMetric("MCacheSys", float64(memStats.MCacheSys)),
		buildMetric("MSpanInuse", float64(memStats.MSpanInuse)),
		buildMetric("MSpanSys", float64(memStats.MSpanSys)),
		buildMetric("Mallocs", float64(memStats.Mallocs)),
		buildMetric("NextGC", float64(memStats.NextGC)),
		buildMetric("NumForcedGC", float64(memStats.NumForcedGC)),
		buildMetric("NumGC", float64(memStats.NumGC)),
		buildMetric("OtherSys", float64(memStats.OtherSys)),
		buildMetric("PauseTotalNs", float64(memStats.PauseTotalNs)),
		buildMetric("StackInuse", float64(memStats.StackInuse)),
		buildMetric("StackSys", float64(memStats.StackSys)),
		buildMetric("Sys", float64(memStats.Sys)),
		buildMetric("TotalAlloc", float64(memStats.TotalAlloc)),
		buildMetric("RandomValue", sm.randVal),
		buildMetric("PollCount", sm.pollCount),
	}
}

func (sm *SelfMonitor) send(metric m.Metrics) error {
	baseurl := "http://" + sm.Address + "/update/"
	switch sm.SendFormat {
	case "json":
		jsonBytes, err := metric.MarshalJSON()
		if err != nil {
			return fmt.Errorf("coulnd't Marshall JSON: %w", err)
		}
		compJSON, err := c.Compress(jsonBytes)
		if err != nil {
			return fmt.Errorf("couldn't compress: %w", err)
		}
		req, err := http.NewRequest(http.MethodPost, baseurl, bytes.NewReader(compJSON))
		if err != nil {
			l.Warn("Create request error", zap.Error(err))
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Content-Encoding", "gzip")
		return sendMetric(req)

	default:
		url := fmt.Sprintf("%s%s/%s/%v", baseurl, metric.MType, metric.ID, metric.Data())
		req, err := http.NewRequest(http.MethodPost, url, nil)
		if err != nil {
			l.Warn("Couldn't create a req", zap.Error(err))
		}
		return sendMetric(req)
	}
}

func sendMetric(req *http.Request) error {
	r, err := http.DefaultClient.Do(req)
	if err != nil {
		l.Warn("No connection", zap.String("err", err.Error()))
	}
	if r != nil && r.Body != nil {
		r.Body.Close()
	}
	return err
}

func buildMetric(name string, val any) m.Metrics {
	var metric m.Metrics
	metric.ID = name
	switch v := val.(type) {
	case float64:
		metric.MType = m.Gauge
		metric.Value = &v
	case int64:
		metric.MType = m.Counter
		metric.Delta = &v
	}

	return metric
}
