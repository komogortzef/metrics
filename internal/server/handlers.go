package server

import (
	"errors"
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
	io.Writer
	Read(*[]byte) (int, error)
}

var storage Repository

func UpdateHandler(rw http.ResponseWriter, req *http.Request) {
	logger.Debug("UPDATE HANDLER starts ...")

	kind := chi.URLParam(req, m.Mtype)
	name := chi.URLParam(req, m.Id)
	val := chi.URLParam(req, m.Value)

	metric, err := m.NewMetric(name, kind, val)
	fmt.Println("pass metric:", metric.String())
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

	metric, err := m.NewMetric(name, kind, 1000)
	if errors.Is(m.ErrInvalidType, err) {
		logger.Warn("Invalid metric type")
		http.Error(rw, m.BadRequestMessage, http.StatusBadRequest)
		return
	}

	bytes, err := metric.MarshalJSON()
	if err != nil {
		logger.Warn("Marshal problem", zap.Error(err))
	}

	_, err = storage.Read(&bytes)
	if err != nil {
		logger.Warn("Coundn't fetch the metric from storage")
		http.Error(rw, m.NotFoundMessage, http.StatusNotFound)
		return
	}

	var numStr string
	if kind == m.Counter {
		logger.Info("is counter...")
		numBytes := gjson.GetBytes(bytes, m.Delta)
		numStr = strconv.FormatInt(numBytes.Int(), 10)
	} else {
		logger.Info("is gauge...")
		numBytes := gjson.GetBytes(bytes, m.Value)
		numStr = strconv.FormatFloat(numBytes.Float(), 'f', -1, 64)
	}

	rw.WriteHeader(http.StatusOK)
	_, _ = rw.Write([]byte(numStr))
}

func GetAllHandler(rw http.ResponseWriter, req *http.Request) {
	logger.Debug("GET ALL HANDLER starts ...")
	var metric m.Metrics
	list := make([]Item, 0, metricsNumber)

	for _, bytes := range getList(storage) {
		metric.UnmarshalJSON(bytes)
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
	mtype := typeBytes.String()
	if mtype != m.Counter && mtype != m.Gauge {
		logger.Info("Invalid metric type")
		http.Error(rw, m.BadRequestMessage, http.StatusBadRequest)
		return
	}
	if _, err = storage.Write(bytes); err != nil {
		logger.Warn("Coudn't save data to storage")
	}
	if mtype == m.Counter {
		_, err = storage.Read(&bytes)
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

	typeBytes := gjson.GetBytes(bytes, m.Mtype)
	mtype := typeBytes.String()
	if mtype != m.Counter && mtype != m.Gauge {
		logger.Info("Invalid metric type")
		http.Error(rw, m.BadRequestMessage, http.StatusBadRequest)
		return
	}

	_, err = storage.Read(&bytes)
	if err != nil {
		http.Error(rw, m.NotFoundMessage, http.StatusNotFound)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	_, _ = rw.Write(bytes)
}
