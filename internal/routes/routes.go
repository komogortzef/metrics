package routes

import (
	"handlers"
	"net/http"
)

var Mux = http.NewServeMux()

func init() {
	Mux.HandleFunc("/update/", handlers.Update)
}
