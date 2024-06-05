package conf

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
		FileStoragePath string `env:"FILE_STORAGE_PATH" envDefault:"none"`
		Restore         bool   `env:"RESTORE" envDefault:"true"`
	}

	Option func(Config) error
)

func WithEnvCmd(cfg Config) error {
	err := env.Parse(cfg)
	if err != nil {
		return fmt.Errorf("env parse error: %w", err)
	}

	addr := flag.String("a", m.DefaultEndpoint, "Endpoint arg: -a <host:port>")
	switch c := cfg.(type) {
	case *agentConfig:
		poll := flag.Int("p", m.DefaultPollInterval, "Poll Interval arg: -p <sec>")
		rep := flag.Int("r", m.DefaultReportInterval, "Report interval arg: -r <sec>")
		flag.Parse()
		if c.Address == "none" {
			c.Address = *addr
		}
		if c.PollInterval < 0 {
			c.PollInterval = *poll
		}
		if c.ReportInterval < 0 {
			c.ReportInterval = *rep
		}
	case *serverConfig:
		storeInterv := flag.Int("i", m.DefaultStoreInterval, "Store interval arg: -i <sec>")
		filePath := flag.String("f", m.DefaultStorePath, "File path arg: -f </path/to/file>")
		rest := flag.Bool("r", m.DefaultRestore, "Restore storage arg: -r <true|false")
		flag.Parse()
		if c.Address == "none" {
			c.Address = *addr
		}
		if c.StoreInterval < 0 {
			c.StoreInterval = *storeInterv
		}
		if c.FileStoragePath == "none" {
			c.FileStoragePath = *filePath
		}
		if c.Restore {
			c.Restore = *rest
		}
	}

	return err
}

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

	// if cfg.FileStoragePath == "" {
	mem := server.NewMemStorage()
	manager.Store = &mem
	// } else {
	// 	fileStore, err := server.NewFileStorage(
	// 		cfg.StoreInterval,
	// 		cfg.FileStoragePath,
	// 		cfg.Restore,
	// 	)
	// 	if err != nil {
	// 		return nil, fmt.Errorf("set storage error: %w", err)
	// 	}
	// 	manager.Store = fileStore
	// }

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
