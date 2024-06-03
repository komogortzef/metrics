package config

import (
	"sync"

	"metrics/internal/agent"
	l "metrics/internal/logger"

	"go.uber.org/zap"
)

func newAgent(opts *options) (*agent.SelfMonitor, error) {
	agent := agent.SelfMonitor{
		Address:        opts.Address,
		PollInterval:   opts.PollInterval,
		ReportInterval: opts.ReportInterval,
		SendFormat:     "json",
		Mtx:            &sync.RWMutex{},
	}

	l.Info("Agent config",
		zap.String("addr", opts.Address),
		zap.Int("poll count", opts.PollInterval),
		zap.Int("report interval", opts.ReportInterval))

	return &agent, nil
}
