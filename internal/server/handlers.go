// server тип сервера и привязанные к нему обработчики
package server

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/komogortzef/metrics/internal/storage"
)

// SaveToMem сохранить в память(обработчик метода POST).
func (srv *ServerConf) SaveToMem(resp http.ResponseWriter, req *http.Request) {
	log.Println("SaveToMem handler")

	tp := chi.URLParam(req, "tp")
	name := chi.URLParam(req, "name")
	val := chi.URLParam(req, "val")

	log.Println("saving data...")
	err := srv.Store.Save(tp, name, val)
	if err != nil {
		http.Error(resp, "Bad Request", http.StatusBadRequest)
	}
	log.Println("SaveToMem completed")
}

// ShowAll обработчик метода GET.
func (srv *ServerConf) ShowAll(resp http.ResponseWriter, _ *http.Request) {
	log.Println("ShowAll handler")

	res := strings.Builder{}

	// в зависимости от типа хранилища выбираем логику извлечения данных
	switch T := srv.Store.(type) {
	case storage.MemStorage:
		store, _ := srv.Store.(storage.MemStorage)
		for name, val := range store {
			str := fmt.Sprintf("%s: %v\n", name, val)
			res.WriteString(str)
		}

		resp.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, err := resp.Write([]byte(res.String()))
		if err != nil {
			log.Println(err)
		}
		log.Println("ShowAll completed")
	default:
		log.Println("Unknown type:", T)
		resp.WriteHeader(http.StatusNotFound)
		return
	}
}

// GetMetric обработчик метода GET.
func (srv *ServerConf) GetMetric(resp http.ResponseWriter, req *http.Request) {
	log.Println("GetMetric start..")

	name := chi.URLParam(req, "name")

	// в зависимости от типа хранилища выбираем логику извлечения данных
	switch T := srv.Store.(type) {
	case storage.MemStorage:
		store, _ := srv.Store.(storage.MemStorage)
		val, ok := store[name]
		if !ok {
			resp.WriteHeader(http.StatusNotFound)
			return
		}
		res := fmt.Sprintf("%v", val)
		_, err := resp.Write([]byte(res))
		if err != nil {
			log.Println(err)
		}
	default:
		log.Println("Unknown type:", T)
		resp.WriteHeader(http.StatusNotFound)
		return
	}
}
