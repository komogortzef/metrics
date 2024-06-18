package agent

import (
	"context"
	"math/rand"
	"runtime"
	"sync"
	"time"

	log "metrics/internal/logger"
	m "metrics/internal/service"

	"go.uber.org/zap"
)

type SelfMonitor struct {
	metrics        []m.Metrics
	randVal        float64
	pollCount      int64
	Address        string `env:"ADDRESS" envDefault:"none"`
	PollInterval   int    `env:"POLL_INTERVAL" envDefault:"-1"`
	ReportInterval int    `env:"REPORT_INTERVAL" envDefault:"-1"`
	mtx            *sync.RWMutex
}

func (sm *SelfMonitor) collect(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Debug("Goodbye from collect")
			return
		default:
			sm.mtx.Lock()
			runtime.ReadMemStats(&memStats)
			sm.randVal = rand.Float64()
			sm.pollCount++
			sm.collectMetrics()
			log.Debug("collect", zap.Int64("poll", sm.pollCount))
			sm.mtx.Unlock()
			time.Sleep(time.Duration(sm.PollInterval) * time.Second)
		}
	}
}

func (sm *SelfMonitor) report(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Debug("Goodbye from report")
			return
		default:
			time.Sleep(time.Duration(sm.ReportInterval) * time.Second)
			log.Debug("sending...")
			sm.mtx.RLock()
			if err := m.Retry(ctx, sm.sendBatch); err != nil {
				log.Warn("Sending error", zap.Error(err))
				sm.mtx.RUnlock()
				continue
			}
			sm.pollCount = 0
			sm.mtx.RUnlock()
			log.Debug("Success sending!")
		}
	}
}

func (sm *SelfMonitor) Run(ctx context.Context) {
	log.Info("Agent configuration",
		zap.String("addr", sm.Address),
		zap.Int("poll interval", sm.PollInterval),
		zap.Int("report interval", sm.ReportInterval))

	sm.mtx = &sync.RWMutex{}
	go sm.collect(ctx)
	go sm.report(ctx)

	<-ctx.Done()
	log.Debug("Goodbye!")
}
