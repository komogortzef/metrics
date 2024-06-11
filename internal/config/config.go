package config

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"metrics/internal/agent"
	"metrics/internal/compress"
	log "metrics/internal/logger"
	m "metrics/internal/models"
	"metrics/internal/server"

	"github.com/caarlos0/env/v11"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type (
	Configurable interface {
		Run() error
	}

	Option  func(Configurable) error
	Service uint8
)

func Configure(service Configurable, opts ...Option) (Configurable, error) {
	err := log.InitLog()
	if err != nil {
		return nil, err
	}
	for _, opt := range opts {
		if err = opt(service); err != nil {
			break
		}
	}

	return service, err
}

func WithEnvCmd(service Configurable) (err error) {
	err = env.Parse(service)
	if err != nil {
		return fmt.Errorf("env parse error: %w", err)
	}
	addr := flag.String("a", m.DefaultEndpoint, "Endpoint arg: -a <host:port>")
	switch c := service.(type) {
	case *agent.SelfMonitor:
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
		c.Mtx = &sync.RWMutex{}
	case *server.MetricManager:
		storeInterv := flag.Int("i", m.DefaultStoreInterval, "Store interval arg: -i <sec>")
		filePath := flag.String("f", m.DefaultStorePath, "File path arg: -f </path/to/file>")
		rest := flag.Bool("r", m.DefaultRestore, "Restore storage arg: -r <true|false>")
		dbAddr := flag.String("d", "", "DB address arg: -d <dbserver://username:password@host:port/db_name>")
		flag.Parse()
		if c.Address == "none" {
			c.Address = *addr
		}
		if c.StoreInterval < 0 {
			c.StoreInterval = *storeInterv
		}
		if filestore, ok := os.LookupEnv("FILE_STORAGE_PATH"); !ok {
			c.FileStoragePath = *filePath
		} else {
			c.FileStoragePath = filestore
		}
		if c.Restore {
			c.Restore = *rest
		}
		if c.DBAddress == "none" {
			c.DBAddress = *dbAddr
		}
	}
	return
}

func WithRoutes(service Configurable) (err error) {
	if manager, ok := service.(*server.MetricManager); ok {
		router := chi.NewRouter()
		router.Use(log.WithHandlerLog)
		router.Get("/", compress.GzipMiddleware(manager.GetAllHandler))
		router.Get("/ping", manager.PingHandler)
		router.Post("/value/", compress.GzipMiddleware(manager.GetJSON))
		router.Get("/value/{type}/{id}", manager.GetHandler)
		router.Post("/update/", compress.GzipMiddleware(manager.UpdateJSON))
		router.Post("/update/{type}/{id}/{value}", manager.UpdateHandler)

		manager.Serv = &http.Server{
			Addr:    manager.Address,
			Handler: router,
		}
	}
	return
}

func WithStorage(service Configurable) (err error) {
	if manager, ok := service.(*server.MetricManager); ok {
		if manager.DBAddress != "" {
			manager.Store = &server.DataBase{
				Pool: &pgxpool.Pool{},
			}
		} else if manager.FileStoragePath != "" {
			manager.Store = &server.FileStorage{
				MemStorage: server.MemStorage{
					Items: make(map[string][]byte, m.MetricsNumber),
					Mtx:   &sync.RWMutex{},
				},
				FilePath: manager.FileStoragePath,
				Interval: time.Duration(manager.StoreInterval),
			}
		} else {
			manager.Store = &server.MemStorage{
				Items: make(map[string][]byte, m.MetricsNumber),
				Mtx:   &sync.RWMutex{},
			}
		}
	}
	return
}
