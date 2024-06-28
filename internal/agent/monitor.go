package agent

import (
	ctx "context"
	"encoding/json"
	"math/rand"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	log "metrics/internal/logger"
	s "metrics/internal/service"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
	"go.uber.org/zap"
)

type SelfMonitor struct {
	Mtx            *sync.RWMutex
	Address        string
	Key            string
	mets           []*s.Metrics
	randVal        float64
	pollCount      int64
	PollInterval   time.Duration
	ReportInterval time.Duration
	Rate           int
}

func (sm *SelfMonitor) collectMemStats(cx ctx.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	ticker := time.NewTicker(sm.PollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			sm.updateMemStats()
		case <-cx.Done():
			log.Debug("Stopped collecting memory stats")
			return
		}
	}
}

func (sm *SelfMonitor) updateMemStats() {
	runtime.ReadMemStats(&memStats)
	sm.Mtx.Lock()
	defer sm.Mtx.Unlock()
	sm.randVal = rand.Float64()
	atomic.AddInt64(&sm.pollCount, 1)
	sm.collectRuntime()
	log.Debug("Collected memory stats", zap.Int64("poll", sm.pollCount))
}

func (sm *SelfMonitor) collectPsStats(cx ctx.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	ticker := time.NewTicker(sm.PollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			sm.updatePsStats(cx)
		case <-cx.Done():
			log.Debug("Stopped collecting process stats")
			return
		}
	}
}

func (sm *SelfMonitor) updatePsStats(cx ctx.Context) {
	psMem, _ = mem.VirtualMemoryWithContext(cx)
	psCPUs, _ = cpu.Percent(0, true)
	sm.Mtx.Lock()
	defer sm.Mtx.Unlock()
	sm.collectPs()
	atomic.AddInt64(&sm.pollCount, 1)
	log.Debug("Collected process stats", zap.Int64("poll", sm.pollCount))
}

func (sm *SelfMonitor) report(cx ctx.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	ticker := time.NewTicker(sm.ReportInterval)
	defer ticker.Stop()
	errCh := make(chan error, 1)
	sendCh := make(chan []byte, sm.Rate)

	for i := 0; i < sm.Rate; i++ {
		go sm.sendWorker(cx, i+1, sendCh, errCh)
	}

	for {
		select {
		case err := <-errCh:
			if err == nil {
				atomic.SwapInt64(&sm.pollCount, 0)
			} else {
				log.Warn("Failed to send data", zap.Error(err))
				<-errCh
			}
		case <-ticker.C:
			sm.sendMetrics(sendCh)
		case <-cx.Done():
			log.Debug("Stopped reporting metrics")
			return
		}
	}
}

func (sm *SelfMonitor) sendMetrics(sendCh chan []byte) {
	sm.Mtx.Lock()
	defer sm.Mtx.Unlock()
	data, _ := json.Marshal(sm.mets)
	log.Debug("Reporting metrics")
	sendCh <- data
}

func (sm *SelfMonitor) Run(cx ctx.Context) {
	wg := &sync.WaitGroup{}
	wg.Add(3)
	go sm.collectMemStats(cx, wg)
	go sm.collectPsStats(cx, wg)
	go sm.report(cx, wg)
	wg.Wait()
	log.Debug("Stopped all monitoring")
}
