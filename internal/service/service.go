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
	DefaultEndpoint       = "localhost:8080"
	DefaultPollInterval   = 2
	DefaultReportInterval = 10
	DefaultStoreInterval  = 300
	DefaultStorePath      = "/tmp/metrics-db.json"
	DefaultRestore        = true
	DefaultSendMode       = "text"
	NoStorage             = ""
	InternalErrorMsg      = "internal server error"
	NotFoundMessage       = "not found"
	BadRequestMessage     = "bad request"
	Gauge                 = "gauge"
	Counter               = "counter"
	Mtype                 = "type"
	ID                    = "id"
	Value                 = "value"
	Delta                 = "delta"
)

var (
	ErrInvalidVal  = errors.New("invalid metric value")
	ErrInvalidType = errors.New("invalid metric type")
)

//go:generate ffjson $GOFILE
type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}

func NewMetric(mtype, id string, val string) (*Metrics, error) {
	var metric Metrics
	if mtype != Counter && mtype != Gauge {
		return &metric, ErrInvalidType
	}
	metric.MType = mtype
	metric.ID = id
	if mtype == Counter {
		num, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return &metric, ErrInvalidVal
		}
		metric.Delta = &num
	} else {
		num, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return &metric, ErrInvalidVal
		}
		metric.Value = &num
	}
	return &metric, nil
}

func BuildMetric(name string, val any) Metrics {
	var metric Metrics
	metric.ID = name
	switch v := val.(type) {
	case float64:
		metric.MType = Gauge
		metric.Value = &v
	case int64:
		metric.MType = Counter
		metric.Delta = &v
	}

	return metric
}

func (met Metrics) String() string {
	if met.Delta == nil && met.Value == nil {
		return fmt.Sprintf(" (%s: <empty>)", met.ID)
	}
	if met.MType == Counter {
		return fmt.Sprintf(" (%s: %d)", met.ID, *met.Delta)
	}
	return fmt.Sprintf(" (%s: %g)", met.ID, *met.Value)
}

func (met *Metrics) MergeMetrics(met2 *Metrics) {
	if met2 == nil {
		return
	}
	if met2.MType == Counter {
		if met2.Delta == nil {
			return
		}
		*met.Delta += *met2.Delta
		return
	}
}

func (met Metrics) ToSlice() []any {
	if met.MType == Counter {
		return []any{met.ID, *met.Delta}
	}
	return []any{met.ID, *met.Value}
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
