package config

import (
	ctx "context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"metrics/internal/agent"
	log "metrics/internal/logger"
	"metrics/internal/server"

	"go.uber.org/zap"
)

var ErrInvalidConfig = errors.New("invalid config")

type Config interface {
	Run(ctx.Context)
}

type Option func(ctx.Context, Config) error

func Configure(ctx ctx.Context, cfg Config, opts ...Option) (Config, error) {
	err := log.InitLog()
	if err != nil {
		return nil, fmt.Errorf("init log: %w", err)
	}
	for _, opt := range opts {
		if err := opt(ctx, cfg); err != nil {
			return nil, err
		}
	}
	switch c := cfg.(type) {
	case *server.MetricManager:
		log.Info("MetricManager configuration",
			zap.String("addr", c.Address),
			zap.Int("store interval", c.StoreInterval),
			zap.Bool("restore", c.Restore),
			zap.String("file store", c.FileStoragePath),
			zap.String("database", c.DBAddress))
	case *agent.SelfMonitor:
		log.Info("SelfMonitor configuration",
			zap.String("addr", c.Address),
			zap.Int("poll interval", c.PollInterval),
			zap.Int("report interval", c.ReportInterval))
	}
	return cfg, nil
}

func CompletionCtx() (ctx.Context, ctx.CancelFunc) {
	cx, complete := ctx.WithCancel(ctx.Background())
	signChan := make(chan os.Signal, 1)
	signal.Notify(signChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-signChan
		complete()
	}()
	return cx, complete
}
