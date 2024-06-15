package main

import (
	c "metrics/internal/config"
	log "metrics/internal/logger"
	"metrics/internal/server"

	"go.uber.org/zap"
)

func main() {
	server, err := c.Configure(&server.MetricManager{},
		c.WithEnvCmd,
		c.WithRoutes,
		c.WithStorage,
	)
	if err != nil {
		log.Fatal("Config error", zap.Error(err))
	}

	if err = server.Run(); err != nil {
		log.Fatal("server running error", zap.Error(err))
	}
}
