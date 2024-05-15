package server

import (
	"log"
	"net/http"
	"regexp"
	"sync"

	"github.com/go-chi/chi/v5"
)

const (
	internalServerErrorMessage = "internal server error"
	notFoundMessage            = "not found"
	badRequestMessage          = "bad request"
)

type Operation func([]byte, []byte) ([]byte, error)

type Repository interface {
	Save(key string, value []byte, opers ...Operation) error
	Get(key string) ([]byte, bool)
	GetAll() map[string][]byte
}

var storage Repository

// в будущем пригодится для конфигурирования хранилища
func SetStorage(st string) {
	switch st {
	default:
		storage = &MemStorage{
			Items: make(map[string][]byte, metricsNumber),
			Mtx:   &sync.RWMutex{},
		}
	}
}

func UpdateHandler(rw http.ResponseWriter, req *http.Request) {
	name := chi.URLParam(req, "name")
	val := chi.URLParam(req, "val")

	switch chi.URLParam(req, "kind") {
	case "gauge":
		isReal := regexp.MustCompile(`^-?\d*\.?\d+$`).MatchString
		if !isReal(val) {
			http.Error(rw, badRequestMessage, http.StatusBadRequest)
			return
		}
		if err := storage.Save(name, []byte(val)); err != nil {
			http.Error(rw, internalServerErrorMessage, http.StatusInternalServerError)
			return
		}
		log.Println(name, ":", val, ".", "The value is received")
	case "counter":
		isNatural := regexp.MustCompile(`^\d+$`).MatchString
		if !isNatural(val) {
			http.Error(rw, badRequestMessage, http.StatusBadRequest)
			return
		}
		if err := storage.Save(name, []byte(val), WithAccInt64); err != nil {
			http.Error(rw, badRequestMessage, http.StatusBadRequest)
			return
		}
		log.Println(name, ":", val, "\t", "the value is received")
	default:
		http.Error(rw, badRequestMessage, http.StatusBadRequest)
	}
}

func GetHandler(rw http.ResponseWriter, req *http.Request) {
	name := chi.URLParam(req, "name")

	switch chi.URLParam(req, "kind") {
	case "counter":
		data, ok := storage.Get(name)
		if !ok {
			http.Error(rw, notFoundMessage, http.StatusNotFound)
			return
		}
		if bytes, err := rw.Write(data); err != nil {
			log.Printf("failed to send data. size: %v\n", bytes)
		}
	case "gauge":
		data, ok := storage.Get(name)
		if !ok {
			http.Error(rw, notFoundMessage, http.StatusNotFound)
			return
		}
		if bytes, err := rw.Write(data); err != nil {
			log.Printf("failed to send data. size: %v\n", bytes)
		}
	default:
		http.Error(rw, notFoundMessage, http.StatusNotFound)
	}
}

func GetAllHandler(wr http.ResponseWriter, req *http.Request) {
	list := make([]Item, 0, metricsNumber)

	log.Println("GETALL HANDLER")
	for name, value := range storage.GetAll() {
		list = append(list, Item{Name: name, Value: string(value)})
	}

	html, err := renderGetAll(list)
	if err != nil {
		http.Error(wr, internalServerErrorMessage, http.StatusInternalServerError)
		return
	}

	wr.Header().Set("Content-Type", "text/html")
	if _, err := html.WriteTo(wr); err != nil {
		http.Error(wr, internalServerErrorMessage, http.StatusInternalServerError)
	}
}
