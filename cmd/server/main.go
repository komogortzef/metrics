package main

import (
	"net/http"
	"routes"
)

func main() {

	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	return http.ListenAndServe("localhost:8080", routes.Mux)
}
