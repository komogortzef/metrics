package server

import (
	"context"
	"errors"
	"sync"
)

const (
	metricsNumber = 29
)

type MemStorage struct {
	items map[string][]byte
	len   int
	mtx   *sync.RWMutex
}

func NewMemStore() *MemStorage {
	return &MemStorage{
		items: make(map[string][]byte, metricsNumber),
		mtx:   &sync.RWMutex{},
	}
}

func (ms *MemStorage) Put(ctx context.Context,
	name string, input []byte, helps ...helper) error {
	var err error
	ms.mtx.Lock()
	old, exists := ms.items[name]
	for _, helper := range helps {
		if helper != nil && exists {
			if input, err = helper(old, input); err != nil {
				return err
			}
		}
	}
	ms.items[name] = input
	if !exists {
		ms.len++
	}
	ms.mtx.Unlock()
	return err
}

func (ms *MemStorage) Get(ctx context.Context, name string) ([]byte, error) {
	var err error
	ms.mtx.RLock()
	data, ok := ms.items[name]
	ms.mtx.RUnlock()
	if !ok {
		err = errors.New("no such metric")
	}

	return data, err
}

func (ms *MemStorage) List(ctx context.Context) ([][]byte, error) {
	i := 0
	ms.mtx.RLock()
	metrics := make([][]byte, ms.len)
	for _, met := range ms.items {
		metrics[i] = met
		i++
	}
	ms.mtx.RUnlock()
	return metrics, nil
}
