package server

import (
	ctx "context"
	"errors"
	"sync"

	s "metrics/internal/service"
)

const metricsNumber = 29

var ErrNoValue = errors.New("no such value in storage")

type MemStorage struct {
	items map[string]s.Metrics
	len   int
	mtx   *sync.RWMutex
}

func NewMemStore() *MemStorage {
	return &MemStorage{
		items: make(map[string]s.Metrics, metricsNumber),
		mtx:   &sync.RWMutex{},
	}
}

func (ms *MemStorage) Put(_ ctx.Context, met s.Metrics) (s.Metrics, error) {
	ms.mtx.Lock()
	oldMet, exists := ms.items[met.ID]
	met.MergeMetrics(oldMet)
	ms.items[met.ID] = met
	if !exists {
		ms.len++
	}
	ms.mtx.Unlock()
	return met, nil
}

func (ms *MemStorage) Get(_ ctx.Context, m s.Metrics) (s.Metrics, error) {
	var err error
	ms.mtx.RLock()
	met, ok := ms.items[m.ID]
	ms.mtx.RUnlock()
	if !ok {
		err = ErrNoValue
	}
	return met, err
}

func (ms *MemStorage) List(ctx ctx.Context) ([]s.Metrics, error) {
	i := 0
	ms.mtx.RLock()
	metrics := make([]s.Metrics, ms.len)
	for _, met := range ms.items {
		metrics[i] = met
		i++
	}
	ms.mtx.RUnlock()
	return metrics, nil
}
