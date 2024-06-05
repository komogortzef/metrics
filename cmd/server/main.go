package main

import (
	"metrics/internal/conf"
	l "metrics/internal/logger"
	m "metrics/internal/models"

	"go.uber.org/zap"
)

func main() {
	server, err := conf.Configure(m.MetricsManager, conf.WithEnvCmd)
	if err != nil {
		l.Fatal("Config error", zap.Error(err))
	}

	if err = server.Run(); err != nil {
		l.Fatal("server running error", zap.Error(err))
	}
}
