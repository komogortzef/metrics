package storage

import (
	"errors"
	"log"
	"regexp"
	"strconv"
)

type gauge map[string]float64
type counter map[string]int64

type MemStorage struct {
	gauges   gauge
	counters counter
}

var Mem MemStorage

func init() {
	Mem = MemStorage{
		gauges:   make(gauge),
		counters: make(counter),
	}
}

func (s *MemStorage) UpdateStorage(tp string, name string, val string) error {
	legalName := regexp.MustCompile(`^[A-Za-z]+[0-9]*$`)
	legalVal := regexp.MustCompile(`-?[0-9]*\.?[0-9]+`)

	switch tp {

	case "gauge":
		if !legalName.MatchString(name) {
			return errors.New("NotFound")
		}
		num, err := strconv.ParseFloat(val, 64)
		if err != nil || !legalVal.MatchString(val) {
			return errors.New("BadReq")
		}
		s.gauges[name] = num
		log.Println("gauge has been updated:", s.gauges)

	case "counter":
		if !legalName.MatchString(name) {
			return errors.New("NotFound")
		}
		num, err := strconv.ParseInt(val, 10, 64)
		if err != nil || !legalVal.MatchString(val) {
			return errors.New("BadReq")
		}
		s.counters[name] += num
		log.Println("counter has been updated:", s.counters)

	default:
		return errors.New("BadReq")
	}

	return nil
}
