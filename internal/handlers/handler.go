package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/komogortzef/metrics/internal/storage"
)

type Handler struct {
	store storage.Storage
}

func NewHandler(store storage.Storage) *Handler {
	log.Println("handler creating...")
	return &Handler{
		store: store,
	}
}

func (h *Handler) SaveToMem(resp http.ResponseWriter, req *http.Request) {
	log.Println("SaveToMem handler")

	tp := []byte(chi.URLParam(req, "tp"))
	name := []byte(chi.URLParam(req, "name"))
	val := []byte(chi.URLParam(req, "val"))

	log.Println("saving data...")
	err := h.store.Save(tp, name, val)
	if err != nil {
		http.Error(resp, "Bad Request", http.StatusBadRequest)
	}
	log.Println("SaveToMem completed")
}

func (h *Handler) ShowAll(resp http.ResponseWriter, _ *http.Request) {
	log.Println("ShowAll handler")

	res := strings.Builder{}

	switch T := h.store.(type) {
	case storage.MemStorage:
		store, _ := h.store.(storage.MemStorage)
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

func (h *Handler) GetMetric(resp http.ResponseWriter, req *http.Request) {
	log.Println("GetMetric start..")

	name := chi.URLParam(req, "name")

	switch T := h.store.(type) {
	case storage.MemStorage:
		store, _ := h.store.(storage.MemStorage)
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
