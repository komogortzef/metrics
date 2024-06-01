package server

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	"metrics/internal/logger"
	m "metrics/internal/models"

	"github.com/go-chi/chi/v5"
	"github.com/tidwall/gjson"
	"go.uber.org/zap"
)

type Repository interface {
	io.Reader
	io.Writer
	Put(metName string, data []byte) error
	Get(metname string) ([]byte, error)
}

var storage Repository

func UpdateHandler(rw http.ResponseWriter, req *http.Request) {
	logger.Debug("UPDATE HANDLER starts ...")
	kind := chi.URLParam(req, m.Mtype)
	name := chi.URLParam(req, m.Id)
	val := chi.URLParam(req, m.Value)

	metric, err := m.NewMetric(name, kind, val)
	if err != nil {
		logger.Warn("Invalid metric value or type")
		http.Error(rw, m.BadRequestMessage, http.StatusBadRequest)
		return
	}

	bytes, err := metric.MarshalJSON()
	if err != nil {
		logger.Warn("Marshal error", zap.Error(err))
	}

	if _, err = storage.Write(bytes); err != nil {
		logger.Fatal("Put to storage error", zap.Error(err))
	}

	rw.WriteHeader(http.StatusOK)
}

func GetHandler(rw http.ResponseWriter, req *http.Request) {
	logger.Debug("GET HANDLER starts ...")
	kind := chi.URLParam(req, m.Mtype)
	name := chi.URLParam(req, m.Id)

	if kind != m.Gauge && kind != m.Counter {
		logger.Warn("Invalid metric type")
		http.Error(rw, m.BadRequestMessage, http.StatusBadRequest)
		return
	}

	bytes, err := storage.Get(name)
	if err != nil {
		logger.Warn("Coundn't fetch the metric from storage")
		http.Error(rw, m.NotFoundMessage, http.StatusNotFound)
		return
	}

	var numStr string
	if kind == m.Counter {
		numBytes := gjson.GetBytes(bytes, m.Delta)
		numStr = strconv.FormatInt(numBytes.Int(), 10)
	} else {
		numBytes := gjson.GetBytes(bytes, m.Value)
		numStr = strconv.FormatFloat(numBytes.Float(), 'f', -1, 64)
	}

	rw.WriteHeader(http.StatusOK)
	_, _ = rw.Write([]byte(numStr))
}

func GetAllHandler(rw http.ResponseWriter, req *http.Request) {
	logger.Debug("GET ALL HANDLER starts ...")
	list := make([]Item, 0, metricsNumber)

	for _, metric := range getList(storage) {
		fmt.Println(metric.String())
		list = append(list, Item{Met: metric.String()})
	}

	html, err := renderGetAll(list)
	if err != nil {
		logger.Warn("An error occured during html rendering")
		http.Error(rw, m.InternalErrorMsg, http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "text/html")
	rw.WriteHeader(http.StatusOK)
	_, _ = rw.Write(html.Bytes())
}

func UpdateJSON(rw http.ResponseWriter, req *http.Request) {
	logger.Info("UpdateJSON starts...")
	bytes, err := io.ReadAll(req.Body)
	if err != nil {
		logger.Warn("Couldn't read with decompress")
	}
	defer req.Body.Close()

	typeBytes := gjson.GetBytes(bytes, m.Mtype)
	nameBytes := gjson.GetBytes(bytes, m.Id)
	if err = storage.Put(nameBytes.String(), bytes); err != nil {
		logger.Warn("Coudn't save data to storage")
	}
	if typeBytes.String() == m.Counter {
		bytes = bytes[:0]
		bytes, err = storage.Get(nameBytes.String())
		if err != nil {
			logger.Warn("Coulnd't fetch the metric from storage")
		}
	}

	rw.WriteHeader(http.StatusOK)
	_, _ = rw.Write(bytes)
}

func GetJSON(rw http.ResponseWriter, req *http.Request) {
	logger.Info("GetJSON starts...")
	bytes, err := io.ReadAll(req.Body)
	if err != nil {
		logger.Warn("Couldn't read with decompress")
	}
	defer req.Body.Close()

	nameBytes := gjson.GetBytes(bytes, m.Id)
	jsonBytes, err := storage.Get(nameBytes.String())
	if err != nil {
		http.Error(rw, m.NotFoundMessage, http.StatusNotFound)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	_, _ = rw.Write(jsonBytes)
}
