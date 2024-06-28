package service

import (
	ctx "context"
	"errors"
	"fmt"
	"strconv"
	"time"

	log "metrics/internal/logger"

	"github.com/cenkalti/backoff/v4"
)

const ( // часто используемые строковые литералы в виде констант
	gauge   = "gauge"
	counter = "counter"
)

var (
	ErrInvalidVal  = errors.New("invalid metric value")
	ErrInvalidType = errors.New("invalid metric type")
)

//go:generate ffjson $GOFILE
type Metrics struct {
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
	ID    string   `json:"id"`
	MType string   `json:"type"`
}

func NewMetric(mtype, id string, val string) (*Metrics, error) {
	met := &Metrics{ID: id, MType: mtype}
	if !met.IsCounter() && !met.IsGauge() {
		return nil, ErrInvalidType
	}
	if val == "" {
		return met, nil
	}
	if met.IsCounter() {
		num, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return met, ErrInvalidVal
		}
		met.Delta = &num
	} else {
		num, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return met, ErrInvalidVal
		}
		met.Value = &num
	}
	return met, nil
}

func BuildMetric(name string, val any) *Metrics {
	var metric Metrics
	metric.ID = name
	switch v := val.(type) {
	case float64:
		metric.MType = gauge
		metric.Value = &v
	case int64:
		metric.MType = counter
		metric.Delta = &v
	}

	return &metric
}

func (met Metrics) String() string {
	if met.Delta == nil && met.Value == nil {
		return fmt.Sprintf(" (%s: <empty>)", met.ID)
	}
	if met.IsCounter() {
		return fmt.Sprintf(" (%s: %d)", met.ID, *met.Delta)
	}
	return fmt.Sprintf(" (%s: %g)", met.ID, *met.Value)
}

func (met *Metrics) MergeMetrics(met2 *Metrics) {
	if met2 == nil {
		return
	}
	if met2.IsCounter() {
		if met2.Delta == nil {
			return
		}
		*met.Delta += *met2.Delta
		return
	}
}

func (met Metrics) ToSlice() []any {
	if met.IsCounter() {
		return []any{met.ID, *met.Delta}
	}
	return []any{met.ID, *met.Value}
}

func (met *Metrics) IsGauge() bool {
	return met.MType == gauge
}

func (met *Metrics) IsCounter() bool {
	return met.MType == counter
}

func Retry(cx ctx.Context, fn func() error) error {
	log.Debug("Retry...")
	expBackoff := backoff.NewExponentialBackOff()
	expBackoff.InitialInterval = 1 * time.Second
	expBackoff.Multiplier = 3
	expBackoff.MaxInterval = 5 * time.Second
	expBackoff.MaxElapsedTime = 11 * time.Second

	return backoff.Retry(fn, backoff.WithContext(expBackoff, cx))
}
