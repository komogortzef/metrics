package server

import (
	"sync"
)

type MemStorage struct {
	Items map[string][]byte
	Mtx   *sync.RWMutex
}

func (ms *MemStorage) Save(key string, value []byte, opers ...Operation) error {
	var err error
	ms.Mtx.Lock()
	defer ms.Mtx.Unlock()

	if len(opers) > 0 {
		for _, oper := range opers {
			if oper == nil {
				continue
			}
			value, err = oper(ms.Items[key], value)
		}
	}

	ms.Items[key] = value

	return err
}

func (ms *MemStorage) Get(key string) ([]byte, bool) {
	ms.Mtx.RLock()
	defer ms.Mtx.RUnlock()
	val, ok := ms.Items[key]

	return val, ok
}

func (ms *MemStorage) GetAll() map[string][]byte {
	return ms.Items
}
