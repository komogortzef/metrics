package config

import (
	ctx "context"
	"fmt"
	"net/http"

	c "metrics/internal/compress"
	log "metrics/internal/logger"
	sec "metrics/internal/security"
	"metrics/internal/server"

	"github.com/go-chi/chi/v5"
)

func setStorage(cx ctx.Context, cfg *config) (server.Storage, error) {
	switch {
	case cfg.DBAddress != "":
		db, err := server.NewDB(cx, cfg.DBAddress)
		if err != nil {
			return nil, fmt.Errorf("db configure error: %w", err)
		}
		return db, nil
	case cfg.FileStoragePath != "":
		fs := server.NewFileStore(cfg.FileStoragePath, cfg.StoreInterval)
		if cfg.Restore {
			fs.RestoreFromFile(cx)
		}
		return fs, nil
	}
	return server.NewMemStore(), nil
}

func getRoutes(cx ctx.Context, m *server.MetricManager, cfg *config) *chi.Mux {
	ctxMiddleware := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			r = r.WithContext(cx)
			next.ServeHTTP(w, r)
		}
	}
	router := chi.NewRouter()
	router.Use(log.WithHandlerLog)
	router.Use(c.GzipMiddleware)
	router.Get("/", ctxMiddleware(m.GetAllHandler))
	router.Get("/ping", ctxMiddleware(m.PingHandler))
	router.Post("/value/", ctxMiddleware(sec.HashMiddleware(cfg.Key, m.GetJSON)))
	router.Get("/value/{type}/{id}", ctxMiddleware(m.GetHandler))
	router.Post("/update/", ctxMiddleware(sec.HashMiddleware(cfg.Key, m.UpdateJSON)))
	router.Post("/update/{type}/{id}/{value}", ctxMiddleware(m.UpdateHandler))
	router.Post("/updates/", ctxMiddleware(sec.HashMiddleware(cfg.Key, m.BatchHandler)))

	return router
}
