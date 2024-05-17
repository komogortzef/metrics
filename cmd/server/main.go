package main

import (
	"metrics/internal/config"
	"metrics/internal/logger"

	"go.uber.org/zap"
)

func main() {
	server, err := config.NewServer(config.WithEnv, config.WithCmd)
	if err != nil {
		logger.Fatal("Couldn't configure the server",
			zap.String("address", server.Addr),
			zap.String("error", err.Error()),
		)
	}

	logger.Info("Server is running on:", zap.String("address", server.Addr))
	if err = server.ListenAndServe(); err != nil {
		logger.Fatal("Couldn't start the sever:",
			zap.String("address", server.Addr),
			zap.String("error", err.Error()),
		)
	}
}
