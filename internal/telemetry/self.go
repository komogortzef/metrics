package telemetry

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"strings"
	"time"
)

const (
	pollInterval   = 2 * time.Second
	reportInterval = 10 * time.Second
	baseURL        = "http://localhost:8080"
)

type Report struct {
	tp    string
	name  string
	value string
	resp  *http.Response
}

type SelfMonitor struct {
	runtime.MemStats
	randVal     float64
	pollCount   int64
	serverAddr  string
	sendReports []Report
}

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

	for _, r := range m.sendReports {
		s := fmt.Sprintf(
			"\n%s\t%s\t%s\t%v\t%v",
			r.tp,
			r.name,
			r.value,
			r.resp.Status,
			r.resp.ContentLength,
		)
		builder.WriteString(s)
	}

	return builder.String(), nil
}

func (m *SelfMonitor) Collect() {
	time.Sleep(pollInterval)
	for {
		log.Println("\nstart of data collection")
		runtime.ReadMemStats(&m.MemStats)
		m.randVal = rand.Float64()
		m.pollCount += 1
		m.sendReports = m.sendReports[:0]
		log.Println("End of data collection")
		time.Sleep(pollInterval)
	}
}

func (m *SelfMonitor) Send() {
	time.Sleep(reportInterval)
	for {
		log.Println("\nStart of data sending")
		m.sendPost("gauge", "Alloc", m.Alloc)
		m.sendPost("gauge", "BuckHashSys", m.BuckHashSys)
		m.sendPost("gauge", "Frees", m.Frees)
		m.sendPost("gauge", "GCPUFraction", m.GCCPUFraction)
		m.sendPost("gauge", "GCSys", m.GCSys)
		m.sendPost("gauge", "HeapAlloc", m.HeapAlloc)
		m.sendPost("gauge", "HeapIdle", m.HeapIdle)
		m.sendPost("gauge", "HeapInuse", m.HeapInuse)
		m.sendPost("gauge", "HeapObjects", m.HeapObjects)
		m.sendPost("gauge", "HeapReleased", m.HeapReleased)
		m.sendPost("gauge", "HeapSys", m.HeapSys)
		m.sendPost("gauge", "LastGC", m.LastGC)
		m.sendPost("gauge", "Lookups", m.Lookups)
		m.sendPost("gauge", "MCacheInuse", m.MCacheInuse)
		m.sendPost("gauge", "MCacheSys", m.MCacheSys)
		m.sendPost("gauge", "MSpanInuse", m.MSpanInuse)
		m.sendPost("gauge", "MSpanSys", m.MSpanSys)
		m.sendPost("gauge", "Mallocs", m.Mallocs)
		m.sendPost("gauge", "NextGC", m.NextGC)
		m.sendPost("gauge", "NumForcedGC", m.NumForcedGC)
		m.sendPost("gauge", "NumGC", m.NumGC)
		m.sendPost("gauge", "OtherSys", m.OtherSys)
		m.sendPost("gauge", "PauseTotalNs", m.PauseTotalNs)
		m.sendPost("gauge", "StackInuse", m.StackInuse)
		m.sendPost("gauge", "StackSys", m.StackSys)
		m.sendPost("gauge", "Sys", m.Sys)
		m.sendPost("gauge", "TotalAlloc", m.TotalAlloc)
		m.sendPost("gauge", "RandomValue", m.randVal)
		m.sendPost("counter", "PollCount", m.pollCount)
		m.pollCount = 0

		_, err := m.report()
		if err != nil {
			fmt.Println(err)
		}

		log.Println("End of data sending")
		time.Sleep(reportInterval)
	}
}

func (m *SelfMonitor) sendPost(tp, name string, val any) {
	url := fmt.Sprintf("%s/update/%s/%s/%v", m.serverAddr, tp, name, val)
	log.Println("\nsending a request to:", url)

	log.Println("setting a connection...")
	resp, err := http.Post(url, "text/plain", nil)
	if err != nil {
		log.Println("There is no connection to the server")
	}
	defer resp.Body.Close()
	log.Println("The connection is established and data is data has been sent")

	report := Report{
		tp,
		name,
		fmt.Sprintf("%v", val),
		resp,
	}

	m.sendReports = append(m.sendReports, report)
}

func (m *SelfMonitor) Run() {
	log.Println("\nmonitor running...")
	go m.Collect()
	go m.Send()
	log.Println("monitor is running")
	select {}
}
