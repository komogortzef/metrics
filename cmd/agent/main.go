package main

import (
	"flag"

	"github.com/komogortzef/metrics/internal/telemetry"
)

func main() {

	run()
}

func run() {
	agent := telemetry.SelfMonitor{}

	flag.StringVar(&agent.Endpoint, "a", "localhost:8080", "Endpoint address")
	flag.IntVar(&agent.PollInterval, "p", 2, "Poll Interval")
	flag.IntVar(&agent.ReportInterval, "r", 10, "Report Interval")

	flag.Parse()

	agent.Run()
}
