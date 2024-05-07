package server

import (
	"errors"
	"flag"
	"log"
	"os"
	"regexp"

	"github.com/caarlos0/env/v6"
	"github.com/go-chi/chi/v5"
	"github.com/komogortzef/metrics/internal/storage"
)

const maxArgs = 3

// Server - конфигурация сервера.(сокет и тип хранилища).
type ServerConf struct {
	Endpoint string `env:"ADDRESS"`
	Store    storage.Storage
}

func (srv *ServerConf) GetRoutes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", srv.ShowAll)
	r.Route("/", func(r chi.Router) {
		r.Get("/value/{tp}/{name}", srv.GetMetric)
		r.Post("/update/{tp}/{name}/{val}", srv.SaveToMem)
	})

	return r
}

func (srv *ServerConf) configure() error {
	if len(os.Args) > maxArgs {
		return errors.New("you must specify the endpoint: <./server -a host:port>")
	}

	err := env.Parse(srv)
	log.Println("error:", err)
	if err != nil || srv.Endpoint == "" {
		flag.StringVar(&srv.Endpoint, "a", "localhost:8080", "Endpoint address")
		flag.Parse()
	}

	isHostPort := regexp.MustCompile(`^(.*):(\d+)$`).MatchString
	if !isHostPort(srv.Endpoint) {
		return errors.New("the required format of the endpoint address: <host:port>")
	}

	srv.Store = storage.MemStorage{}

	return nil
}

func GetConfig() (ServerConf, error) {
	srv := ServerConf{}
	err := srv.configure()

	return srv, err
}
