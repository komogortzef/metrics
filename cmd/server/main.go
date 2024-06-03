package main

import (
	c "metrics/internal/config"
	l "metrics/internal/logger"

	"go.uber.org/zap"
)

func main() {
	server, err := c.Configure(c.WithEnvSrv, c.WithCmdSrv)
	if err != nil {
		l.Fatal("Config error", zap.Error(err))
	}

	if err = server.Run(); err != nil {
		l.Fatal("server running error", zap.Error(err))
	}
}
