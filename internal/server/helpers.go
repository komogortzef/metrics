package server

import (
	"strconv"
	"sync"

	"metrics/internal/logger"
)

func withAccInt64(a []byte, b []byte) ([]byte, error) {
	astr := string(a)
	bstr := string(b)

	anum, _ := strconv.ParseInt(astr, 10, 64)
	bnum, err := strconv.ParseInt(bstr, 10, 64)

	return []byte(strconv.FormatInt(anum+bnum, 10)), err
}

func SetStorage(st string) {
	logger.Info("Set storage...")
	switch st {
	default:
		storage = &MemStorage{
			Items: make(map[string][]byte, metricsNumber),
			Mtx:   &sync.RWMutex{},
		}
	}
}
