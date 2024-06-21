package config

import (
	ctx "context"
	"flag"
	"fmt"
	"net/http"
	"os"

	"metrics/internal/agent"
	c "metrics/internal/compress"
	log "metrics/internal/logger"
	"metrics/internal/server"
	s "metrics/internal/service"

	"github.com/caarlos0/env/v11"
	"github.com/go-chi/chi/v5"
)

func EnvFlagsAgent(_ ctx.Context, cfg Config) (err error) {
	ag, ok := cfg.(*agent.SelfMonitor)
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
	if ag.Address == "none" {
		ag.Address = *addr
	}
	if ag.PollInterval < 0 {
		ag.PollInterval = *poll
	}
	if ag.ReportInterval < 0 {
		ag.ReportInterval = *rep
	}
	return
}

func EnvFlagsServer(_ ctx.Context, cfg Config) (err error) {
	serv, ok := cfg.(*server.MetricManager)
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
	if serv.Address == "none" {
		serv.Address = *addr
	}
	if serv.StoreInterval < 0 {
		serv.StoreInterval = *storeInterv
	}
	if filestore, ok := os.LookupEnv("FILE_STORAGE_PATH"); !ok {
		serv.FileStoragePath = *filePath
	} else {
		serv.FileStoragePath = filestore
	}
	if serv.Restore {
		serv.Restore = *rest
	}
	if serv.DBAddress == "none" {
		serv.DBAddress = *dbAddr
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
	return server.NewStorage(ctx, manager)
}
