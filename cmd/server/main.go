package main

import (
	"net/http"
	"routes"
)

func main() {

	err := http.ListenAndServe(":8080", routes.Mux)
	if err != nil {
		panic(err)
	}
}
