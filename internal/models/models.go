package models

const MetricsNumber = 29

//go:generate ffjson $GOFILE
type (
	Metrics struct {
		ID    string   `json:"id"`
		MType string   `json:"type"`
		Delta *int64   `json:"delta,omitempty"`
		Value *float64 `json:"value,omitempty"`
	}

	MetricAccount struct {
		counters map[string]struct{}
		gauges   map[string]struct{}
	}
)

var Accounter = MetricAccount{
	counters: map[string]struct{}{},
	gauges:   map[string]struct{}{},
}

func (a *MetricAccount) Put(kind, name string) {
	if kind == "counter" {
		a.counters[name] = struct{}{}
	}
	a.gauges[name] = struct{}{}
}

func (a *MetricAccount) List() (list [MetricsNumber]string) {
	index := 0
	for key := range a.gauges {
		list[index] = key
		index++
	}

	index--
	for key := range a.counters {
		list[index] = key
		index++
	}

	return list
}

func IsCounter(key string) bool {
	_, ok := Accounter.counters[key]
	return ok
}
