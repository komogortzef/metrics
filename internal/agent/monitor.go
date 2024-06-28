package agent

import (
	ctx "context"
	"math/rand"
	"runtime"
	"sync"
	"time"

	log "metrics/internal/logger"
	s "metrics/internal/service"

	"github.com/pquerna/ffjson/ffjson"
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
			runtime.ReadMemStats(&memStats)
			sm.Mtx.Lock()
			sm.randVal = rand.Float64()
			sm.pollCount++
			sm.collectRuntime()
			log.Debug("collectMemStats", zap.Int64("poll", sm.pollCount))
			sm.Mtx.Unlock()
		case <-cx.Done():
			log.Debug("Goodbye from collect")
			return
		}
	}
}

func (sm *SelfMonitor) collectPsStats(cx ctx.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	ticker := time.NewTicker(sm.PollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			psMem, _ = mem.VirtualMemoryWithContext(cx)
			psCPUs, _ = cpu.Percent(0, true)
			sm.Mtx.Lock()
			sm.collectPs()
			log.Debug("collectPs", zap.Int64("poll", sm.pollCount))
			sm.Mtx.Unlock()
		case <-cx.Done():
			log.Debug("Goodbye from collect2")
			return
		}
	}
}

func (sm *SelfMonitor) report(cx ctx.Context, wg *sync.WaitGroup) {
	ticker := time.NewTicker(sm.ReportInterval)
	defer ticker.Stop()
	defer wg.Done()

	errCh := make(chan error, 1)
	sendCh := make(chan []byte, sm.Rate)
	for i := 0; i < sm.Rate; i++ {
		go sm.sendWorker(cx, i+1, sendCh, errCh)
	}

	for {
		select {
		case err := <-errCh:
			if err == nil {
				sm.Mtx.Lock()
				sm.pollCount = 0
				sm.Mtx.Unlock()
			} else {
				<-errCh
			}
		case <-ticker.C:
			sm.Mtx.Lock()
			data, _ := ffjson.Marshal(sm.mets)
			log.Debug("REPORT...")
			sendCh <- data
			sm.Mtx.Unlock()
		case <-cx.Done():
			log.Debug("Goodbye from report")
			return
		}
	}
}

func (sm *SelfMonitor) Run(cx ctx.Context) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()

	wg.Add(3)
	go sm.collectMemStats(cx, wg)
	go sm.collectPsStats(cx, wg)
	go sm.report(cx, wg)

	<-cx.Done()
	log.Debug("Goodbye!")
}
