// Package storage ...
package storage

import (
	"log"
	"strconv"
)

// MemStorage ...
type MemStorage map[string]any

// Save ...
func (s MemStorage) Save(data ...[]byte) error {
	log.Println("\nstart saving data...")

	typ := string(data[0])
	name := string(data[1])
	val := string(data[2])

	switch typ {
	case "gauge":
		num, err := strconv.ParseFloat(val, 64)
		if err != nil {
			log.Println("Ivalid gauge value")
			return StoreError{"Invalid gauge value"}
		}

		s[name] = num

		log.Println(typ, name, ":", val, ".", "The value is received")

	case "counter":
		num, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			log.Println("Invalid counter value")
			return StoreError{"Invalid counter value"}
		}

		if val, ok := s[name].(int64); ok {
			s[name] = val + num
		} else {
			s[name] = num
		}

		log.Println(typ, name, ":", val, "\t", "the value is received")

	default:
		log.Println("Ivalid metric type")
		return StoreError{"Invalid metric type"}
	}

	return nil
}

// Fetch ...
func (s MemStorage) Fetch(keys ...string) (any, error) {
	key := keys[0]

	val, ok := s[key]
	if !ok {
		log.Println("Failed to get the value")
		return nil, StoreError{"Failure to get the value"}
	}
	return val, nil
}

// StoreError ...
type StoreError struct {
	Err string
}

func (se StoreError) Error() string {
	return se.Err
}
