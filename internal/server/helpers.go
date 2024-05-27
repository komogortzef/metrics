package server

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"strconv"
	"sync"

	"metrics/internal/logger"
	"metrics/internal/models"
)

const bitSize = 8

func SetStorage(st string) {
	logger.Info("Set storage...")
	switch st {
	default:
		storage = &MemStorage{
			Items: make(map[string][]byte, models.MetricsNumber),
			Mtx:   &sync.RWMutex{},
		}
	}
}

func accInt64(a []byte, b []byte) []byte {
	var anum int64 = 0
	if len(a) > 0 {
		anum = int64(binary.LittleEndian.Uint64(a))
	}
	bnum := int64(binary.LittleEndian.Uint64(b))
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, uint64(anum+bnum))

	return bytes
}

func toBytes(kind, val any) ([]byte, error) {
	bytes := make([]byte, bitSize)

	switch v := val.(type) {
	case models.Metrics:
		if kind == gauge {
			val = *v.Value
		} else {
			val = *v.Delta
		}
	}

	switch num := val.(type) {
	case int64:
		binary.LittleEndian.PutUint64(bytes, uint64(num))
		return bytes, nil
	case float64:
		binary.LittleEndian.PutUint64(bytes, math.Float64bits(num))
		return bytes, nil
	case string:
		if kind == gauge {
			v, err := strconv.ParseFloat(num, 64)
			if err != nil {
				return nil, fmt.Errorf("cannot parse float: %w", err)
			}
			binary.LittleEndian.PutUint64(bytes, math.Float64bits(v))
			return bytes, nil
		} else if kind == counter {
			v, err := strconv.ParseInt(num, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("cannot parse int: %w", err)
			}
			binary.LittleEndian.PutUint64(bytes, uint64(v))
			return bytes, nil
		}
	}

	return nil, errors.New("bad request")
}

func bytesToString(name string, val []byte) string {
	var res string

	if models.IsCounter(name) {
		res = strconv.FormatInt(int64(binary.LittleEndian.Uint64(val)), 10)
	} else {
		res = strconv.FormatFloat(
			math.Float64frombits(binary.LittleEndian.Uint64(val)), 'g', -1, 64)
	}

	return res
}
