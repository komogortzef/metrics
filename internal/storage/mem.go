package storage

import (
	"errors"
	"fmt"
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

	switch tp {

	case "gauge":
		legalName := regexp.MustCompile(`^[A-Za-z]+[0-9]*$`)
		if !legalName.MatchString(name) {
			return errors.New("NotFound")
		}
		legalFloat := regexp.MustCompile(`[-]?[0-9]*\.[0-9]+`)
		num, err := strconv.ParseFloat(val, 64)
		if err != nil || !legalFloat.MatchString(val) {
			return errors.New("BadReq")
		}
		s.gauges[name] = num
		fmt.Println(s.gauges)

	case "counter":
		legalName := regexp.MustCompile(`^[A-Za-z]+[0-9]*$`)
		if !legalName.MatchString(name) {
			return errors.New("NotFound")
		}
		legalInt := regexp.MustCompile(`[-]?[0-9]+`)
		num, err := strconv.ParseInt(val, 10, 64)
		if err != nil || !legalInt.MatchString(val) {
			return errors.New("BadReq")
		}
		s.counters[name] += num
		fmt.Println(s.counters)

	default:
		return errors.New("BadReq")
	}

	return nil
}
