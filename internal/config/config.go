package config

import (
	ctx "context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"metrics/internal/agent"
	c "metrics/internal/compress"
	log "metrics/internal/logger"
	"metrics/internal/server"
	s "metrics/internal/service"

	"github.com/caarlos0/env/v11"
	"github.com/go-chi/chi/v5"
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
		return nil, err
	}
	for _, opt := range opts {
		if err = opt(ctx, cfg); err != nil {
			break
		}
	}
	return cfg, err
}

func CompletionCtx() (ctx.Context, ctx.CancelFunc) {
	ctx, complete := ctx.WithCancel(ctx.Background())
	signChan := make(chan os.Signal, 1)
	signal.Notify(signChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-signChan
		complete()
	}()
	return ctx, complete
}

func EnvFlagsAgent(_ ctx.Context, cfg Config) (err error) {
	agent, ok := cfg.(*agent.SelfMonitor)
	if !ok {
		return ErrInvalidConfig
	}
	err = env.Parse(cfg)
	if err != nil {
		return fmt.Errorf("env parse error: %w", err)
	}
	addr := flag.String("a", s.DefaultEndpoint, "Endpoint arg: -a <host:port>")
	poll := flag.Int("p", s.DefaultPollInterval, "Poll Interval arg: -p <sec>")
	rep := flag.Int("r", s.DefaultReportInterval, "Report interval arg: -r <sec>")
	flag.Parse()
	if agent.Address == "none" {
		agent.Address = *addr
	}
	if agent.PollInterval < 0 {
		agent.PollInterval = *poll
	}
	if agent.ReportInterval < 0 {
		agent.ReportInterval = *rep
	}
	return
}

func EnvFlagsServer(_ ctx.Context, cfg Config) (err error) {
	server, ok := cfg.(*server.MetricManager)
	if !ok {
		return ErrInvalidConfig
	}
	err = env.Parse(cfg)
	if err != nil {
		return fmt.Errorf("env parse error: %w", err)
	}
	addr := flag.String("a", s.DefaultEndpoint, "Endpoint arg: -a <host:port>")
	storeInterv := flag.Int("i", s.DefaultStoreInterval, "Store interval arg: -i <sec>")
	filePath := flag.String("f", s.DefaultStorePath, "File path arg: -f </path/to/file>")
	rest := flag.Bool("r", s.DefaultRestore, "Restore storage arg: -r <true|false>")
	dbAddr := flag.String("d", s.NoStorage, "DB address arg: -d <dbserver://username:password@host:port/db_name>")
	flag.Parse()
	if server.Address == "none" {
		server.Address = *addr
	}
	if server.StoreInterval < 0 {
		server.StoreInterval = *storeInterv
	}
	if filestore, ok := os.LookupEnv("FILE_STORAGE_PATH"); !ok {
		server.FileStoragePath = *filePath
	} else {
		server.FileStoragePath = filestore
	}
	if server.Restore {
		server.Restore = *rest
	}
	if server.DBAddress == "none" {
		server.DBAddress = *dbAddr
	}
	return
}

func WithRoutes(ctx ctx.Context, cfg Config) (_ error) {
	manager, ok := cfg.(*server.MetricManager)
	if !ok {
		return ErrInvalidConfig
	}
	ctxMiddleware := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		}
	}
	router := chi.NewRouter()
	router.Use(log.WithHandlerLog)
	router.Get("/", ctxMiddleware(c.GzipMiddleware(manager.GetAllHandler)))
	router.Get("/ping", ctxMiddleware(manager.PingHandler))
	router.Post("/value/", ctxMiddleware(c.GzipMiddleware(manager.GetJSON)))
	router.Get("/value/{type}/{id}", ctxMiddleware(manager.GetHandler))
	router.Post("/update/", ctxMiddleware(c.GzipMiddleware(manager.UpdateJSON)))
	router.Post("/update/{type}/{id}/{value}", ctxMiddleware(manager.UpdateHandler))
	router.Post("/updates/", ctxMiddleware(c.GzipMiddleware(manager.BatchHandler)))

	manager.Serv = &http.Server{
		Addr:    manager.Address,
		Handler: router,
	}
	return
}

func WithStorage(ctx ctx.Context, cfg Config) (err error) {
	manager, ok := cfg.(*server.MetricManager)
	if !ok {
		return ErrInvalidConfig
	}
	if manager.DBAddress != "" {
		if manager.Store, err = server.NewDB(ctx, manager.DBAddress); err != nil {
			return err
		}
		manager.FileStoragePath = s.NoStorage
	} else if manager.FileStoragePath != "" {
		store := server.NewFileStore(manager.FileStoragePath)
		if manager.StoreInterval > 0 {
			store.SyncDump = false
		}
		if manager.Restore {
			if err := store.RestoreFromFile(ctx); err != nil {
				log.Warn("restore from file error", zap.Error(err))
			}
		}
		manager.Store = store
	} else {
		manager.Store = server.NewMemStore()
	}
	return
}
