package models

import (
	"errors"
	"fmt"
	"metrics/internal/logger"
	"strconv"
)

//go:generate ffjson $GOFILE
type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}

// обертки для того чтобы обойти систему типов для методов интерферса Repository

const (
	InternalErrorMsg  = "internal server error"
	NotFoundMessage   = "not found"
	BadRequestMessage = "bad request"
	Gauge             = "gauge"
	Counter           = "counter"
	Mtype             = "type"
	Id                = "id"
	Value             = "value"
	Delta             = "delta"
)

var (
	ErrNoVal       = errors.New("metric without val")
	ErrInvalidType = errors.New("invalid type")
)

func NewMetric(id, mtype string, val any) (Metrics, error) {
	logger.Info("New Metric func...")

	var metric Metrics
	if mtype != Counter && mtype != Gauge {
		return metric, ErrInvalidType
	}
	metric.MType = mtype
	metric.ID = id

	switch v := val.(type) {
	case int64:
		metric.Delta = &v
	case float64:
		metric.Value = &v
	case string:
		logger.Info("string...")
		if mtype == Counter {
			logger.Info("int64string...")
			num, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				return metric, ErrNoVal
			}
			metric.Delta = &num
		} else {
			logger.Info("float64string...")
			num, err := strconv.ParseFloat(v, 64)
			if err != nil {
				return metric, ErrNoVal
			}
			metric.Value = &num
		}
	}

	return metric, nil
}

func (met Metrics) String() string {
	if met.Delta == nil && met.Value == nil {
		return fmt.Sprintf("%s, %s: <empty>", met.MType, met.ID)
	}
	if met.MType == Counter {
		return fmt.Sprintf("%s, %s: %d", met.MType, met.ID, *met.Delta)
	}
	return fmt.Sprintf("%s, %s: %g", met.MType, met.ID, *met.Value)
}
