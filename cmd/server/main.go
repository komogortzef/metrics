package main

import (
	"net/http"

	"github.com/komogortzef/metrics/internal/server"
)

func main() {

	srv, err := server.GetConfig()
	if err != nil {
		panic(err)
	}

	srv.ShowConfig()
	if err := http.ListenAndServe(srv.Endpoint, srv.GetRoutes()); err != nil {
		panic(err)
	}
}
