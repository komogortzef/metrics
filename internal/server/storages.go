package server

import (
	"os"
	"sync"

	"metrics/internal/logger"
	m "metrics/internal/models"
)

var (
	storeInterval int
	fileStorePath string
	restore       bool
)

type (
	MemStorage struct {
		items map[string][]byte
		Mtx   *sync.RWMutex
	}

	FileStorage struct {
		Repository
		file *os.File
	}
)

func (ms *MemStorage) Write(input []byte) (int, error) {
	mtype, name := getInfo(input)
	ms.Mtx.Lock()
	if mtype == m.Counter {
		if old, ok := ms.items[name]; ok {
			input = addCounter(old, input)
		}
	}
	ms.items[name] = input
	lenItems := len(ms.items)
	ms.Mtx.Unlock()

	return lenItems, nil
}

func (ms *MemStorage) Read(output *[]byte) (int, error) {
	_, name := getInfo(*output)

	var err error
	ms.Mtx.RLock()
	data, ok := ms.items[name]
	ms.Mtx.RUnlock()
	if !ok {
		logger.Warn("NO VALUE!!")
		err = m.ErrNoVal
	}

	*output = make([]byte, len(data))
	copy(*output, data)

	return len(*output), err
}
