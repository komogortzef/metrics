package storage

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
)

type MemStorage map[string]any

var Mem = make(MemStorage)

func (s *MemStorage) Save(data ...[]byte) error {
	legalName := regexp.MustCompile(`^[A-Za-z]+[0-9]*$`)

	tp := string(data[0])
	name := string(data[1])
	val := string(data[2])

	switch tp {

	case "gauge":
		if !legalName.MatchString(name) {
			return errors.New("NotFound")
		}
		num, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return errors.New("BadReq")
		}
		(*s)[name] = num

	case "counter":
		if !legalName.MatchString(name) {
			return errors.New("NotFound")
		}
		num, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return errors.New("BadReq")
		}

		if val, ok := (*s)[name].(int64); ok {
			(*s)[name] = val + num
		} else {
			(*s)[name] = num
		}

		fmt.Println((*s)["PollCount"])

	default:
		return errors.New("BadReq")
	}

	return nil
}

func (s *MemStorage) Fetch(keys ...string) (any, error) {
	key := keys[0]

	val, ok := (*s)[key]
	if !ok {
		return nil, errors.New("no data available")
	}

	return val, nil

}
