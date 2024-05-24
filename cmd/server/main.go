package main

import (
	"metrics/internal/config"
	"metrics/internal/logger"

	"go.uber.org/zap"
)

func main() {
	server, err := config.NewServer(config.WithEnv, config.WithCmd)
	if err != nil {
		logger.Fatal("Config error", zap.String("error", err.Error()))
	}

	if err = server.ListenAndServe(); err != nil {
		logger.Fatal("Couldn't start the sever:", zap.String("err", err.Error()))
	}
}
