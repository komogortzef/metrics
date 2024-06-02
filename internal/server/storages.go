package server

import (
	"sync"

	m "metrics/internal/models"
)

type (
	MemStorage struct {
		items map[string][]byte
		len   int
		Mtx   *sync.RWMutex
	}
)

func (ms *MemStorage) Write(input []byte) (int, error) {
	mtype, name := getInfo(input)
	ms.Mtx.Lock()
	old, exists := ms.items[name]
	if mtype == m.Counter && exists {
		input = addCounter(old, input)
	}
	ms.items[name] = input
	if !exists {
		ms.len++
	}
	ms.Mtx.Unlock()

	return len(input), nil
}

func (ms *MemStorage) Read(p []byte) ([]byte, error) {
	_, name := getInfo(p)
	var err error
	ms.Mtx.RLock()
	data, ok := ms.items[name]
	ms.Mtx.RUnlock()
	if !ok {
		err = m.ErrNoVal
	}

	return data, err
}
