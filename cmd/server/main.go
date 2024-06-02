package main

import (
	"metrics/internal/config"
	l "metrics/internal/logger"

	"go.uber.org/zap"
)

func main() {
	server, err := config.NewServer(config.WithEnvSrv, config.WithCmdSrv)
	if err != nil {
		l.Fatal("Config error", zap.String("error", err.Error()))
	}

	if err = server.ListenAndServe(); err != nil {
		l.Fatal("Couldn't start the sever:", zap.String("err", err.Error()))
	}
}
