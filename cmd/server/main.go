package main

import (
	"net/http"

	"github.com/komogortzef/metrics/internal/routes"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	return http.ListenAndServe("localhost:8080", routes.SetRouter())
}
