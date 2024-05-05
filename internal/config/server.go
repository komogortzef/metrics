// config конфигурация с помощью аргументов cmd и маршрутизация сервера
package config

import (
	"flag"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/komogortzef/metrics/internal/handlers"
	"github.com/komogortzef/metrics/internal/storage"
)

var srv handlers.Server

// привязка флагов к полям Server
func init() {
	srv.Store = storage.MemStorage{}

	flag.StringVar(&srv.Endpoint, "a", "localhost:8080", "set Endpoint address: <host:port>")

	// пригодится когда будут добавлены новые экземпляры storage.Storage
	flag.Func("s", "set storage type: <mem> or <db>", func(flagVal string) error {
		log.Println("flag.Func starts performing...")
		srv.Store = storage.MemStorage{}

		return nil
	})
}

func Run() error {
	flag.Parse()
	log.Println("server address:", srv.Endpoint)
	log.Println("storage:", srv.Store)

	return http.ListenAndServe(srv.Endpoint, SetRouter())
}

func SetRouter() chi.Router {
	r := chi.NewRouter()

	r.Get("/", srv.ShowAll)
	r.Route("/", func(r chi.Router) {
		r.Get("/value/{tp}/{name}", srv.GetMetric)
		r.Post("/update/{tp}/{name}/{val}", srv.SaveToMem)
	})

	return r
}
