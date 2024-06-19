package server

import (
	"metrics/internal/service"
	"strings"
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

func getQuery(oper dbOperation, mtype string) string {
	switch oper {
	case insertMetric:
		if mtype == service.Gauge {
			return insertGauge
		}
		return insertCounter
	default:
		if mtype == service.Gauge {
			return selectGauge
		}
		return selectCounter
	}
}
