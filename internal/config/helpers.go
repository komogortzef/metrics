package config

import (
	ctx "context"
	"fmt"
	"net/http"

	c "metrics/internal/compress"
	log "metrics/internal/logger"
	"metrics/internal/server"
	s "metrics/internal/service"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func newStorage(cx ctx.Context, cfg *config) (store server.Storage, err error) {
	switch {
	case cfg.DBAddress != "":
		if store, err = server.NewDB(cx, cfg.DBAddress); err != nil {
			return nil, fmt.Errorf("db configure error: %w", err)
		}
		cfg.FileStoragePath = s.NoStorage
	case cfg.FileStoragePath != "":
		fsStore := server.NewFileStore(cfg.FileStoragePath, cfg.StoreInterval)
		if cfg.Restore {
			if err := fsStore.RestoreFromFile(cx); err != nil {
				log.Warn("restore from file error", zap.Error(err))
			}
		}
		if !fsStore.SyncDump {
			fsStore.DumpWait(cx, cfg.StoreInterval)
		}
		store = fsStore
	default:
		store = server.NewMemStore()
	}
	return store, nil
}

func getRoutes(cx ctx.Context, m *server.MetricManager) *chi.Mux {
	ctxMiddleware := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			r = r.WithContext(cx)
			next.ServeHTTP(w, r)
		}
	}
	router := chi.NewRouter()
	router.Use(log.WithHandlerLog)
	router.Get("/", ctxMiddleware(c.GzipMiddleware(m.GetAllHandler)))
	router.Get("/ping", ctxMiddleware(m.PingHandler))
	router.Post("/value/", ctxMiddleware(c.GzipMiddleware(m.GetJSON)))
	router.Get("/value/{type}/{id}", ctxMiddleware(m.GetHandler))
	router.Post("/update/", ctxMiddleware(c.GzipMiddleware(m.UpdateJSON)))
	router.Post("/update/{type}/{id}/{value}", ctxMiddleware(m.UpdateHandler))
	router.Post("/updates/", ctxMiddleware(c.GzipMiddleware(m.BatchHandler)))

	return router
}
