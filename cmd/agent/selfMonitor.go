package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"sync"
	"time"
)

const (
	pollInterval   = 2 * time.Second
	reportInterval = 10 * time.Second
	baseUrl        = "http://localhost:8080/update"
)

type SelfMonitor map[string]any

func NewSelfMonitor() SelfMonitor {
	return make(SelfMonitor)
}

func (monitor SelfMonitor) Collect(wg *sync.WaitGroup) {
	memStats := runtime.MemStats{}
	monitor["PollCounter"] = int64(0)
	defer wg.Done()

	for {
		runtime.ReadMemStats(&memStats)

		// Сбор метрик из пакета runtime
		monitor["Alloc"] = float64(memStats.Alloc)
		monitor["BuckHashSys"] = float64(memStats.BuckHashSys)
		monitor["Frees"] = float64(memStats.Frees)
		monitor["GCCPUFraction"] = memStats.GCCPUFraction
		monitor["GCSys"] = float64(memStats.GCSys)
		monitor["HeapAlloc"] = float64(memStats.HeapAlloc)
		monitor["HeapIdle"] = float64(memStats.HeapIdle)
		monitor["HeapInuse"] = float64(memStats.HeapInuse)
		monitor["HeapObjects"] = float64(memStats.HeapObjects)
		monitor["HeapReleased"] = float64(memStats.HeapReleased)
		monitor["HeapSys"] = float64(memStats.HeapSys)
		monitor["LastGC"] = float64(memStats.LastGC)
		monitor["Lookups"] = float64(memStats.Lookups)
		monitor["MCacheInuse"] = float64(memStats.MCacheInuse)
		monitor["MCacheSys"] = float64(memStats.MCacheSys)
		monitor["MSpanInuse"] = float64(memStats.MSpanInuse)
		monitor["MSpanSys"] = float64(memStats.MSpanSys)
		monitor["Mallocs"] = float64(memStats.Mallocs)
		monitor["NextGC"] = float64(memStats.NextGC)
		monitor["NumForcedGC"] = float64(memStats.NumForcedGC)
		monitor["NumGC"] = float64(memStats.NumGC)
		monitor["OtherSys"] = float64(memStats.OtherSys)
		monitor["PauseTotalNs"] = float64(memStats.PauseTotalNs)
		monitor["StackInuse"] = float64(memStats.StackInuse)
		monitor["StackSys"] = float64(memStats.StackSys)
		monitor["Sys"] = float64(memStats.Sys)
		monitor["TotalAlloc"] = float64(memStats.TotalAlloc)
		monitor["RandomValue"] = rand.Float64

		if num, ok := monitor["PollCounter"].(int64); ok {
			monitor["PollCounter"] = num + 1
		} else {
			log.Fatalln("неверный тип PollCounter")
		}
		time.Sleep(pollInterval)
	}
}

func (monitor SelfMonitor) Send(client *http.Client, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		for name, value := range monitor {
			url := buildUrl(name, value)
			log.Println("url:  ", url)

			req, err := http.NewRequest("POST", url, nil)
			if err != nil {
				log.Fatalln("Ошибка создания запроса")
			}
			req.Header.Set("Content-Type", "text/plain")

			resp, err := client.Do(req)
			if err != nil {
				log.Fatalln("Ошибка подключения к серверу")
			}

			if resp.StatusCode != http.StatusOK {
				log.Println("Сервер не принял данные")
			}
		}
		log.Println()
		time.Sleep(reportInterval)
	}
}

func buildUrl(name string, val any) string {
	typ := "gauge"
	switch val.(type) {
	case int64:
		typ = "counter"
	}
	return fmt.Sprintf("%s/%s/%s/%v", baseUrl, typ, name, val)
}
