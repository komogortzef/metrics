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
	Mtx            *sync.RWMutex
	Address        string
	Key            string
	metrics        []s.Metrics
	randVal        float64
	pollCount      int64
	PollInterval   int
	ReportInterval int
}

func (sm *SelfMonitor) collect(cx ctx.Context) {
	for {
		select {
		case <-cx.Done():
			log.Debug("Goodbye from collect")
			return
		default:
			sm.Mtx.Lock()
			runtime.ReadMemStats(&memStats)
			sm.randVal = rand.Float64()
			sm.pollCount++
			sm.collectMetrics()
			log.Debug("collect", zap.Int64("poll", sm.pollCount))
			sm.Mtx.Unlock()
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
			sm.Mtx.RLock()
			if err := sm.sendBatch(cx); err != nil {
				log.Warn("Sending error", zap.Error(err))
				sm.Mtx.RUnlock()
				continue
			}
			sm.pollCount = 0
			sm.Mtx.RUnlock()
			log.Debug("Success sending!")
		}
	}
}

func (sm *SelfMonitor) Run(cx ctx.Context) {
	go sm.collect(cx)
	go sm.report(cx)

	<-cx.Done()
	log.Debug("Goodbye!")
}
