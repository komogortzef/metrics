package server

import (
	"errors"
	"io"
	"net/http"
	"strconv"

	l "metrics/internal/logger"
	m "metrics/internal/models"

	"github.com/go-chi/chi/v5"
	"github.com/tidwall/gjson"
	"go.uber.org/zap"
)

type Repository interface {
	io.Writer
	Read([]byte) ([]byte, error)
}

var storage Repository

func UpdateHandler(rw http.ResponseWriter, req *http.Request) {
	l.Debug("UPDATE HANDLER starts ...")
	kind := chi.URLParam(req, m.Mtype)
	name := chi.URLParam(req, m.ID)
	val := chi.URLParam(req, m.Value)

	metric, err := m.NewMetric(name, kind, val)
	if err != nil {
		l.Warn("Invalid metric value or type")
		http.Error(rw, m.BadRequestMessage, http.StatusBadRequest)
		return
	}
	bytes, err := metric.MarshalJSON()
	if err != nil {
		l.Warn("Marshal error", zap.Error(err))
	}

	_, err = storage.Write(bytes)
	if err != nil {
		l.Fatal("Put to storage error", zap.Error(err))
	}

	rw.WriteHeader(http.StatusOK)
}

func GetHandler(rw http.ResponseWriter, req *http.Request) {
	l.Debug("GET HANDLER starts ...")

	kind := chi.URLParam(req, m.Mtype)
	name := chi.URLParam(req, m.ID)

	metric, err := m.NewMetric(name, kind, nil)
	if errors.Is(m.ErrInvalidType, err) {
		l.Warn("Invalid metric type")
		http.Error(rw, m.BadRequestMessage, http.StatusBadRequest)
		return
	}

	bytes, err := metric.MarshalJSON()
	if err != nil {
		l.Warn("Marshal problem", zap.Error(err))
	}

	newBytes, err := storage.Read(bytes)
	if err != nil {
		l.Warn("Coundn't fetch the metric from storage")
		http.Error(rw, m.NotFoundMessage, http.StatusNotFound)
		return
	}
	var numStr string
	if kind == m.Counter {
		numBytes := gjson.GetBytes(newBytes, m.Delta)
		numStr = strconv.FormatInt(numBytes.Int(), 10)
	} else {
		numBytes := gjson.GetBytes(newBytes, m.Value)
		numStr = strconv.FormatFloat(numBytes.Float(), 'f', -1, 64)
	}

	rw.WriteHeader(http.StatusOK)
	_, _ = rw.Write([]byte(numStr))
}

func GetAllHandler(rw http.ResponseWriter, req *http.Request) {
	l.Debug("GET ALL HANDLER starts ...")
	list := make([]Item, 0, m.MetricsNumber)

	var metric m.Metrics
	for _, bytes := range getList(storage) {
		_ = metric.UnmarshalJSON(bytes)
		list = append(list, Item{Met: metric.String()})
	}

	html, err := renderGetAll(list)
	if err != nil {
		l.Warn("An error occured during html rendering")
		http.Error(rw, m.InternalErrorMsg, http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "text/html")
	rw.WriteHeader(http.StatusOK)
	_, _ = rw.Write(html.Bytes())
}

func UpdateJSON(rw http.ResponseWriter, req *http.Request) {
	l.Info("UpdateJSON starts...")
	bytes, err := io.ReadAll(req.Body)
	if err != nil {
		l.Warn("Couldn't read with decompress")
	}
	defer req.Body.Close()

	typeBytes := gjson.GetBytes(bytes, m.Mtype)
	mtype := typeBytes.String()
	if mtype != m.Counter && mtype != m.Gauge {
		l.Info("Invalid metric type")
		http.Error(rw, m.BadRequestMessage, http.StatusBadRequest)
		return
	}
	if _, err = storage.Write(bytes); err != nil {
		l.Warn("Coudn't save data to storage")
	}
	newBytes := bytes
	if mtype == m.Counter {
		newBytes, err = storage.Read(bytes)
		if err != nil {
			l.Warn("Coulnd't fetch the metric from storage")
		}
	}

	rw.WriteHeader(http.StatusOK)
	_, _ = rw.Write(newBytes)
}

func GetJSON(rw http.ResponseWriter, req *http.Request) {
	l.Info("GetJSON starts...")
	bytes, err := io.ReadAll(req.Body)
	if err != nil {
		l.Warn("Couldn't read with decompress")
	}
	defer req.Body.Close()

	typeBytes := gjson.GetBytes(bytes, m.Mtype)
	mtype := typeBytes.String()
	if mtype != m.Counter && mtype != m.Gauge {
		l.Info("Invalid metric type")
		http.Error(rw, m.BadRequestMessage, http.StatusBadRequest)
		return
	}

	newBytes, err := storage.Read(bytes)
	if err != nil {
		http.Error(rw, m.NotFoundMessage, http.StatusNotFound)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	_, _ = rw.Write(newBytes)
}
