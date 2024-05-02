package storage

import (
	"log"
	"strconv"
)

type MemStorage map[string]any

func (s MemStorage) Save(data ...[]byte) error {
	log.Println("\nstart saving data...")

	tp := string(data[0])
	name := string(data[1])
	val := string(data[2])

	switch tp {

	case "gauge":
		num, err := strconv.ParseFloat(val, 64)
		if err != nil {
			log.Println("Ivalid gauge value")
			return StoreErr{"Invalid gauge value"}
		}
		s[name] = num
		log.Println(tp, name, ":", val, ".", "The value is received")

	case "counter":
		num, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			log.Println("Invalid counter value")
			return StoreErr{"Invalid counter value"}
		}

		if val, ok := s[name].(int64); ok {
			s[name] = val + num
		} else {
			s[name] = num
		}
		log.Println(tp, name, ":", val, "\t", "the value is received")

	default:
		log.Println("Ivalid metric type")
		return StoreErr{"Invalid metric type"}
	}

	return nil
}

func (s MemStorage) Fetch(keys ...string) (any, error) {
	key := keys[0]

	val, ok := s[key]
	if !ok {
		log.Println("Failed to get the value")
		return nil, StoreErr{"Failure to get the value"}
	}
	return val, nil
}

type StoreErr struct {
	Err string
}

func (se StoreErr) Error() string {
	return se.Err
}
