package agent

import (
	"bytes"
	ctx "context"
	"fmt"
	"net/http"
	"runtime"

	"metrics/internal/compress"
	"metrics/internal/security"
	s "metrics/internal/service"

	"github.com/pquerna/ffjson/ffjson"
)

var memStats = runtime.MemStats{}

func (sm *SelfMonitor) collectMetrics() {
	sm.metrics = []s.Metrics{
		s.BuildMetric("Alloc", float64(memStats.Alloc)),
		s.BuildMetric("BuckHashSys", float64(memStats.BuckHashSys)),
		s.BuildMetric("Frees", float64(memStats.Frees)),
		s.BuildMetric("GCCPUFraction", memStats.GCCPUFraction),
		s.BuildMetric("GCSys", float64(memStats.GCSys)),
		s.BuildMetric("HeapAlloc", float64(memStats.HeapAlloc)),
		s.BuildMetric("HeapIdle", float64(memStats.HeapIdle)),
		s.BuildMetric("HeapInuse", float64(memStats.HeapInuse)),
		s.BuildMetric("HeapObjects", float64(memStats.HeapObjects)),
		s.BuildMetric("HeapReleased", float64(memStats.HeapReleased)),
		s.BuildMetric("HeapSys", float64(memStats.HeapSys)),
		s.BuildMetric("LastGC", float64(memStats.LastGC)),
		s.BuildMetric("Lookups", float64(memStats.Lookups)),
		s.BuildMetric("MCacheInuse", float64(memStats.MCacheInuse)),
		s.BuildMetric("MCacheSys", float64(memStats.MCacheSys)),
		s.BuildMetric("MSpanInuse", float64(memStats.MSpanInuse)),
		s.BuildMetric("MSpanSys", float64(memStats.MSpanSys)),
		s.BuildMetric("Mallocs", float64(memStats.Mallocs)),
		s.BuildMetric("NextGC", float64(memStats.NextGC)),
		s.BuildMetric("NumForcedGC", float64(memStats.NumForcedGC)),
		s.BuildMetric("NumGC", float64(memStats.NumGC)),
		s.BuildMetric("OtherSys", float64(memStats.OtherSys)),
		s.BuildMetric("PauseTotalNs", float64(memStats.PauseTotalNs)),
		s.BuildMetric("StackInuse", float64(memStats.StackInuse)),
		s.BuildMetric("StackSys", float64(memStats.StackSys)),
		s.BuildMetric("Sys", float64(memStats.Sys)),
		s.BuildMetric("TotalAlloc", float64(memStats.TotalAlloc)),
		s.BuildMetric("RandomValue", sm.randVal),
		s.BuildMetric("PollCount", sm.pollCount),
	}
}

func (sm *SelfMonitor) sendBatch(cx ctx.Context) error {
	url := "http://" + sm.Address + "/updates/"
	data, err := ffjson.Marshal(sm.metrics)
	if err != nil {
		return fmt.Errorf("sendBatch marshal err: %w", err)
	}
	compressData, err := compress.Compress(data)
	if err != nil {
		return fmt.Errorf("sendBatch compress err: %w", err)
	}

	req, _ := http.NewRequest(http.MethodPost, url, bytes.NewReader(compressData))
	if sm.Key != "" {
		sign := security.Hash(&data, sm.Key)
		req.Header.Set("HashSHA256", sign)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")

	r, err := http.DefaultClient.Do(req)
	if err != nil {
		if err = s.Retry(cx, func() error {
			r2, err := http.DefaultClient.Do(req)
			closeBody(r2)
			return err
		}); err != nil {
			closeBody(r)
			return fmt.Errorf("sendBatch client error: %w", err)
		}
	}
	closeBody(r)
	return nil
}

func closeBody(r *http.Response) {
	if r != nil && r.Body != nil {
		r.Body.Close()
	}
}
