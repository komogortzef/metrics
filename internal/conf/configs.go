package conf

import (
	"fmt"
	"net/http"
	"sync"

	"metrics/internal/agent"
	c "metrics/internal/compress"
	l "metrics/internal/logger"
	m "metrics/internal/models"
	"metrics/internal/server"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type (
	Configurable interface {
		Run() error
	}

	Config interface {
		SetConfig() (Configurable, error)
	}

	agentConfig struct {
		Address        string `env:"ADDRESS" envDefault:"none"`
		PollInterval   int    `env:"POLL_INTERVAL" envDefault:"-1"`
		ReportInterval int    `env:"REPORT_INTERVAL" envDefault:"-1"`
	}

	serverConfig struct {
		Address         string `env:"ADDRESS" envDefault:"none"`
		StoreInterval   int    `env:"STORE_INTERVAL" envDefault:"-1"`
		Restore         bool   `env:"RESTORE" envDefault:"true"`
		FileStoragePath string
	}

	Option func(Config) error
)

func (cfg *serverConfig) SetConfig() (Configurable, error) {
	var manager server.MetricsManager

	router := chi.NewRouter()
	router.Use(l.WithHandlerLog)
	router.Get("/", c.GzipMiddleware(manager.GetAllHandler))
	router.Post("/value/", c.GzipMiddleware(manager.GetJSON))
	router.Get("/value/{type}/{id}", manager.GetHandler)
	router.Post("/update/", c.GzipMiddleware(manager.UpdateJSON))
	router.Post("/update/{type}/{id}/{value}", manager.UpdateHandler)

	manager.Serv = &http.Server{
		Addr:    cfg.Address,
		Handler: router,
	}

	if cfg.FileStoragePath == "" {
		mem := server.NewMemStorage()
		manager.Store = &mem
	} else {
		fileStore, err := server.NewFileStorage(
			cfg.StoreInterval,
			cfg.FileStoragePath,
			cfg.Restore,
		)
		if err != nil {
			return nil, fmt.Errorf("set storage error: %w", err)
		}
		manager.Store = fileStore

		if cfg.StoreInterval > 0 {
			fileStore.StartTicker()
		}
	}

	l.Info("Serv config:",
		zap.String("addr", cfg.Address),
		zap.Int("store interval", cfg.StoreInterval),
		zap.String("file stor path", cfg.FileStoragePath),
		zap.Bool("restore", cfg.Restore),
	)

	return &manager, nil
}

func (cfg *agentConfig) SetConfig() (Configurable, error) {
	agent := agent.SelfMonitor{
		Address:        cfg.Address,
		PollInterval:   cfg.PollInterval,
		ReportInterval: cfg.ReportInterval,
		SendFormat:     "json",
		Mtx:            &sync.RWMutex{},
	}

	l.Info("Agent config",
		zap.String("addr", cfg.Address),
		zap.Int("poll count", cfg.PollInterval),
		zap.Int("report interval", cfg.ReportInterval))

	return &agent, nil
}

func Configure(service m.ServiceType, opts ...Option) (Configurable, error) {
	err := l.InitLog()
	if err != nil {
		return nil, fmt.Errorf("init logger error: %w", err)
	}

	var cfg Config
	switch service {
	case m.MetricsManager:
		l.Info("metric manager")
		cfg = &serverConfig{}
	case m.SelfMonitor:
		l.Info("self monitor")
		cfg = &agentConfig{}
	}

	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			return nil, fmt.Errorf("error while options applying: %w", err)
		}
	}

	return cfg.SetConfig()
}
