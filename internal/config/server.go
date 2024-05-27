package config

import (
	"fmt"
	"net/http"

	"metrics/internal/logger"
	"metrics/internal/server"

	"github.com/go-chi/chi/v5"
)

var router = chi.NewRouter()

func NewServer(opts ...Option) (*http.Server, error) {
	var err error
	err = logger.InitLog()
	if err != nil {
		return nil, fmt.Errorf("init logger error: %w", err)
	}

	var options options
	for _, opt := range opts {
		err = opt(&options)
	}

	server.SetStorage("mem")

	srv := &http.Server{
		Addr:    options.Address,
		Handler: router,
	}

	return srv, err
}

func init() {
	router.Use(logger.WithHandlerLog)
	router.Get("/", server.GetAllHandler)
	router.Post("/value/", server.GetJSON)
	router.Get("/value/{kind}/{name}", server.GetHandler)
	router.Post("/update/", server.UpdateJSON)
	router.Post("/update/{kind}/{name}/{val}", server.UpdateHandler)
}
