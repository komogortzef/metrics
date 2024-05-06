// handlers тип сервера и привязанные к нему обработчики
package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/komogortzef/metrics/internal/storage"
)

// Server - конфигурация сервера.(сокет и тип хранилища).
type Config struct {
	Endpoint string
	Store    storage.Storage
}

// SaveToMem сохранить в память(обработчик метода POST).
func (h *Config) SaveToMem(resp http.ResponseWriter, req *http.Request) {
	log.Println("SaveToMem handler")

	tp := chi.URLParam(req, "tp")
	name := chi.URLParam(req, "name")
	val := chi.URLParam(req, "val")

	log.Println("saving data...")
	err := h.Store.Save(tp, name, val)
	if err != nil {
		http.Error(resp, "Bad Request", http.StatusBadRequest)
	}
	log.Println("SaveToMem completed")
}

// ShowAll обработчик метода GET.
func (h *Config) ShowAll(resp http.ResponseWriter, _ *http.Request) {
	log.Println("ShowAll handler")

	res := strings.Builder{}

	// в зависимости от типа хранилища выбираем логику извлечения данных
	switch T := h.Store.(type) {
	case storage.MemStorage:
		store, _ := h.Store.(storage.MemStorage)
		for name, val := range store {
			str := fmt.Sprintf("%s: %v\n", name, val)
			res.WriteString(str)
		}
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
func (h *Config) GetMetric(resp http.ResponseWriter, req *http.Request) {
	log.Println("GetMetric start..")

	name := chi.URLParam(req, "name")

	// в зависимости от типа хранилища выбираем логику извлечения данных
	switch T := h.Store.(type) {
	case storage.MemStorage:
		store, _ := h.Store.(storage.MemStorage)
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
