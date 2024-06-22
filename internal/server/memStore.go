package server

import (
	ctx "context"
	"errors"
	"sync"

	log "metrics/internal/logger"
	s "metrics/internal/service"
)

const metricsNumber = 29

var ErrNoValue = errors.New("no such value in storage")

type MemStorage struct {
	items map[string]*s.Metrics
	mtx   *sync.RWMutex
	len   int
}

func NewMemStore() *MemStorage {
	return &MemStorage{
		items: make(map[string]*s.Metrics, metricsNumber),
		mtx:   &sync.RWMutex{},
	}
}

func (ms *MemStorage) Put(_ ctx.Context, met *s.Metrics) (*s.Metrics, error) {
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

func (ms *MemStorage) Get(_ ctx.Context, m *s.Metrics) (*s.Metrics, error) {
	var err error
	ms.mtx.RLock()
	met, ok := ms.items[m.ID]
	ms.mtx.RUnlock()
	if !ok {
		err = ErrNoValue
	}
	return met, err
}

func (ms *MemStorage) List(_ ctx.Context) ([]*s.Metrics, error) {
	i := 0
	ms.mtx.RLock()
	metrics := make([]*s.Metrics, ms.len)
	for _, met := range ms.items {
		metrics[i] = met
		i++
	}
	ms.mtx.RUnlock()
	return metrics, nil
}

func (ms *MemStorage) PutBatch(cx ctx.Context, mets []*s.Metrics) error {
	for _, metric := range mets {
		if _, err := ms.Put(cx, metric); err != nil {
			log.Warn("Mem PutBatch: couldn't insert batch to file or mem store")
			return err
		}
	}
	return nil
}

func (ms *MemStorage) Ping(_ ctx.Context) error {
	log.Info("Memory storage is working...")
	return nil
}

func (ms *MemStorage) Close() {
	log.Info("Memory storage is closed;)")
}
