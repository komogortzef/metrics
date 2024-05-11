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
	ENDPOINT       = "localhost:8080"
	POLLINTERVAL   = 2
	REPORTINTERVAL = 10
)

type Kind string

const (
	gauge   Kind = "gauge"
	counter Kind = "counter"
)

// SelfMonitor объединил в тип данные из runtime, доп. Значения в соответствии
// с заданием.
type SelfMonitor struct {
	runtime.MemStats
	randVal   float64
	pollCount int64
	Mtx       *sync.RWMutex
}

// Collect. Сбор данных.
func (m *SelfMonitor) Collect() {
	for {
		m.Mtx.Lock()
		runtime.ReadMemStats(&m.MemStats)
		m.randVal = rand.Float64()
		m.pollCount++
		m.Mtx.Unlock()
		log.Printf("DATA COLLECTION: %v\n", m.pollCount)
		time.Sleep(time.Duration(POLLINTERVAL) * time.Second)
	}
}

// Send отпарвка данных на сервер.
func (m *SelfMonitor) Send() {
	client := resty.New()
	for {
		time.Sleep(time.Duration(REPORTINTERVAL) * time.Second)
		log.Println("DATA SENDING")
		m.Mtx.RLock()
		sendPost(gauge, "Alloc", m.Alloc, client)
		sendPost(gauge, "BuckHashSys", m.BuckHashSys, client)
		sendPost(gauge, "Frees", m.Frees, client)
		sendPost(gauge, "GCPUFraction", m.GCCPUFraction, client)
		sendPost(gauge, "GCSys", m.GCSys, client)
		sendPost(gauge, "HeapAlloc", m.HeapAlloc, client)
		sendPost(gauge, "HeapIdle", m.HeapIdle, client)
		sendPost(gauge, "HeapInuse", m.HeapInuse, client)
		sendPost(gauge, "HeapObjects", m.HeapObjects, client)
		sendPost(gauge, "HeapReleased", m.HeapReleased, client)
		sendPost(gauge, "HeapSys", m.HeapSys, client)
		sendPost(gauge, "LastGC", m.LastGC, client)
		sendPost(gauge, "Lookups", m.Lookups, client)
		sendPost(gauge, "MCacheInuse", m.MCacheInuse, client)
		sendPost(gauge, "MCacheSys", m.MCacheSys, client)
		sendPost(gauge, "MSpanInuse", m.MSpanInuse, client)
		sendPost(gauge, "MSpanSys", m.MSpanSys, client)
		sendPost(gauge, "Mallocs", m.Mallocs, client)
		sendPost(gauge, "NextGC", m.NextGC, client)
		sendPost(gauge, "NumForcedGC", m.NumForcedGC, client)
		sendPost(gauge, "NumGC", m.NumGC, client)
		sendPost(gauge, "OtherSys", m.OtherSys, client)
		sendPost(gauge, "PauseTotalNs", m.PauseTotalNs, client)
		sendPost(gauge, "StackInuse", m.StackInuse, client)
		sendPost(gauge, "StackSys", m.StackSys, client)
		sendPost(gauge, "Sys", m.Sys, client)
		sendPost(gauge, "TotalAlloc", m.TotalAlloc, client)
		sendPost(gauge, "RandomValue", m.randVal, client)
		sendPost(counter, "PollCount", m.pollCount, client)
		m.Mtx.RUnlock()
		m.pollCount = 0
	}
}

// отправка одной метрики.
func sendPost(kind Kind, name string, val any, client *resty.Client) {
	var URL string
	if !strings.Contains(ENDPOINT, "http://") {
		URL = fmt.Sprintf("http://%s/update/%s/%s/%v", ENDPOINT, kind, name, val)
	} else {
		URL = fmt.Sprintf("%s/update/%s/%s/%v", ENDPOINT, kind, name, val)
	}
	log.Println("Sending a request to:", URL)

	_, err := client.R().Post(URL)

	if err != nil {
		log.Println("No connection:", err)
	}
}

func (m *SelfMonitor) Run() {
	go m.Collect()
	go m.Send()

	select {}
}
