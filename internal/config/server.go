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

var router = chi.NewRouter()

func NewServer(opts ...func(*options)) (*http.Server, error) {
	err := l.InitLog()
	if err != nil {
		return nil, fmt.Errorf("init l error: %w", err)
	}
	var options options
	for _, opt := range opts {
		opt(&options)
	}
	server.SetStorage("mem")
	srv := &http.Server{
		Addr:    options.Address,
		Handler: router,
	}
	l.Info("Serv config:",
		zap.String("addr", options.Address),
		zap.Int("store interval", options.storeInterval),
		zap.String("file stor path", options.fileStorage),
		zap.Bool("restore", options.restore),
	)

	return srv, nil
}

func init() {
	router.Use(l.WithHandlerLog)
	router.Get("/", c.GzipMiddleware(server.GetAllHandler))
	router.Post("/value/", c.GzipMiddleware(server.GetJSON))
	router.Get("/value/{type}/{id}", server.GetHandler)
	router.Post("/update/", c.GzipMiddleware(server.UpdateJSON))
	router.Post("/update/{type}/{id}/{value}", server.UpdateHandler)
}
