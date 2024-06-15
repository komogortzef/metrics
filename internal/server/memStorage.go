package server

import (
	"errors"
	"sync"
)

type MemStorage struct {
	Items map[string][]byte
	len   int
	Mtx   *sync.RWMutex
}

func (ms *MemStorage) Put(name string, input []byte, helps ...helper) (int, error) {
	var err error
	ms.Mtx.Lock()
	old, exists := ms.Items[name]
	for _, helper := range helps {
		if helper != nil && exists {
			input, err = helper(old, input)
		}
	}
	ms.Items[name] = input
	if !exists {
		ms.len++
	}
	ms.Mtx.Unlock()

	return ms.len, err
}

func (ms *MemStorage) Get(name string) ([]byte, error) {
	var err error
	ms.Mtx.RLock()
	data, ok := ms.Items[name]
	ms.Mtx.RUnlock()
	if !ok {
		err = errors.New("no such metric")
	}

	return data, err
}

func (ms *MemStorage) List() ([][]byte, error) {
	i := 0
	ms.Mtx.RLock()
	metrics := make([][]byte, ms.len)
	for _, met := range ms.Items {
		metrics[i] = met
		i++
	}
	ms.Mtx.RUnlock()
	return metrics, nil
}
