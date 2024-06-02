package models

import (
	"errors"
	"fmt"
	"math"
	"strconv"
)

//go:generate ffjson $GOFILE
type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}

const (
	MetricsNumber = 29

	DefaultEndpoint       = "localhost:8080"
	DefaultPollInterval   = 2
	DefaultReportInterval = 10
	DefaultStoreInterval  = 300
	DefaultStorePath      = "/tmp/metrics-db.json"
	DefaultRestore        = "true"
	DefaultSendMode       = "text"

	InternalErrorMsg  = "internal server error"
	NotFoundMessage   = "not found"
	BadRequestMessage = "bad request"

	Gauge   = "gauge"
	Counter = "counter"
	Mtype   = "type"
	ID      = "id"
	Value   = "value"
	Delta   = "delta"
)

var (
	ErrNoVal       = errors.New("metric without val")
	ErrInvalidType = errors.New("invalid type")
)

func NewMetric(id, mtype string, val any) (Metrics, error) {
	var metric Metrics
	if mtype != Counter && mtype != Gauge {
		return metric, ErrInvalidType
	}
	metric.MType = mtype
	metric.ID = id
	switch v := val.(type) {
	case nil:
		if mtype == Counter {
			delta := int64(math.MaxInt64)
			metric.Delta = &delta
		} else {
			value := float64(math.MaxFloat64)
			metric.Value = &value
		}
	case int64:
		metric.Delta = &v
	case float64:
		metric.Value = &v
	case string:
		if mtype == Counter {
			num, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				return metric, ErrNoVal
			}
			metric.Delta = &num
		} else {
			num, err := strconv.ParseFloat(v, 64)
			if err != nil {
				return metric, ErrNoVal
			}
			metric.Value = &num
		}
	}

	return metric, nil
}

func (met Metrics) Data() any {
	if met.MType == Counter {
		return *met.Delta
	}
	return *met.Value
}

func (met Metrics) String() string {
	if met.Delta == nil && met.Value == nil {
		return fmt.Sprintf(" %s: <empty>", met.ID)
	}
	if met.MType == Counter {
		return fmt.Sprintf(" %s: %d", met.ID, *met.Delta)
	}
	return fmt.Sprintf(" %s: %g", met.ID, *met.Value)
}
