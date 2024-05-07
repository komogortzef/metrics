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
	"strings"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/go-resty/resty/v2"
)

const maxArgs = 7

type TelemetryProvider interface {
	Collect()
	Send()
	Perform()
}

// Report включил скорее для тестов...
type Report struct {
	typ    string
	name   string
	value  string
	status int
}

// SelfMonitor объединил в тип данные из runtime, доп. значения в соответствии
// с заданием и поля конфигурации агента
type SelfMonitor struct {
	Endpoint       string `env:"ADDRESS"`
	PollInterval   int    `env:"POLL_INTERVAL"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	sendReports    []Report
	randVal        float64
	pollCount      int64
	runtime.MemStats
}

// Collect. Сбор данных
func (m *SelfMonitor) Collect() {
	time.Sleep(time.Duration(m.PollInterval) * time.Second)

	for {
		log.Println("START of data COLLECTION")
		runtime.ReadMemStats(&m.MemStats)
		m.randVal = rand.Float64()
		m.pollCount++
		m.sendReports = m.sendReports[:0]

		time.Sleep(time.Duration(m.PollInterval) * time.Second)
	}
}

// Send отпарвка данных на сервер
func (m *SelfMonitor) Send() {
	client := resty.New()
	time.Sleep(time.Duration(m.ReportInterval) * time.Second)
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

		time.Sleep(time.Duration(m.ReportInterval) * time.Second)
	}
}

// отпарвка одной метрики
func (m *SelfMonitor) sendPost(typ, name string, val any, client *resty.Client) {
	var URL string
	if !strings.Contains(m.Endpoint, "http://") {
		URL = fmt.Sprintf("http://%s/update/%s/%s/%v", m.Endpoint, typ, name, val)
	} else {
		URL = fmt.Sprintf("%s/update/%s/%s/%v", m.Endpoint, typ, name, val)
	}
	log.Println("Sending a request to:", URL)

	resp, err := client.R().Post(URL)

	if err != nil {
		log.Println("No connection:", err)
	}

	// заполнение отчета
	report := Report{
		typ,
		name,
		fmt.Sprintf("%v", val),
		resp.StatusCode(),
	}

	m.sendReports = append(m.sendReports, report)
}

// Запуск. плохо понимаю как это работет, но работает...
func (m *SelfMonitor) Perform() {
	go m.Collect()
	go m.Send()
	select {}
}

func (m *SelfMonitor) configure() error {
	if len(os.Args) > maxArgs {
		return errors.New(`max number of configuration parameters(6): 
			-a <host:port> -p <Poll Interval> -r <Report Interval>`)
	}

	err := env.Parse(m)
	if err != nil {
		return err
	}

	endpoint := flag.String("a", "localhost:8080", "input the endpoint address")
	poll := flag.Int("p", 2, "input the poll interval")
	rep := flag.Int("r", 10, "input the report interval")
	flag.Parse()

	// если данные не записались из окружения, то пишем из cmd(или по умолч)
	switch {
	case m.Endpoint == "":
		m.Endpoint = *endpoint
	case m.PollInterval == 0:
		m.PollInterval = *poll
	case m.ReportInterval == 0:
		m.ReportInterval = *rep
	}

	isHostPort := regexp.MustCompile(`^(.*):(\d+)$`)
	if !isHostPort.MatchString(m.Endpoint) {
		return errors.New(
			"the required format of the Endpoint address: <host:port>")
	}

	return nil
}

func GetConfig() (SelfMonitor, error) {
	agent := SelfMonitor{}
	err := agent.configure()

	return agent, err
}
