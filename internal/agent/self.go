package agent

import (
	ctx "context"
	"math/rand"
	"runtime"
	"sync"
	"time"

	log "metrics/internal/logger"
	s "metrics/internal/service"

	"go.uber.org/zap"
)

type SelfMonitor struct {
	metrics        []s.Metrics
	randVal        float64
	pollCount      int64
	Address        string `env:"ADDRESS" envDefault:"none"`
	PollInterval   int    `env:"POLL_INTERVAL" envDefault:"-1"`
	ReportInterval int    `env:"REPORT_INTERVAL" envDefault:"-1"`
	mtx            *sync.RWMutex
}

func (sm *SelfMonitor) collect(cx ctx.Context) {
	for {
		select {
		case <-cx.Done():
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

func (sm *SelfMonitor) report(cx ctx.Context) {
	for {
		select {
		case <-cx.Done():
			log.Debug("Goodbye from report")
			return
		default:
			time.Sleep(time.Duration(sm.ReportInterval) * time.Second)
			log.Debug("sending...")
			sm.mtx.RLock()
			if err := sm.sendBatch(); err != nil {
				if err = s.Retry(cx, sm.sendBatch); err != nil {
					log.Warn("Sending error", zap.Error(err))
					sm.mtx.RUnlock()
					continue
				}
			}
			sm.pollCount = 0
			sm.mtx.RUnlock()
			log.Debug("Success sending!")
		}
	}
}

func (sm *SelfMonitor) Run(cx ctx.Context) {
	sm.mtx = &sync.RWMutex{}
	go sm.collect(cx)
	go sm.report(cx)

	<-cx.Done()
	log.Debug("Goodbye!")
}
