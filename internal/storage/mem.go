// Package storage ...
package storage

import (
	"log"
	"strconv"
)

type Storage interface {
	Save(data ...string) error
}

// MemStorage просто map
type MemStorage map[string]any

// Save с проверкой релевантных для типа зн-ий
func (s MemStorage) Save(data ...string) error {
	log.Println("Start saving data...")

	metType := data[0]
	metName := data[1]
	metVal := data[2]

	switch metType {
	case "gauge":
		num, err := strconv.ParseFloat(metVal, 64)
		if err != nil {
			log.Println("Ivalid gauge value")
			return StoreError{"Invalid gauge value"}
		}

		s[metName] = num

		log.Println(metType, metName, ":", metVal, ".", "The value is received")

	case "counter":
		num, err := strconv.ParseInt(metVal, 10, 64)
		if err != nil {
			log.Println("Invalid counter value")
			return StoreError{"Invalid counter value"}
		}

		// обновление(либо помещение в хранилище) зн-ия счетчика
		if metVal, ok := s[metName].(int64); ok {
			s[metName] = metVal + num
		} else {
			s[metName] = num
		}

		log.Println(metType, metName, ":", metVal, "\t", "the value is received")

	default:
		log.Println("Ivalid metric type")
		return StoreError{"Invalid metric type"}
	}

	return nil
}

// StoreError реализовал из-за придирок линтера
type StoreError struct {
	Err string
}

func (se StoreError) Error() string {
	return se.Err
}
