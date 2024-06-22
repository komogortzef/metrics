package config

import (
	ctx "context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"metrics/internal/agent"
	log "metrics/internal/logger"
	"metrics/internal/server"

	"go.uber.org/zap"
)

var ErrInvalidConfig = errors.New("invalid config")

type Configurable interface {
	Run(ctx.Context)
}

type AppType uint8

const (
	Server AppType = iota
	Agent
)

func Configure(cx ctx.Context, appType AppType, opts ...Option) (Configurable, error) {
	err := log.InitLog()
	if err != nil {
		return nil, fmt.Errorf("init log: %w", err)
	}
	cfg := &config{}
	for _, op := range opts {
		if err := op(cx, cfg); err != nil {
			return nil, err
		}
	}
	switch appType {
	case Server:
		log.Info("MetricManager configuration",
			zap.String("addr", cfg.Address),
			zap.Int("store interval", cfg.StoreInterval),
			zap.Bool("restore", cfg.Restore),
			zap.String("file store", cfg.FileStoragePath),
			zap.String("database", cfg.DBAddress))
		return NewManager(cx, cfg)
	default:
		log.Info("SelfMonitor configuration",
			zap.String("addr", cfg.Address),
			zap.Int("poll interval", cfg.PollInterval),
			zap.Int("report interval", cfg.ReportInterval))
		return NewMonitor(cx, cfg)
	}
}

func NewManager(cx ctx.Context, ops *config) (*server.MetricManager, error) {
	var err error
	manager := &server.MetricManager{
		Server: &http.Server{
			Addr: ops.Address,
		},
	}
	manager.Handler = getRoutes(cx, manager)
	if manager.Store, err = newStorage(cx, ops); err != nil {
		return nil, err
	}
	return manager, nil
}

func NewMonitor(cx ctx.Context, op *config) (*agent.SelfMonitor, error) {
	monitor := &agent.SelfMonitor{
		Mtx:            &sync.RWMutex{},
		Address:        op.Address,
		PollInterval:   op.PollInterval,
		ReportInterval: op.ReportInterval,
	}

	return monitor, nil
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
