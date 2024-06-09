package conf

import (
	"fmt"
	"net/http"
	"sync"

	"metrics/internal/agent"
	c "metrics/internal/compress"
	l "metrics/internal/logger"
	"metrics/internal/server"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type (
	Configurable interface {
		Run() error
	}

	config interface {
		setConfig() (Configurable, error)
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
		DBAddress       string `env:"DATABASE_DSN" envDefault:"none"`
	}

	ServiceType uint8
	Option      func(config) error
)

const (
	MetricServer ServiceType = iota
	MetricAgent
)

func (cfg *serverConfig) setConfig() (Configurable, error) {
	var manager server.MetricsManager

	router := chi.NewRouter()
	router.Use(l.WithHandlerLog)
	router.Get("/", c.GzipMiddleware(manager.GetAllHandler))
	router.Get("/ping", manager.PingHandler)
	router.Post("/value/", c.GzipMiddleware(manager.GetJSON))
	router.Get("/value/{type}/{id}", manager.GetHandler)
	router.Post("/update/", c.GzipMiddleware(manager.UpdateJSON))
	router.Post("/update/{type}/{id}/{value}", manager.UpdateHandler)

	manager.Serv = &http.Server{
		Addr:    cfg.Address,
		Handler: router,
	}

	cfg.setStorage(&manager)

	l.Info("Serv config:",
		zap.String("addr", cfg.Address),
		zap.Int("store interval", cfg.StoreInterval),
		zap.String("file stor path", cfg.FileStoragePath),
		zap.Bool("restore", cfg.Restore),
		zap.String("DB", cfg.DBAddress),
	)

	return &manager, nil
}

func (cfg *agentConfig) setConfig() (Configurable, error) {
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

// установка параметров конфига в зависимости от типа конфигурируемого объекта
// и инициализация логера
func Configure(service ServiceType, opts ...Option) (Configurable, error) {
	err := l.InitLog()
	if err != nil {
		return nil, fmt.Errorf("init logger error: %w", err)
	}

	var cfg config
	switch service {
	case MetricServer:
		cfg = &serverConfig{}
	case MetricAgent:
		cfg = &agentConfig{}
	}

	// установка зн-ий для полей структуры представляющей конфигурацию
	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			return nil, fmt.Errorf("error while options applying: %w", err)
		}
	}

	l.Info(service.String())
	return cfg.setConfig()
}

func (service ServiceType) String() string {
	switch service {
	case MetricServer:
		return "Metric Server"
	default:
		return "Metric Agent"
	}
}
