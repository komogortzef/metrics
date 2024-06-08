package main

import (
	"metrics/internal/conf"
	l "metrics/internal/logger"

	"go.uber.org/zap"
)

func main() {
	agent, err := conf.Configure(conf.MetricAgent, conf.WithEnvCmd)
	if err != nil {
		l.Fatal("agent config error", zap.Error(err))
	}

	if err = agent.Run(); err != nil {
		l.Fatal("agent run error", zap.Error(err))
	}
}
