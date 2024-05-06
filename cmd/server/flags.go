package main

import (
	"errors"
	"flag"
	"net/http"
	"os"
	"regexp"

	"github.com/go-chi/chi/v5"
	"github.com/komogortzef/metrics/internal/handlers"
	"github.com/komogortzef/metrics/internal/storage"
)

var isHostPort = regexp.MustCompile(`^(.*):(\d+)$`)

func run() error {

	if len(os.Args) > 3 {
		return errors.New("you must specify the endpoint: <./server -a host:port>")
	}

	var srv handlers.Config
	srv.Store = storage.MemStorage{}

	flag.StringVar(&srv.Endpoint, "a", "localhost:8080", "set Endpoint address: <host:port>")
	flag.Parse()

	if !isHostPort.MatchString(srv.Endpoint) {
		return errors.New("the required format of the endpoint address: <host:port>")
	}

	return http.ListenAndServe(srv.Endpoint, SetRouter(&srv))
}

func SetRouter(srv *handlers.Config) chi.Router {
	r := chi.NewRouter()

	r.Get("/", srv.ShowAll)
	r.Route("/", func(r chi.Router) {
		r.Get("/value/{tp}/{name}", srv.GetMetric)
		r.Post("/update/{tp}/{name}/{val}", srv.SaveToMem)
	})

	return r
}
