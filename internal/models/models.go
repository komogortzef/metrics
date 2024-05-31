package models

//go:generate ffjson $GOFILE
type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}

func NewMetric(id string, val any) Metrics {
	var metric Metrics
	metric.ID = id
	switch v := val.(type) {
	case int64:
		metric.MType = "counter"
		metric.Delta = &v
	case float64:
		metric.MType = "gauge"
		metric.Value = &v
	}

	return metric
}
