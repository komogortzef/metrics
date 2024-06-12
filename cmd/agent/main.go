package main

import (
	"metrics/internal/agent"
	c "metrics/internal/config"
	log "metrics/internal/logger"

	"go.uber.org/zap"
)

func main() {
	agent, err := c.Configure(&agent.SelfMonitor{}, c.WithEnvCmd)
	if err != nil {
		log.Fatal("agent config error", zap.Error(err))
	}

	if err = agent.Run(); err != nil {
		log.Fatal("agent run error", zap.Error(err))
	}
}
