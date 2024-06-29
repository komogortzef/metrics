package server

import (
	"strings"

	s "metrics/internal/service"
)

func processURL(url string) (string, string, string) {
	strs := strings.Split(url, "/")[2:]
	switch len(strs) {
	case 0:
		return "", "", ""
	case 1:
		return strs[0], "", ""
	case 2:
		return strs[0], strs[1], ""
	}
	return strs[0], strs[1], strs[2]
}

func getQuery(oper dbOperation, met *s.Metrics) string {
	switch oper {
	case insertMetric:
		if met.IsGauge() {
			return insertGauge
		}
		return insertCounter
	default:
		if met.IsGauge() {
			return selectGauge
		}
		return selectCounter
	}
}

func setVal(met *s.Metrics, val any) {
	if v, ok := val.(int64); ok {
		met.Delta = &v
	} else {
		v, _ := val.(float64)
		met.Value = &v
	}
}
