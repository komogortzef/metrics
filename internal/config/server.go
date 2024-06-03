package config

import (
	"fmt"
	"net/http"

	c "metrics/internal/compress"
	l "metrics/internal/logger"
	"metrics/internal/server"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func newServer(options *options) (*server.MetricsManager, error) {
	var manager server.MetricsManager

	if options.fileStorage == "" {
		mem := server.NewMemStorage()
		manager.Storage = &mem
	} else {
		fileStore, err := server.NewFileStorage(
			options.storeInterval,
			options.fileStorage,
			options.restore,
		)
		if err != nil {
			return nil, fmt.Errorf("set storage error: %w", err)
		}
		manager.Storage = fileStore
	}

	router := chi.NewRouter()
	router.Use(l.WithHandlerLog)
	router.Get("/", c.GzipMiddleware(manager.GetAllHandler))
	router.Post("/value/", c.GzipMiddleware(manager.GetJSON))
	router.Get("/value/{type}/{id}", manager.GetHandler)
	router.Post("/update/", c.GzipMiddleware(manager.UpdateJSON))
	router.Post("/update/{type}/{id}/{value}", manager.UpdateHandler)

	manager.Serv = &http.Server{
		Addr:    options.Address,
		Handler: router,
	}
	l.Info("Serv config:",
		zap.String("addr", options.Address),
		zap.Int("store interval", options.storeInterval),
		zap.String("file stor path", options.fileStorage),
		zap.Bool("restore", options.restore),
	)

	return &manager, nil
}
