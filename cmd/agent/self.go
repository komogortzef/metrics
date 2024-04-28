package main

import (
	"fmt"
	"log"
	"net/http"
	"reflect"
	"runtime"
	"sync"
	"time"
)

const (
	pollInterval   = 2 * time.Second
	reportInterval = 10 * time.Second
	baseUrl        = "http://localhost:8080/update"
)

func NewSelf() SelfMonitor {
	log.Println("Инициализация монитора")
	monitor := SelfMonitor{}

	monitor.gauges = map[string]struct{}{
		"Alloc":         {},
		"BuckHashSys":   {},
		"Frees":         {},
		"GCCPUFraction": {},
		"GCSys":         {},
		"HeapAlloc":     {},
		"HeapIdle":      {},
		"HeapInuse":     {},
		"HeapObjects":   {},
		"HeapReleased":  {},
		"HeapSys":       {},
		"LastGC":        {},
		"Lookups":       {},
		"MCacheInuse":   {},
		"MCacheSys":     {},
		"MSpanInuse":    {},
		"MSpanSys":      {},
		"Mallocs":       {},
		"NextGC":        {},
		"NumForcedGC":   {},
		"NumGC":         {},
		"OtherSys":      {},
		"PauseTotalNs":  {},
		"StackInuse":    {},
		"StackSys":      {},
		"Sys":           {},
		"TotalAlloc":    {},
	}

	return monitor
}

type SelfMonitor struct {
	memStats  runtime.MemStats
	gauges    map[string]struct{}
	pollCount int64
}

func (monitor *SelfMonitor) Collect(wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		log.Println("данные прочитаны в структуру")
		runtime.ReadMemStats(&monitor.memStats)
		monitor.pollCount += 1
		log.Println("Монитор заполнен")
		time.Sleep(pollInterval)
	}
}

func (monitor *SelfMonitor) Send(client *http.Client, wg *sync.WaitGroup) {
	defer wg.Done()
	log.Println("Чтение данных из монитора")
	for {
		r := reflect.ValueOf(monitor.memStats)
		iType := r.Type()

		for i := 0; i < r.NumField(); i++ {
			fieldName := iType.Field(i).Name
			if _, ok := monitor.gauges[fieldName]; !ok {
				continue
			}
			fieldVal := fmt.Sprintf("%v", r.Field(i).Interface())
			url := fmt.Sprintf("%s/gauge/%s/%s", baseUrl, fieldName, fieldVal)
			log.Println("url адрес:", url)
			log.Printf("метрика %s:  значение - %v\n", fieldName, fieldVal)

			log.Println("СЧЕТЧИК:", monitor.pollCount)
			sendReq(client, url)

			url = fmt.Sprintf("%s/counter/PollCount/%v",
				baseUrl,
				monitor.pollCount,
			)
			log.Println("url адрес:", url)
			sendReq(client, url)
		}

		time.Sleep(reportInterval)
	}
}

func sendReq(client *http.Client, url string) {
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		log.Fatalln(err)
	}
	req.Header.Set("Content-Type", "text/plain")
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
	}
	log.Println(resp)
}
