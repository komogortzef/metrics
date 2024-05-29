package server

import (
	"errors"
	"metrics/internal/logger"
	"metrics/internal/models"
	"regexp"
	"strconv"
)

const metricsNumber = 29

type Repository interface {
	Update(key string, value string) error
	Get(key string) (string, bool)
}

var storage Repository

var (
	isFloat = regexp.MustCompile(`^-?\d+(\.\d+)?([eE][+-]?\d+)?$`).MatchString
	isInt   = regexp.MustCompile(`^(0|[1-9]\d*)$`).MatchString

	errBadRequest = errors.New("invalid value or metric type")
)

type MetricAccount struct {
	counters map[string]struct{}
	gauges   map[string]struct{}
}

var accounter = MetricAccount{
	counters: map[string]struct{}{},
	gauges:   map[string]struct{}{},
}

func (a *MetricAccount) put(kind, name, val string) error {
	logger.Info("account putting...")
	var err error
	switch kind {
	case counter:
		if isInt(val) {
			logger.Info("is counter")
			a.counters[name] = struct{}{}
			err = storage.Update(name, val)
		} else {
			err = errBadRequest
		}
	case gauge:
		if isFloat(val) {
			logger.Info("is float")
			a.gauges[name] = struct{}{}
			err = storage.Update(name, val)
		} else {
			err = errBadRequest
		}
	default:
		err = errBadRequest
	}
	return err
}

func (a *MetricAccount) putJSON(jsonData models.Metrics) error {
	logger.Info("accounter json putting...")
	var val string
	if jsonData.MType == counter {
		val = strconv.FormatInt(*jsonData.Delta, 10)
	} else {
		val = strconv.FormatFloat(*jsonData.Value, 'f', -1, 64)
	}

	return a.put(jsonData.MType, jsonData.ID, val)
}

func (a *MetricAccount) list() (list []string) {
	for key := range a.gauges {
		list = append(list, key)
	}
	for key := range a.counters {
		list = append(list, key)
	}

	return list
}

func isCounter(key string) bool {
	_, ok := accounter.counters[key]
	return ok
}
