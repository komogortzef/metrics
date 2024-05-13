package config

import (
	"net/http"
	"sync"

	"metrics/internal/server"

	"github.com/go-chi/chi/v5"
)

const metricsNumber = 29

func NewServer(opts ...Option) (*http.Server, error) {
	var options options
	for _, opt := range opts {
		opt(&options)
	}

	server.STORAGE = &server.MemStorage{
		Items: make(map[string][]byte, metricsNumber),
		Mtx:   &sync.RWMutex{},
	}
	srv := &http.Server{
		Addr:    options.Address,
		Handler: getRoutes(),
	}

	return srv, nil
}

func getRoutes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", server.GetAllHandler)
	r.Get("/value/{kind}/{name}", server.GetHandler)
	r.Post("/update/{kind}/{name}/{val}", server.UpdateHandler)

	return r
}
