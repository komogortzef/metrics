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
	var err error
	var options options
	for _, option := range opts {
		err = option(&options)
	}

	// установка глобальных переменных в пакете agent
	agent.SetParam(options.Address, options.PollInterval, options.ReportInterval)

	return &agent.SelfMonitor{Mtx: &sync.Mutex{}}, err
}
