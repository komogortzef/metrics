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

func (a *MetricAccount) List() []string {
	list := make([]string, MetricsNumber)

	for key := range a.gauges {
		list = append(list, key)
	}

	for key := range a.counters {
		list = append(list, key)
	}

	return list
}

func IsCounter(key string) bool {
	_, ok := Accounter.counters[key]
	return ok
}
