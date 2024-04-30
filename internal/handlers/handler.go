package handlers

import (
	"net/http"
	"storage"
	"strings"
)

type Handler struct {
	store storage.Storage
}

func NewHandler(store storage.Storage) *Handler {
	return &Handler{
		store: store,
	}
}

func (h *Handler) SaveToMem(resp http.ResponseWriter, req *http.Request) {

	if req.Method != http.MethodPost {
		http.Error(resp, "Method not allowed", http.StatusMethodNotAllowed)
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

	err := h.store.Save(tp, name, val)
	if err != nil {
		http.Error(resp, "Bad Request", http.StatusBadRequest)
	}
}
