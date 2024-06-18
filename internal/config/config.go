package config

import (
	ctx "context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"metrics/internal/agent"
	c "metrics/internal/compress"
	log "metrics/internal/logger"
	m "metrics/internal/models"
	"metrics/internal/server"

	"github.com/caarlos0/env/v11"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

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

func WithEnvCmd(_ ctx.Context, cfg Config) (err error) {
	err = env.Parse(cfg)
	if err != nil {
		return fmt.Errorf("env parse error: %w", err)
	}
	addr := flag.String("a", m.DefaultEndpoint, "Endpoint arg: -a <host:port>")
	switch c := cfg.(type) {
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

func WithRoutes(ctx ctx.Context, cfg Config) (_ error) {
	if manager, ok := cfg.(*server.MetricManager); ok {
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
	}
	return
}

func WithStorage(ctx ctx.Context, cfg Config) (err error) {
	if manager, ok := cfg.(*server.MetricManager); ok {
		if manager.DBAddress != "" {
			if manager.Store, err = server.NewDB(ctx, manager.DBAddress); err != nil {
				return err
			}
			manager.FileStoragePath = m.NoStorage
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
	}
	return err
}
