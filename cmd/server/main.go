package main

import (
	"metrics/internal/config"
	"metrics/internal/logger"
	"metrics/internal/server"

	"go.uber.org/zap"
)

func main() {
	ctx, complete := config.CompletionCtx()
	defer complete()

	server, err := config.Configure(ctx, &server.MetricManager{},
		config.EnvFlagsServer,
		config.WithRoutes,
		config.WithStorage,
	)
	if err != nil {
		logger.Fatal("server config error", zap.Error(err))
	}

	server.Run(ctx)
}
