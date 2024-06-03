package server

import (
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
	Get(string) ([]byte, bool)
}

type MetricsManager struct {
	Serv    *http.Server
	Storage Repository
}

func (mm *MetricsManager) Run() error {

	return mm.Serv.ListenAndServe()
}

func (mm *MetricsManager) UpdateHandler(rw http.ResponseWriter, req *http.Request) {
	l.Debug("UPDATE HANDLER starts ...")
	metric, err := m.NewMetric(
		chi.URLParam(req, m.ID),
		chi.URLParam(req, m.Mtype),
		chi.URLParam(req, m.Value),
	)
	if err != nil {
		l.Warn("Invalid metric value or type")
		http.Error(rw, m.BadRequestMessage, http.StatusBadRequest)
		return
	}
	bytes, err := metric.MarshalJSON()
	if err != nil {
		l.Warn("Marshal error", zap.Error(err))
	}

	_, err = mm.Storage.Write(bytes)
	if err != nil {
		l.Warn("Put to mm.Storage error", zap.Error(err))
	}

	rw.WriteHeader(http.StatusOK)
}

func (mm *MetricsManager) GetHandler(rw http.ResponseWriter, req *http.Request) {
	l.Debug("GET HANDLER starts ...")
	kind := chi.URLParam(req, m.Mtype)
	name := chi.URLParam(req, m.ID)

	newBytes, ok := mm.Storage.Get(name)
	if !ok {
		l.Warn("Coundn't fetch the metric from mm.Storage")
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

func (mm *MetricsManager) GetAllHandler(rw http.ResponseWriter, req *http.Request) {
	l.Debug("GET ALL HANDLER starts ...")
	list := make([]Item, 0, m.MetricsNumber)

	var metric m.Metrics
	for _, bytes := range getList(mm.Storage) {
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

func (mm *MetricsManager) UpdateJSON(rw http.ResponseWriter, req *http.Request) {
	l.Info("UpdateJSON starts...")
	bytes, err := io.ReadAll(req.Body)
	if err != nil {
		l.Warn("Couldn't read with decompress")
	}
	defer req.Body.Close()

	mtype := gjson.GetBytes(bytes, m.Mtype).String()
	if mtype != m.Counter && mtype != m.Gauge {
		l.Info("Invalid metric type")
		http.Error(rw, m.BadRequestMessage, http.StatusBadRequest)
		return
	}
	if _, err = mm.Storage.Write(bytes); err != nil {
		l.Warn("Coudn't save data to mm.Storage")
	}
	newBytes := bytes
	if mtype == m.Counter {
		name := gjson.GetBytes(bytes, m.Delta).String()
		var ok bool
		newBytes, ok = mm.Storage.Get(name)
		if !ok {
			l.Warn("Coulnd't fetch the metric from mm.Storage")
		}
	}

	rw.WriteHeader(http.StatusOK)
	_, _ = rw.Write(newBytes)
}

func (mm *MetricsManager) GetJSON(rw http.ResponseWriter, req *http.Request) {
	l.Info("GetJSON starts...")
	bytes, err := io.ReadAll(req.Body)
	if err != nil {
		l.Warn("Couldn't read with decompress")
	}
	defer req.Body.Close()

	mtype := gjson.GetBytes(bytes, m.Mtype).String()
	if mtype != m.Counter && mtype != m.Gauge {
		l.Info("Invalid metric type")
		http.Error(rw, m.BadRequestMessage, http.StatusBadRequest)
		return
	}

	name := gjson.GetBytes(bytes, m.ID).String()
	newBytes, ok := mm.Storage.Get(name)
	if !ok {
		http.Error(rw, m.NotFoundMessage, http.StatusNotFound)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	_, _ = rw.Write(newBytes)
}
