package agent

import (
	"math/rand"
	"runtime"
	"sync"
	"time"

	log "metrics/internal/logger"
	m "metrics/internal/models"

	"go.uber.org/zap"
)

type SelfMonitor struct {
	metrics        [m.MetricsNumber]m.Metrics
	randVal        float64
	pollCount      int64
	Address        string `env:"ADDRESS" envDefault:"none"`
	PollInterval   int    `env:"POLL_INTERVAL" envDefault:"-1"`
	ReportInterval int    `env:"REPORT_INTERVAL" envDefault:"-1"`
	mtx            *sync.RWMutex
}

func (sm *SelfMonitor) collect() {
	for {
		sm.mtx.Lock()
		runtime.ReadMemStats(&memStats)
		sm.randVal = rand.Float64()
		sm.pollCount++
		sm.getMetrics()
		log.Debug("collect", zap.Int64("poll", sm.pollCount))
		sm.mtx.Unlock()
		time.Sleep(time.Duration(sm.PollInterval) * time.Second)
	}
}

func (sm *SelfMonitor) report() {
	var err error
	for {
	sleep:
		time.Sleep(time.Duration(sm.ReportInterval) * time.Second)
		log.Debug("sending...")
		sm.mtx.RLock()
		for _, metric := range &sm.metrics {
			err = sm.send(metric)
			if err != nil {
				log.Warn("Sending error", zap.Error(err))
				sm.mtx.RUnlock()
				goto sleep
			}
		}
		sm.pollCount = 0
		log.Info("Success sending!")
		sm.mtx.RUnlock()
	}
}

func (sm *SelfMonitor) Run() error {
	log.Info("Agent configuration",
		zap.String("addr", sm.Address),
		zap.Int("poll interval", sm.PollInterval),
		zap.Int("report interval", sm.ReportInterval))

	sm.mtx = &sync.RWMutex{}
	go sm.collect()
	sm.report()

	return nil // чтобы удовлетворить Configurable
}
