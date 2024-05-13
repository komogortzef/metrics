package config

import (
	"sync"

	"metrics/internal/agent"
)

type TelemetryProvider interface {
	Collect()
	Report()
	Run()
}

func NewAgent(opts ...Option) (TelemetryProvider, error) {
	var options options
	for _, option := range opts {
		option(&options)
	}

	agent.Address = options.Address
	agent.PollInterval = options.PollInterval
	agent.ReportInterval = options.ReportInterval

	return &agent.SelfMonitor{Mtx: &sync.Mutex{}}, nil
}
