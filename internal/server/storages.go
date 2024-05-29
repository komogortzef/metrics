package server

import (
	"sync"

	"metrics/internal/logger"
)

type MemStorage struct {
	Items map[string]string
	Mtx   *sync.RWMutex
}

func (ms *MemStorage) Update(key string, newVal string) error {
	if isCounter(key) {
		old, ok := storage.Get(key)
		if ok {
			newVal = countAdd(old, newVal)
		}
	}

	ms.Mtx.Lock()
	ms.Items[key] = newVal
	ms.Mtx.Unlock()

	logger.Info("value saved")
	return nil
}

func (ms *MemStorage) Get(key string) (string, bool) {
	ms.Mtx.RLock()
	val, ok := ms.Items[key]
	ms.Mtx.RUnlock()

	return val, ok
}
