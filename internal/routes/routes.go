package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/komogortzef/metrics/internal/handlers"
	"github.com/komogortzef/metrics/internal/storage"
)

// var (
// 	Mux     = http.NewServeMux()
// 	handler = handlers.NewHandler(storage.MemStorage{})
// )

// func init() {
// 	Mux.HandleFunc("/update/", handler.SaveToMem)
// }

func SetRouter(storeType string) chi.Router {
	r := chi.NewRouter()

	switch {
	default:
		h := handlers.NewHandler(storage.MemStorage{})
		r.Get("/", h.ShowAll)
		r.Route("/", func(r chi.Router) {
			r.Get("/value/{tp}/{name}", h.GetMetric)
			r.Post("/update/{tp}/{name}/{val}", h.SaveToMem)
		})
	}

	return r
}
