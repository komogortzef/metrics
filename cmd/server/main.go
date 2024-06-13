package main

import (
	"metrics/internal/config"
	log "metrics/internal/logger"
	"metrics/internal/server"

	"go.uber.org/zap"
)

func main() {
	server, err := config.Configure(&server.MetricManager{},
		config.WithEnvCmd, config.WithRoutes, config.WithStorage)
	if err != nil {
		log.Fatal("server config error", zap.Error(err))
	}

	if err = server.Run(); err != nil {
		log.Fatal("server running error", zap.Error(err))
	}
}
