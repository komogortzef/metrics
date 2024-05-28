package main

import (
	"metrics/internal/config"
	"metrics/internal/logger"

	"go.uber.org/zap"
)

func main() {
	agent, err := config.NewAgent(config.WithEnv, config.WithCmd)
	if err != nil {
		logger.Fatal("agent config error", zap.String("err", err.Error()))
	}

	agent.Run()
}
