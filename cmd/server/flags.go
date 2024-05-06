package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/komogortzef/metrics/internal/handlers"
	"github.com/komogortzef/metrics/internal/storage"
)

var srv handlers.Server

func run() error {
	srv.Store = storage.MemStorage{}
	flag.StringVar(&srv.Endpoint, "a", "localhost:8080", "set Endpoint address: <host:port>")
	flag.Parse()
	log.Println("server address:", srv.Endpoint)
	log.Println("storage:", srv.Store)

	return http.ListenAndServe(srv.Endpoint, SetRouter())
}

func SetRouter() chi.Router {
	r := chi.NewRouter()

	r.Get("/", srv.ShowAll)
	r.Route("/", func(r chi.Router) {
		r.Get("/value/{tp}/{name}", srv.GetMetric)
		r.Post("/update/{tp}/{name}/{val}", srv.SaveToMem)
	})

	return r
}
