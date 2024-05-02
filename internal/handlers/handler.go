package handlers

import (
	"log"
	"net/http"
	"strings"

	"github.com/komogortzef/metrics/internal/storage"
)

type Handler struct {
	store storage.Storage
}

func NewHandler(store storage.Storage) *Handler {
	log.Println("\nhandler creating...")
	return &Handler{
		store: store,
	}
}

func (h *Handler) SaveToMem(resp http.ResponseWriter, req *http.Request) {
	log.Println("\nstart of request processing...")
	if req.Method != http.MethodPost {
		log.Println("Method is not allowed")
		http.Error(resp, "Method is not allowed", http.StatusMethodNotAllowed)
		return
	}

	uri := req.URL.Path[1:]
	reqElem := strings.Split(uri, "/")
	if len(reqElem) != 4 {
		http.Error(resp, "Not Found", http.StatusNotFound)
		return
	}

	tp := []byte(reqElem[1])
	name := []byte(reqElem[2])
	val := []byte(reqElem[3])

	log.Println("saving data...")
	err := h.store.Save(tp, name, val)
	if err != nil {
		http.Error(resp, "Bad Request", http.StatusBadRequest)
	}
}
