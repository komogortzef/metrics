package config

import (
	"fmt"
	"sync"

	"metrics/internal/agent"
	"metrics/internal/logger"
)

type Telemetry interface {
	Collect()
	Report()
}

func NewAgent(opts ...Option) (Telemetry, error) {
	if err := logger.InitLog(); err != nil {
		return nil, fmt.Errorf("init logger error: %w", err)
	}

	var err error
	var options options
	for _, option := range opts {
		err = option(&options)
	}

	agent.SetCond(options.Address, "", options.PollInterval, options.ReportInterval)

	return &agent.SelfMonitor{Mtx: &sync.RWMutex{}}, err
}
