package main

import (
	"metrics/internal/agent"
	"metrics/internal/config"
	log "metrics/internal/logger"

	"go.uber.org/zap"
)

func main() {
	ctx, complete := config.CompletionCtx()
	defer complete()

	agent, err := config.Configure(ctx, &agent.SelfMonitor{}, config.WithEnvCmd)
	if err != nil {
		log.Fatal("agent config error", zap.Error(err))
	}

	agent.Run(ctx)
}
