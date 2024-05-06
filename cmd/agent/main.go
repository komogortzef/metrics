package main

import (
	"errors"
	"flag"
	"os"
	"regexp"

	"github.com/komogortzef/metrics/internal/telemetry"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

const maxArgs = 7

func run() error {
	agent := telemetry.SelfMonitor{}

	if len(os.Args) > maxArgs {
		return errors.New("max number of configuration parameters: -a <host:port> -p <Poll Interval> -r <Report Interval>")
	}

	flag.StringVar(&agent.Endpoint, "a", "localhost:8080", "Endpoint address")
	flag.IntVar(&agent.PollInterval, "p", 2, "Poll Interval")
	flag.IntVar(&agent.ReportInterval, "r", 10, "Report Interval")
	flag.Parse()

	var isHostPort = regexp.MustCompile(`^(.*):(\d+)$`)
	if !isHostPort.MatchString(agent.Endpoint) {
		return errors.New("the required format of the endpoint address: <host:port>")
	}

	agent.Run()

	return nil
}
