package agent

import (
	"bytes"
	"fmt"
	"net/http"
	"runtime"

	c "metrics/internal/compress"
	m "metrics/internal/models"
)

var memStats = runtime.MemStats{}

// формирование метрик и сбор в массив,
// присваивание массива переменной структуры
func (sm *SelfMonitor) getMetrics() {
	sm.metrics = [m.MetricsNumber]m.Metrics{
		m.BuildMetric("Alloc", float64(memStats.Alloc)),
		m.BuildMetric("BuckHashSys", float64(memStats.BuckHashSys)),
		m.BuildMetric("Frees", float64(memStats.Frees)),
		m.BuildMetric("GCCPUFraction", memStats.GCCPUFraction),
		m.BuildMetric("GCSys", float64(memStats.GCSys)),
		m.BuildMetric("HeapAlloc", float64(memStats.HeapAlloc)),
		m.BuildMetric("HeapIdle", float64(memStats.HeapIdle)),
		m.BuildMetric("HeapInuse", float64(memStats.HeapInuse)),
		m.BuildMetric("HeapObjects", float64(memStats.HeapObjects)),
		m.BuildMetric("HeapReleased", float64(memStats.HeapReleased)),
		m.BuildMetric("HeapSys", float64(memStats.HeapSys)),
		m.BuildMetric("LastGC", float64(memStats.LastGC)),
		m.BuildMetric("Lookups", float64(memStats.Lookups)),
		m.BuildMetric("MCacheInuse", float64(memStats.MCacheInuse)),
		m.BuildMetric("MCacheSys", float64(memStats.MCacheSys)),
		m.BuildMetric("MSpanInuse", float64(memStats.MSpanInuse)),
		m.BuildMetric("MSpanSys", float64(memStats.MSpanSys)),
		m.BuildMetric("Mallocs", float64(memStats.Mallocs)),
		m.BuildMetric("NextGC", float64(memStats.NextGC)),
		m.BuildMetric("NumForcedGC", float64(memStats.NumForcedGC)),
		m.BuildMetric("NumGC", float64(memStats.NumGC)),
		m.BuildMetric("OtherSys", float64(memStats.OtherSys)),
		m.BuildMetric("PauseTotalNs", float64(memStats.PauseTotalNs)),
		m.BuildMetric("StackInuse", float64(memStats.StackInuse)),
		m.BuildMetric("StackSys", float64(memStats.StackSys)),
		m.BuildMetric("Sys", float64(memStats.Sys)),
		m.BuildMetric("TotalAlloc", float64(memStats.TotalAlloc)),
		m.BuildMetric("RandomValue", sm.randVal),
		m.BuildMetric("PollCount", sm.pollCount),
	}
}

// отправка метрики
func (sm *SelfMonitor) send(metric m.Metrics) error {
	baseurl := "http://" + sm.Address + "/update/"
	jsonBytes, err := metric.MarshalJSON()
	if err != nil {
		return fmt.Errorf("couldn't Marshall JSON: %w", err)
	}
	compJSON, err := c.Compress(jsonBytes)
	if err != nil {
		return fmt.Errorf("couldn't compress: %w", err)
	}
	req, _ := http.NewRequest(http.MethodPost, baseurl, bytes.NewReader(compJSON))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")

	r, err := http.DefaultClient.Do(req)
	if r != nil && r.Body != nil {
		r.Body.Close()
	}

	return err
}
