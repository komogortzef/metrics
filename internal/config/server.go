package config

import (
	"fmt"
	"net/http"

	"metrics/internal/compress"
	"metrics/internal/logger"
	"metrics/internal/server"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

var router = chi.NewRouter()

func NewServer(opts ...Option) (*http.Server, error) {
	err := logger.InitLog()
	if err != nil {
		return nil, fmt.Errorf("init logger error: %w", err)
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

	logger.Info("Serv config:",
		zap.String("addr", options.Address),
		zap.Int("store interval", options.storeInterval),
		zap.String("file stor path", options.fileStorage),
		zap.Bool("restore", options.restore),
	)

	return srv, err
}

func init() {
	router.Use(logger.WithHandlerLog)
	router.Get("/", compress.GzipMiddleware(server.GetAllHandler))
	router.Post("/value/", compress.GzipMiddleware(server.GetJSON))
	router.Get("/value/{type}/{id}", server.GetHandler)
	router.Post("/update/", compress.GzipMiddleware(server.UpdateJSON))
	router.Post("/update/{type}/{id}/{value}", server.UpdateHandler)
}
