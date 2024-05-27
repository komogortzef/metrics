package server

import (
	"sync"

	"metrics/internal/logger"
	"metrics/internal/models"
)

type MemStorage struct {
	Items map[string][]byte
	Mtx   *sync.RWMutex
}

func (ms *MemStorage) Update(key string, value []byte) error {
	logger.Info("Save mem")
	ms.Mtx.Lock()
	defer ms.Mtx.Unlock()
	if models.IsCounter(key) {
		value = accInt64(ms.Items[key], value)
	}
	ms.Items[key] = value

	return nil
}

func (ms *MemStorage) Get(key string) ([]byte, bool) {
	ms.Mtx.RLock()
	val, ok := ms.Items[key]
	ms.Mtx.RUnlock()

	return val, ok
}
