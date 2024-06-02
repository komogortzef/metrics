package config

import (
	"fmt"
	"sync"

	"metrics/internal/agent"
	"metrics/internal/logger"

	"go.uber.org/zap"
)

func NewAgent(opts ...func(*options)) (*agent.SelfMonitor, error) {
	if err := logger.InitLog(); err != nil {
		return nil, fmt.Errorf("init logger error: %w", err)
	}

	var options options
	for _, option := range opts {
		option(&options)
	}

	agent.SetCond(
		options.Address, "json", options.PollInterval, options.ReportInterval)

	logger.Info("Agent config",
		zap.String("addr", options.Address),
		zap.Int("poll count", options.PollInterval),
		zap.Int("report interval", options.ReportInterval))

	return &agent.SelfMonitor{Mtx: &sync.RWMutex{}}, nil
}
