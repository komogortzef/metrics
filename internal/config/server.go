package config

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	"metrics/internal/server"

	"github.com/go-chi/chi/v5"
)

const (
	maxArgsServer = 1

	basePath   = "/"
	getValPath = "/value/{kind}/{name}"
	updatePath = "/update/{kind}/{name}/{val}"
)

func NewServer(opts ...Option) (*http.Server, error) {
	var options options
	for _, opt := range opts {
		opt(&options)
	}

	// проверка на нужный набор аргументов для сервера
	parsedFlags := flag.NFlag()
	if parsedFlags > maxArgsServer || len(os.Args) > 3 || parsedFlags == 1 && os.Args[1] != "-a" {
		fmt.Fprintln(os.Stderr, "\nInvalid set of args:")
		usage()
		os.Exit(1)
	}
	if options.Address == "" {
		options.Address = defaultAddr
	}

	server.STORAGE = &server.MemStorage{
		Items: make(map[string][]byte),
		Mtx:   &sync.RWMutex{},
	}
	srv := &http.Server{
		Addr:    options.Address,
		Handler: getRoutes(),
	}
	log.Println("linstening on:", options.Address)

	return srv, nil
}

func getRoutes() chi.Router {
	r := chi.NewRouter()
	r.Get(basePath, server.GetAllHandler)
	r.Get(getValPath, server.GetHandler)
	r.Post(updatePath, server.UpdateHandler)

	return r
}
