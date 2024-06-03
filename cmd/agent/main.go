package main

import (
	c "metrics/internal/config"
	l "metrics/internal/logger"

	"go.uber.org/zap"
)

func main() {
	agent, err := c.Configure(c.WithEnvAg, c.WithCmdAg)
	if err != nil {
		l.Fatal("agent config error", zap.Error(err))
	}

	if err = agent.Run(); err != nil {
		l.Fatal("agent run error", zap.Error(err))
	}
}
