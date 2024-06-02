package main

import (
	"metrics/internal/config"
	l "metrics/internal/logger"

	"go.uber.org/zap"
)

func main() {
	agent, err := config.NewAgent(config.WithEnvAg, config.WithCmdAg)
	if err != nil {
		l.Fatal("agent config error", zap.String("err", err.Error()))
	}

	agent.Run()
}
