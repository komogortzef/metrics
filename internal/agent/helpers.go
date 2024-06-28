package agent

import (
	"bytes"
	ctx "context"
	"fmt"
	"net/http"
	"runtime"

	"metrics/internal/compress"
	"metrics/internal/logger"
	"metrics/internal/security"
	s "metrics/internal/service"

	"github.com/shirou/gopsutil/v4/mem"
	"go.uber.org/zap"
)

const numMemMetrics = 31

var (
	memStats      = runtime.MemStats{}
	psMem         *mem.VirtualMemoryStat
	psCPUs        []float64
	numCPUs       = runtime.NumCPU()
	numAllMetrics = numCPUs + numMemMetrics
)

func NewSelfMonitor() *SelfMonitor {
	return &SelfMonitor{
		mets: make([]*s.Metrics, numAllMetrics),
	}
}

func (sm *SelfMonitor) collectRuntime() {
	sm.mets[0] = s.BuildMetric("Alloc", float64(memStats.Alloc))
	sm.mets[1] = s.BuildMetric("BuckHashSys", float64(memStats.BuckHashSys))
	sm.mets[2] = s.BuildMetric("Frees", float64(memStats.Frees))
	sm.mets[3] = s.BuildMetric("GCCPUFraction", memStats.GCCPUFraction)
	sm.mets[4] = s.BuildMetric("GCSys", float64(memStats.GCSys))
	sm.mets[5] = s.BuildMetric("HeapAlloc", float64(memStats.HeapAlloc))
	sm.mets[6] = s.BuildMetric("HeapIdle", float64(memStats.HeapIdle))
	sm.mets[7] = s.BuildMetric("HeapInuse", float64(memStats.HeapInuse))
	sm.mets[8] = s.BuildMetric("HeapObjects", float64(memStats.HeapObjects))
	sm.mets[9] = s.BuildMetric("HeapReleased", float64(memStats.HeapReleased))
	sm.mets[10] = s.BuildMetric("HeapSys", float64(memStats.HeapSys))
	sm.mets[11] = s.BuildMetric("LastGC", float64(memStats.LastGC))
	sm.mets[12] = s.BuildMetric("Lookups", float64(memStats.Lookups))
	sm.mets[13] = s.BuildMetric("MCacheInuse", float64(memStats.MCacheInuse))
	sm.mets[14] = s.BuildMetric("MCacheSys", float64(memStats.MCacheSys))
	sm.mets[15] = s.BuildMetric("MSpanInuse", float64(memStats.MSpanInuse))
	sm.mets[16] = s.BuildMetric("MSpanSys", float64(memStats.MSpanSys))
	sm.mets[17] = s.BuildMetric("Mallocs", float64(memStats.Mallocs))
	sm.mets[18] = s.BuildMetric("NextGC", float64(memStats.NextGC))
	sm.mets[19] = s.BuildMetric("NumForcedGC", float64(memStats.NumForcedGC))
	sm.mets[20] = s.BuildMetric("NumGC", float64(memStats.NumGC))
	sm.mets[21] = s.BuildMetric("OtherSys", float64(memStats.OtherSys))
	sm.mets[22] = s.BuildMetric("PauseTotalNs", float64(memStats.PauseTotalNs))
	sm.mets[23] = s.BuildMetric("StackInuse", float64(memStats.StackInuse))
	sm.mets[24] = s.BuildMetric("StackSys", float64(memStats.StackSys))
	sm.mets[25] = s.BuildMetric("Sys", float64(memStats.Sys))
	sm.mets[26] = s.BuildMetric("TotalAlloc", float64(memStats.TotalAlloc))
	sm.mets[27] = s.BuildMetric("RandomValue", sm.randVal)
	sm.mets[28] = s.BuildMetric("PollCount", sm.pollCount)
}

func (sm *SelfMonitor) collectPs() {
	sm.mets[29] = s.BuildMetric("TotalMemory", float64(psMem.Total))
	sm.mets[30] = s.BuildMetric("FreeMemory", float64(psMem.Free))

	for i, v := range psCPUs {
		idx := i + numMemMetrics
		name := fmt.Sprintf("CPUutilization%d", i+1)
		sm.mets[idx] = s.BuildMetric(name, v)
	}
}

func (sm *SelfMonitor) sendWorker(cx ctx.Context, id int, ch chan []byte, errCh chan error) {
	url := "http://" + sm.Address + "/updates/"
	for {
		select {
		case <-cx.Done():
			close(errCh)
			logger.Debug("goodbye from sendWorker")
			return
		case data := <-ch:
			logger.Info("worker works...",
				zap.Int("woker", id))
			compressData, _ := compress.Compress(data)
			req, _ := http.NewRequest(http.MethodPost, url, bytes.NewReader(compressData))
			if sm.Key != "" {
				sign := security.Hash(&data, sm.Key)
				req.Header.Set("HashSHA256", sign)
			}
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Content-Encoding", "gzip")

			r, err := http.DefaultClient.Do(req)
			if err != nil {
				errCh <- err
				err = s.Retry(cx, func() error {
					r2, err2 := http.DefaultClient.Do(req)
					closeBody(r2)
					return err2
				})
			}
			closeBody(r)
			errCh <- err
		}
	}
}

func closeBody(r *http.Response) {
	if r != nil && r.Body != nil {
		r.Body.Close()
	}
}
