package agent

import (
	"math/rand"
	"runtime"
	"sync"
	"time"

	"metrics/internal/logger"

	"github.com/go-resty/resty/v2"
)

const gauge = "gauge"

type SelfMonitor struct {
	runtime.MemStats
	randVal   float64
	pollCount int64
	Mtx       *sync.RWMutex
}

func (m *SelfMonitor) Collect() {
	for {
		m.Mtx.Lock()
		runtime.ReadMemStats(&m.MemStats)
		m.randVal = rand.Float64()
		m.pollCount++
		successSend = true
		logger.Info("collect...")
		m.Mtx.Unlock()
		time.Sleep(time.Duration(pollInterval) * time.Second)
	}
}

func (m *SelfMonitor) Report() {
	client := resty.New().SetBaseURL("http://" + address)
	for {
		time.Sleep(time.Duration(reportInterval) * time.Second)
		logger.Info("sending...")
		m.Mtx.RLock()
		send(gauge, "Alloc", float64(m.Alloc), client)
		send(gauge, "BuckHashSys", float64(m.BuckHashSys), client)
		send(gauge, "Frees", float64(m.Frees), client)
		send(gauge, "GCPUFraction", float64(m.GCCPUFraction), client)
		send(gauge, "GCSys", float64(m.GCSys), client)
		send(gauge, "HeapAlloc", float64(m.HeapAlloc), client)
		send(gauge, "HeapIdle", float64(m.HeapIdle), client)
		send(gauge, "HeapInuse", float64(m.HeapInuse), client)
		send(gauge, "HeapObjects", float64(m.HeapObjects), client)
		send(gauge, "HeapReleased", float64(m.HeapReleased), client)
		send(gauge, "HeapSys", float64(m.HeapSys), client)
		send(gauge, "LastGC", float64(m.LastGC), client)
		send(gauge, "Lookups", float64(m.Lookups), client)
		send(gauge, "MCacheInuse", float64(m.MCacheInuse), client)
		send(gauge, "MCacheSys", float64(m.MCacheSys), client)
		send(gauge, "MSpanInuse", float64(m.MSpanInuse), client)
		send(gauge, "MSpanSys", float64(m.MSpanSys), client)
		send(gauge, "Mallocs", float64(m.Mallocs), client)
		send(gauge, "NextGC", float64(m.NextGC), client)
		send(gauge, "NumForcedGC", float64(m.NumForcedGC), client)
		send(gauge, "NumGC", float64(m.NumGC), client)
		send(gauge, "OtherSys", float64(m.OtherSys), client)
		send(gauge, "PauseTotalNs", float64(m.PauseTotalNs), client)
		send(gauge, "StackInuse", float64(m.StackInuse), client)
		send(gauge, "StackSys", float64(m.StackSys), client)
		send(gauge, "Sys", float64(m.Sys), client)
		send(gauge, "TotalAlloc", float64(m.TotalAlloc), client)
		send(gauge, "RandomValue", float64(m.randVal), client)
		send("counter", "PollCount", m.pollCount, client)
		if successSend {
			m.pollCount = 0
			logger.Info("success sending")
		}
		m.Mtx.RUnlock()
	}
}
