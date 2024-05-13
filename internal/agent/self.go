package agent

import (
	"fmt"
	"log"
	"math/rand"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
)

var (
	Address        string
	PollInterval   int
	ReportInterval int
)

const gauge = "gauge"

type SelfMonitor struct {
	runtime.MemStats
	randVal     float64
	pollCount   int64
	successSend bool
	Mtx         *sync.Mutex
}

func (m *SelfMonitor) Collect() {
	for {
		m.Mtx.Lock()
		runtime.ReadMemStats(&m.MemStats)
		m.randVal = rand.Float64()
		m.pollCount++
		m.successSend = true
		log.Printf("DATA COLLECTION: %v\n", m.pollCount)
		m.Mtx.Unlock()

		time.Sleep(time.Duration(PollInterval) * time.Second)
	}
}

func (m *SelfMonitor) Report() {
	client := resty.New()
	for {
		time.Sleep(time.Duration(ReportInterval) * time.Second)
		log.Println("DATA SENDING")
		m.Mtx.Lock()
		m.send(gauge, "Alloc", m.Alloc, client)
		m.send(gauge, "BuckHashSys", m.BuckHashSys, client)
		m.send(gauge, "Frees", m.Frees, client)
		m.send(gauge, "GCPUFraction", m.GCCPUFraction, client)
		m.send(gauge, "GCSys", m.GCSys, client)
		m.send(gauge, "HeapAlloc", m.HeapAlloc, client)
		m.send(gauge, "HeapIdle", m.HeapIdle, client)
		m.send(gauge, "HeapInuse", m.HeapInuse, client)
		m.send(gauge, "HeapObjects", m.HeapObjects, client)
		m.send(gauge, "HeapReleased", m.HeapReleased, client)
		m.send(gauge, "HeapSys", m.HeapSys, client)
		m.send(gauge, "LastGC", m.LastGC, client)
		m.send(gauge, "Lookups", m.Lookups, client)
		m.send(gauge, "MCacheInuse", m.MCacheInuse, client)
		m.send(gauge, "MCacheSys", m.MCacheSys, client)
		m.send(gauge, "MSpanInuse", m.MSpanInuse, client)
		m.send(gauge, "MSpanSys", m.MSpanSys, client)
		m.send(gauge, "Mallocs", m.Mallocs, client)
		m.send(gauge, "NextGC", m.NextGC, client)
		m.send(gauge, "NumForcedGC", m.NumForcedGC, client)
		m.send(gauge, "NumGC", m.NumGC, client)
		m.send(gauge, "OtherSys", m.OtherSys, client)
		m.send(gauge, "PauseTotalNs", m.PauseTotalNs, client)
		m.send(gauge, "StackInuse", m.StackInuse, client)
		m.send(gauge, "StackSys", m.StackSys, client)
		m.send(gauge, "Sys", m.Sys, client)
		m.send(gauge, "TotalAlloc", m.TotalAlloc, client)
		m.send(gauge, "RandomValue", m.randVal, client)
		m.send("counter", "PollCount", m.pollCount, client)

		if m.successSend {
			m.pollCount = 0
			log.Println("ALL DATA HAS BEEN SEND SUCCESSFULLY")
		}
		m.Mtx.Unlock()
	}
}

// отправка одной метрики.
func (m *SelfMonitor) send(kind string, name string, val any, cl *resty.Client) {
	var URL string

	if !strings.Contains(Address, "http://") {
		URL = fmt.Sprintf("http://%s/update/%s/%s/%v", Address, kind, name, val)
	} else {
		URL = fmt.Sprintf("%s/update/%s/%s/%v", Address, kind, name, val)
	}

	_, err := cl.R().Post(URL)
	if err != nil {
		log.Println("No connection:", err)
		m.successSend = false
	}
}

func (m *SelfMonitor) Run() {
	go m.Collect()
	go m.Report()

	select {}
}
