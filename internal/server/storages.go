package server

import (
	"os"
	"sync"

	"metrics/internal/logger"
)

var ()

type (
	MemStorage struct {
		Items map[string]string
		Mtx   *sync.RWMutex
	}

	FileStorage struct {
		File   *os.File
		Buffer *MemStorage
	}
)

func SetStorage(st string) {
	switch st {
	default:
		storage = &MemStorage{
			Items: make(map[string]string, metricsNumber),
			Mtx:   &sync.RWMutex{},
		}
	}
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

func (fs *FileStorage) Update(key string, newVal string) error {

	return nil
}

func (fs *FileStorage) Get(key string) (string, bool) {

	return "", false
}
