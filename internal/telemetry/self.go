// Package telemetry ...
package telemetry

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"runtime"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

const (
	pollInterval   = 2 * time.Second
	reportInterval = 10 * time.Second
	baseURL        = "http://localhost:8080"
)

// Report ..
type Report struct {
	typ    string
	name   string
	value  string
	status int
}

// SelfMonitor ...
type SelfMonitor struct {
	runtime.MemStats
	randVal     float64
	pollCount   int64
	serverAddr  string
	sendReports []Report
}

// NewSelfMonitor ...
func NewSelfMonitor() *SelfMonitor {
	var memSt runtime.MemStats

	runtime.ReadMemStats(&memSt)
	return &SelfMonitor{memSt, 0, 0, baseURL, []Report{}}
}

func (m *SelfMonitor) report() (string, error) {
	if len(m.sendReports) < 1 {
		return "", errors.New("no report")
	}

	var builder strings.Builder

	builder.WriteString("type\tname\tvalue\tsize\tstatus")

	for _, report := range m.sendReports {
		str := fmt.Sprintf(
			"\n%s\t%s\t%s\t%v",
			report.typ,
			report.name,
			report.value,
			report.status,
		)
		builder.WriteString(str)
	}

	return builder.String(), nil
}

// Collect ...
func (m *SelfMonitor) Collect() {
	time.Sleep(pollInterval)

	for {
		log.Println("\nstart of data collection")
		runtime.ReadMemStats(&m.MemStats)
		m.randVal = rand.Float64()
		m.pollCount++
		m.sendReports = m.sendReports[:0]

		log.Println("End of data collection")
		time.Sleep(pollInterval)
	}
}

// Send ...
func (m *SelfMonitor) Send() {
	client := resty.New()
	time.Sleep(reportInterval)
	for {
		log.Println("\nStart of data sending", client)
		m.sendPost("gauge", "Alloc", m.Alloc, client)
		m.sendPost("gauge", "BuckHashSys", m.BuckHashSys, client)
		m.sendPost("gauge", "Frees", m.Frees, client)
		m.sendPost("gauge", "GCPUFraction", m.GCCPUFraction, client)
		m.sendPost("gauge", "GCSys", m.GCSys, client)
		m.sendPost("gauge", "HeapAlloc", m.HeapAlloc, client)
		m.sendPost("gauge", "HeapIdle", m.HeapIdle, client)
		m.sendPost("gauge", "HeapInuse", m.HeapInuse, client)
		m.sendPost("gauge", "HeapObjects", m.HeapObjects, client)
		m.sendPost("gauge", "HeapReleased", m.HeapReleased, client)
		m.sendPost("gauge", "HeapSys", m.HeapSys, client)
		m.sendPost("gauge", "LastGC", m.LastGC, client)
		m.sendPost("gauge", "Lookups", m.Lookups, client)
		m.sendPost("gauge", "MCacheInuse", m.MCacheInuse, client)
		m.sendPost("gauge", "MCacheSys", m.MCacheSys, client)
		m.sendPost("gauge", "MSpanInuse", m.MSpanInuse, client)
		m.sendPost("gauge", "MSpanSys", m.MSpanSys, client)
		m.sendPost("gauge", "Mallocs", m.Mallocs, client)
		m.sendPost("gauge", "NextGC", m.NextGC, client)
		m.sendPost("gauge", "NumForcedGC", m.NumForcedGC, client)
		m.sendPost("gauge", "NumGC", m.NumGC, client)
		m.sendPost("gauge", "OtherSys", m.OtherSys, client)
		m.sendPost("gauge", "PauseTotalNs", m.PauseTotalNs, client)
		m.sendPost("gauge", "StackInuse", m.StackInuse, client)
		m.sendPost("gauge", "StackSys", m.StackSys, client)
		m.sendPost("gauge", "Sys", m.Sys, client)
		m.sendPost("gauge", "TotalAlloc", m.TotalAlloc, client)
		m.sendPost("gauge", "RandomValue", m.randVal, client)
		m.sendPost("counter", "PollCount", m.pollCount, client)
		m.pollCount = 0

		_, err := m.report()
		if err != nil {
			fmt.Println(err)
		}

		log.Println("End of data sending")
		time.Sleep(reportInterval)
	}
}

func (m *SelfMonitor) sendPost(typ, name string, val any, client *resty.Client) {
	URL := fmt.Sprintf("%s/update/%s/%s/%v", m.serverAddr, typ, name, val)
	log.Println("Sending a request to:", URL)

	log.Println("setting a connection...")

	resp, err := client.R().
		Post(URL)

	if err != nil {
		log.Println("There is no connection to the server")
	}

	log.Println("The connection is established and data is data has been sent")

	report := Report{
		typ,
		name,
		fmt.Sprintf("%v", val),
		resp.StatusCode(),
	}

	m.sendReports = append(m.sendReports, report)
}

// Run ...
func (m *SelfMonitor) Run() {
	log.Println("\nmonitor running...")

	go m.Collect()
	go m.Send()
	log.Println("monitor is running")
	select {}
}
