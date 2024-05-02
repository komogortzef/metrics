package routes

import (
	"net/http"

	"github.com/komogortzef/metrics/internal/handlers"
	"github.com/komogortzef/metrics/internal/storage"
)

var (
	Mux     = http.NewServeMux()
	handler = handlers.NewHandler(storage.MemStorage{})
)

func init() {
	Mux.HandleFunc("/update/", handler.SaveToMem)
}
