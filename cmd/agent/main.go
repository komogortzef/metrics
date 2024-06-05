package main

import (
	c "metrics/internal/config"
	l "metrics/internal/logger"
	m "metrics/internal/models"

	"go.uber.org/zap"
)

func main() {
	agent, err := c.Configure(m.SelfMonitor, c.WithEnv, c.WithCmd)
	if err != nil {
		l.Fatal("agent config error", zap.Error(err))
	}

	if err = agent.Run(); err != nil {
		l.Fatal("agent run error", zap.Error(err))
	}
}
