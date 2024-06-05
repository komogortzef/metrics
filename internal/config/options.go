package config

import (
	"flag"
	"fmt"
	"net/http"
	"sync"

	"metrics/internal/agent"
	c "metrics/internal/compress"
	l "metrics/internal/logger"
	m "metrics/internal/models"
	"metrics/internal/server"

	"github.com/caarlos0/env/v11"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

const (
	noStrValue = "none"
	noIntValue = "-1"
)

type (
	Configurable interface {
		Run() error
	}

	Config interface {
		SetConfig() (Configurable, error)
	}

	agentConfig struct {
		Address        string `env:"ADDRESS, notEmpty"`
		PollInterval   int    `env:"POLL_INTERVAL, notEmpty"`
		ReportInterval int    `env:"REPORT_INTERVAL, notEmpty"`
	}

	serverConfig struct {
		Address         string `env:"ADDRESS, notEmpty"`
		StoreInterval   int    `env:"STORE_INTERVA, notEmpty"`
		FileStoragePath string `env:"FILE_STORAGE_PATH, notEmpty"`
		Restore         string `env:"RESTORE, notEmpty"`
	}

	Option func(Config) error
)

func WithEnv(cfg Config) error {

	return nil
}

func WithEnvAndCmd(cfg Config) error {
	addrOpt := env.Options{
		OnSet: func(tag string, value any, isDefault bool) {
			address := flag.String("a", m.DefaultEndpoint, "Endpoint arg: -a <HOST:PORT>")

		},
	}

	switch c := cfg.(type) {
	case agentConfig:
		poll := flag.Int("p", m.DefaultPollInterval, "Poll Interval arg: -p <sec>")
		report := flag.Int("r", m.DefaultReportInterval, "Report Interval arg: -r <sec>")
		flag.Parse()
	case serverConfig:
		storeInterval := flag.Int("i", m.DefaultStoreInterval, "Store Interval arg: -i <sec>")
		fStorePath := flag.String("f", m.DefaultStorePath, "File path arg: -f </path/to/file>")
		restore := flag.Bool("r", m.DefaultRestore, "Restore option arg: -r <true|false>")
	}

	return nil
}

func (cfg serverConfig) SetConfig() (Configurable, error) {
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
	}

	l.Info("Serv config:",
		zap.String("addr", cfg.Address),
		zap.Int("store interval", cfg.StoreInterval),
		zap.String("file stor path", cfg.FileStoragePath),
		zap.Bool("restore", cfg.Restore),
	)

	return &manager, nil
}

func (cfg agentConfig) SetConfig() (Configurable, error) {
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
		cfg = serverConfig{}
	case m.SelfMonitor:
		cfg = agentConfig{}
	}

	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			return nil, fmt.Errorf("error while options applying: %w", err)
		}
	}

	return cfg.SetConfig()
}
