package routes

import (
	"handlers"
	"net/http"
	"storage"
)

var Mux = http.NewServeMux()
var handler = handlers.NewHandler(storage.MemStorage{})

func init() {
	Mux.HandleFunc("/update/", handler.SaveToMem)
}
