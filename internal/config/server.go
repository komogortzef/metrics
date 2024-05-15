package config

import (
	"log"
	"net/http"

	"metrics/internal/server"

	"github.com/go-chi/chi/v5"
)

func NewServer(opts ...Option) (*http.Server, error) {
	var err error
	var options options
	for _, opt := range opts {
		err = opt(&options)
	}

	server.SetStorage("mem")

	srv := &http.Server{
		Addr:    options.Address,
		Handler: getRoutes(),
	}
	log.Println("serve and listen:", options.Address)

	return srv, err
}

func getRoutes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", server.GetAllHandler)
	r.Get("/value/{kind}/{name}", server.GetHandler)
	r.Post("/update/{kind}/{name}/{val}", server.UpdateHandler)

	return r
}
