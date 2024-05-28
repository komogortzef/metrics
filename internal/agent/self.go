package agent

import (
	"math/rand"
	"runtime"
	"sync"
	"time"

	"metrics/internal/logger"
)

type SelfMonitor struct {
	runtime.MemStats
	randVal   float64
	pollCount float64
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
	for {
		time.Sleep(time.Duration(reportInterval) * time.Second)
		logger.Info("sending...")
		m.Mtx.RLock()
		send(gauge, "Alloc", float64(m.Alloc))
		send(gauge, "BuckHashSys", float64(m.BuckHashSys))
		send(gauge, "Frees", float64(m.Frees))
		send(gauge, "GCCPUFraction", m.GCCPUFraction)
		send(gauge, "GCSys", float64(m.GCSys))
		send(gauge, "HeapAlloc", float64(m.HeapAlloc))
		send(gauge, "HeapIdle", float64(m.HeapIdle))
		send(gauge, "HeapInuse", float64(m.HeapInuse))
		send(gauge, "HeapObjects", float64(m.HeapObjects))
		send(gauge, "HeapReleased", float64(m.HeapReleased))
		send(gauge, "HeapSys", float64(m.HeapSys))
		send(gauge, "LastGC", float64(m.LastGC))
		send(gauge, "Lookups", float64(m.Lookups))
		send(gauge, "MCacheInuse", float64(m.MCacheInuse))
		send(gauge, "MCacheSys", float64(m.MCacheSys))
		send(gauge, "MSpanInuse", float64(m.MSpanInuse))
		send(gauge, "MSpanSys", float64(m.MSpanSys))
		send(gauge, "Mallocs", float64(m.Mallocs))
		send(gauge, "NextGC", float64(m.NextGC))
		send(gauge, "NumForcedGC", float64(m.NumForcedGC))
		send(gauge, "NumGC", float64(m.NumGC))
		send(gauge, "OtherSys", float64(m.OtherSys))
		send(gauge, "PauseTotalNs", float64(m.PauseTotalNs))
		send(gauge, "StackInuse", float64(m.StackInuse))
		send(gauge, "StackSys", float64(m.StackSys))
		send(gauge, "Sys", float64(m.Sys))
		send(gauge, "TotalAlloc", float64(m.TotalAlloc))
		send(gauge, "RandomValue", float64(m.randVal))
		send(counter, "PollCount", m.pollCount)
		if successSend {
			m.pollCount = 0
			logger.Info("success sending")
		}
		m.Mtx.RUnlock()
	}
}

func (m *SelfMonitor) Run() {
	go m.Collect()
	m.Report()
}
