package server

import "sync"

const metricsNumber = 29

type MemStorage struct {
	Items map[string][]byte
	Mtx   *sync.RWMutex
}

func (ms *MemStorage) Save(key string, value []byte, opers ...Operation) error {
	var err error
	ms.Mtx.Lock()
	defer ms.Mtx.Unlock()

	// если нужно произвести какие-либо действия со знчениями перед сохранением
	if len(opers) > 0 {
		for _, oper := range opers {
			if oper != nil {
				value, err = oper(ms.Items[key], value)
			}
		}
	}

	ms.Items[key] = value

	return err
}

func (ms *MemStorage) Get(key string) ([]byte, bool) {
	ms.Mtx.RLock()
	val, ok := ms.Items[key]
	ms.Mtx.RUnlock()

	return val, ok
}

func (ms *MemStorage) GetAll() map[string][]byte {
	res := make(map[string][]byte, metricsNumber)

	ms.Mtx.RLock()
	for name, val := range ms.Items {
		res[name] = val
	}
	ms.Mtx.RUnlock()

	return res
}
