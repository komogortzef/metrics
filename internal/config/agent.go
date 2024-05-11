package config

import (
	"fmt"
	"log"
	"os"
	"sync"

	"metrics/internal/agent"
)

const (
	maxArgsAgent = 7
)

type TelemetryProvider interface {
	Collect()
	Send()
	Run()
}

func NewAgent(opts ...Option) (TelemetryProvider, error) {
	var options options
	for _, option := range opts {
		option(&options)
	}

	if len(os.Args) > maxArgsAgent {
		fmt.Fprintln(os.Stderr, "\nInvalid set of args:")
		usage()
		os.Exit(1)
	}

	agent.ENDPOINT = options.Address
	agent.POLLINTERVAL = options.PollInterval
	agent.REPORTINTERVAL = options.ReportInterval

	log.Println(options)

	return &agent.SelfMonitor{Mtx: &sync.RWMutex{}}, nil
}
