package config

import (
	ctx "context"
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
		if err := op(cfg); err != nil {
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
		return NewMonitor(cfg)
	}
}

func NewManager(cx ctx.Context, cfg *config) (*server.MetricManager, error) {
	var err error
	manager := &server.MetricManager{Server: http.Server{}}
	manager.Addr = cfg.Address
	manager.Handler = getRoutes(cx, manager)
	if manager.Store, err = newStorage(cx, cfg); err != nil {
		return nil, err
	}
	return manager, nil
}

func NewMonitor(cfg *config) (*agent.SelfMonitor, error) {
	monitor := &agent.SelfMonitor{
		Mtx:            &sync.RWMutex{},
		Address:        cfg.Address,
		PollInterval:   cfg.PollInterval,
		ReportInterval: cfg.ReportInterval,
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
