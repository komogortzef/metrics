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

func (srv *ServerConf) Configure() error {
	if len(os.Args) > maxArgs {
		return errors.New("you must specify the endpoint: <./server -a host:port>")
	}

	err := env.Parse(srv)
	if err != nil {
		log.Println("use cmd args")
		flag.StringVar(&srv.Endpoint, "a", "localhost:8080", "set Endpoint address: <host:port>")
		flag.Parse()
	}

	isHostPort := regexp.MustCompile(`^(.*):(\d+)$`).MatchString
	if !isHostPort(srv.Endpoint) {
		return errors.New("the required format of the endpoint address: <host:port>")
	}

	srv.Store = storage.MemStorage{}

	return nil
}

func (srv *ServerConf) ShowConfig() {
	log.Printf("Server configuration:\nEndpoint: %s", srv.Endpoint)
}

func GetConfig() (ServerConf, error) {
	srv := ServerConf{}
	err := srv.Configure()

	return srv, err
}
