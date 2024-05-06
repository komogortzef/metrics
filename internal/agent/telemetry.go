package agent

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

const maxArgs = 7

type TelemetryProvider interface {
	Collect()
	Send()
	Perform()
}

// Report ..
type Report struct {
	typ    string
	name   string
	value  string
	status int
}

// SelfMonitor ...
type SelfMonitor struct {
	endpoint       string
	pollInterval   int
	reportInterval int
	sendReports    []Report
	randVal        float64
	pollCount      int64
	runtime.MemStats
}

// Collect ...
func (m *SelfMonitor) Collect() {
	time.Sleep(time.Duration(m.pollInterval) * time.Second)

	for {
		//log.Println("START of data COLLECTION")
		runtime.ReadMemStats(&m.MemStats)
		m.randVal = rand.Float64()
		m.pollCount++
		m.sendReports = m.sendReports[:0]

		time.Sleep(time.Duration(m.pollInterval) * time.Second)
	}
}

// Send ...
func (m *SelfMonitor) Send() {
	client := resty.New()
	time.Sleep(time.Duration(m.reportInterval) * time.Second)
	for {
		log.Println("START of data SENDING")
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

		time.Sleep(time.Duration(m.reportInterval) * time.Second)
	}
}

func (m *SelfMonitor) sendPost(typ, name string, val any, client *resty.Client) {
	var URL string
	if !strings.Contains(m.endpoint, "http://") {
		URL = fmt.Sprintf("http://%s/update/%s/%s/%v", m.endpoint, typ, name, val)
	} else {
		URL = fmt.Sprintf("%s/update/%s/%s/%v", m.endpoint, typ, name, val)
	}
	log.Println("Sending a request to:", URL)

	resp, err := client.R().
		Post(URL)

	if err != nil {
		log.Println("No connection:", err)
	}

	report := Report{
		typ,
		name,
		fmt.Sprintf("%v", val),
		resp.StatusCode(),
	}

	m.sendReports = append(m.sendReports, report)
}

// Run ...
func (m *SelfMonitor) Perform() {
	go m.Collect()
	go m.Send()
	select {}
}

func (m *SelfMonitor) Configure() error {
	if len(os.Args) > maxArgs {
		return errors.New(`max number of configuration parameters(6): 
			-a <host:port> -p <Poll Interval> -r <Report Interval>`)
	}

	addr, ok := os.LookupEnv("ADDRESS")
	if !ok {
		flag.StringVar(&m.endpoint, "a", "localhost:8080", "Endpoint address")
	} else {
		m.endpoint = addr
	}

	poll, ok := os.LookupEnv("POLL_INTERVAL")
	if !ok {
		flag.IntVar(&m.pollInterval, "p", 2, "Poll Interval")
	} else {
		num, _ := strconv.Atoi(poll)
		m.pollInterval = num
	}

	rep, ok := os.LookupEnv("REPORT_INTERVAL")
	if !ok {
		flag.IntVar(&m.reportInterval, "r", 10, "Report Interval")
	} else {
		num, _ := strconv.Atoi(rep)
		m.reportInterval = num
	}

	flag.Parse()

	isHostPort := regexp.MustCompile(`^(.*):(\d+)$`)
	if !isHostPort.MatchString(m.endpoint) {
		return errors.New(
			"the required format of the endpoint address: <host:port>")
	}

	return nil
}

func (m *SelfMonitor) ShowConfig() {
	log.Printf("Monitor configuration:\nEndpoint: %s\nPoll Interval: %v\nReport Interval: %v",
		m.endpoint, m.pollInterval, m.reportInterval)
}

func GetConfig() (SelfMonitor, error) {
	agent := SelfMonitor{}
	err := agent.Configure()
	agent.ShowConfig()

	return agent, err
}
