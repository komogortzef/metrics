package config

import (
	ctx "context"
	"fmt"
	"net/http"

	c "metrics/internal/compress"
	log "metrics/internal/logger"
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
