package main

import (
	"metrics/internal/agent"
	"metrics/internal/config"
	"metrics/internal/logger"

	"go.uber.org/zap"
)

func main() {
	ctx, complete := config.CompletionCtx()
	defer complete()

	agent, err := config.Configure(ctx,
		&agent.SelfMonitor{},
		config.EnvFlagsAgent,
	)
	if err != nil {
		logger.Fatal("agent config error", zap.Error(err))
	}

	agent.Run(ctx)
}
