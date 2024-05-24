package config

import (
	"fmt"
	"net/http"

	"metrics/internal/logger"
	"metrics/internal/server"

	"github.com/go-chi/chi/v5"
)

func NewServer(opts ...Option) (*http.Server, error) {
	var err error
	err = logger.InitLog()
	if err != nil {
		return nil, fmt.Errorf("init logger error: %w", err)
	}

	logger.Info("Applying configuration options...")
	var options options
	for _, opt := range opts {
		err = opt(&options)
	}

	server.SetStorage("mem")

	srv := &http.Server{
		Addr:    options.Address,
		Handler: getRoutes(),
	}

	return srv, err
}

func getRoutes() chi.Router {
	logger.Info("Defining routes...")
	r := chi.NewRouter()

	r.Use(logger.WithHandlerLog)

	r.Get("/", server.GetAllHandler)
	r.Post("/value", server.GetJSON)
	r.Get("/value/{kind}/{name}", server.GetHandler)
	r.Post("/update", server.UpdateJSON)
	r.Post("/update/{kind}/{name}/{val}", server.UpdateHandler)
	logger.Info("The routes are defined")

	return r
}
