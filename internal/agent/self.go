package agent

import (
	"math/rand"
	"runtime"
	"sync"
	"time"

	l "metrics/internal/logger"
	m "metrics/internal/models"

	"go.uber.org/zap"
)

type SelfMonitor struct {
	metrics   [m.MetricsNumber]m.Metrics
	randVal   float64
	pollCount int64
	Mtx       *sync.RWMutex
}

func (sm *SelfMonitor) Collect() {
	for {
		sm.Mtx.Lock()
		runtime.ReadMemStats(&memStats)
		sm.randVal = rand.Float64()
		sm.pollCount++
		sm.getMetrics()
		l.Debug("collect", zap.Int64("poll", sm.pollCount))
		sm.Mtx.Unlock()
		time.Sleep(time.Duration(pollInterval) * time.Second)
	}
}

func (sm *SelfMonitor) Report() {
	var err error
	for {
	sleep:
		time.Sleep(time.Duration(reportInterval) * time.Second)
		l.Debug("sending...")
		sm.Mtx.RLock()
		for _, metric := range &sm.metrics {
			err = send(metric)
			if err != nil {
				l.Warn("Sending error", zap.Error(err))
				sm.Mtx.RUnlock()
				goto sleep
			}
		}
		sm.pollCount = 0
		l.Info("Success sending!")
		sm.Mtx.RUnlock()
	}
}

func (sm *SelfMonitor) Run() {
	go sm.Collect()
	sm.Report()
}
