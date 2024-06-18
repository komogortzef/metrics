package agent

import (
	"bytes"
	"net/http"
	"runtime"

	c "metrics/internal/compress"
	"metrics/internal/service"

	"github.com/pquerna/ffjson/ffjson"
)

var memStats = runtime.MemStats{}

func (sm *SelfMonitor) collectMetrics() {
	sm.metrics = []service.Metrics{
		service.BuildMetric("Alloc", float64(memStats.Alloc)),
		service.BuildMetric("BuckHashSys", float64(memStats.BuckHashSys)),
		service.BuildMetric("Frees", float64(memStats.Frees)),
		service.BuildMetric("GCCPUFraction", memStats.GCCPUFraction),
		service.BuildMetric("GCSys", float64(memStats.GCSys)),
		service.BuildMetric("HeapAlloc", float64(memStats.HeapAlloc)),
		service.BuildMetric("HeapIdle", float64(memStats.HeapIdle)),
		service.BuildMetric("HeapInuse", float64(memStats.HeapInuse)),
		service.BuildMetric("HeapObjects", float64(memStats.HeapObjects)),
		service.BuildMetric("HeapReleased", float64(memStats.HeapReleased)),
		service.BuildMetric("HeapSys", float64(memStats.HeapSys)),
		service.BuildMetric("LastGC", float64(memStats.LastGC)),
		service.BuildMetric("Lookups", float64(memStats.Lookups)),
		service.BuildMetric("MCacheInuse", float64(memStats.MCacheInuse)),
		service.BuildMetric("MCacheSys", float64(memStats.MCacheSys)),
		service.BuildMetric("MSpanInuse", float64(memStats.MSpanInuse)),
		service.BuildMetric("MSpanSys", float64(memStats.MSpanSys)),
		service.BuildMetric("Mallocs", float64(memStats.Mallocs)),
		service.BuildMetric("NextGC", float64(memStats.NextGC)),
		service.BuildMetric("NumForcedGC", float64(memStats.NumForcedGC)),
		service.BuildMetric("NumGC", float64(memStats.NumGC)),
		service.BuildMetric("OtherSys", float64(memStats.OtherSys)),
		service.BuildMetric("PauseTotalNs", float64(memStats.PauseTotalNs)),
		service.BuildMetric("StackInuse", float64(memStats.StackInuse)),
		service.BuildMetric("StackSys", float64(memStats.StackSys)),
		service.BuildMetric("Sys", float64(memStats.Sys)),
		service.BuildMetric("TotalAlloc", float64(memStats.TotalAlloc)),
		service.BuildMetric("RandomValue", sm.randVal),
		service.BuildMetric("PollCount", sm.pollCount),
	}
}

func (sm *SelfMonitor) sendBatch() error {
	url := "http://" + sm.Address + "/updates/"
	data, err := ffjson.Marshal(sm.metrics)
	if err != nil {
		return err
	}
	compressData, err := c.Compress(data)
	if err != nil {
		return err
	}
	req, _ := http.NewRequest(http.MethodPost, url, bytes.NewReader(compressData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
	r, err := http.DefaultClient.Do(req)
	if r != nil && r.Body != nil {
		r.Body.Close()
	}
	return err
}
