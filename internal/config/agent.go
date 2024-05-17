package config

import (
	"fmt"
	"sync"

	"metrics/internal/agent"
	"metrics/internal/logger"
)

type TelemetryProvider interface {
	Collect()
	Report()
	Run()
}

func NewAgent(opts ...Option) (TelemetryProvider, error) {
	if err := logger.InitLog(); err != nil {
		return nil, fmt.Errorf("init logger error: %w", err)
	}
	logger.Info("Creating an agent...")

	var err error
	var options options
	for _, option := range opts {
		err = option(&options)
	}

	// установка глобальных переменных в пакете agent
	agent.SetParam(options.Address, options.PollInterval, options.ReportInterval)

	return &agent.SelfMonitor{Mtx: &sync.Mutex{}}, err
}
