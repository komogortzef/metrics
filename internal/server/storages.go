package server

import (
	"os"
	"sync"

	m "metrics/internal/models"

	"github.com/tidwall/gjson"
)

type (
	MemStorage struct {
		items map[string][]byte
		len   int
		Mtx   *sync.RWMutex
	}

	FileStorage struct {
		MemStorage
		filePath string
	}
)

func (ms *MemStorage) Write(input []byte) (int, error) {
	mtype := gjson.GetBytes(input, m.Mtype).String()
	name := gjson.GetBytes(input, m.ID).String()
	var err error
	ms.Mtx.Lock()
	old, exists := ms.items[name]
	if mtype == m.Counter && exists {
		input, err = addCounter(old, input)
	}
	ms.items[name] = input
	if !exists {
		ms.len++
	}
	ms.Mtx.Unlock()

	return ms.len, err
}

func (ms *MemStorage) Get(name string) ([]byte, bool) {
	ms.Mtx.RLock()
	data, ok := ms.items[name]
	ms.Mtx.RUnlock()

	return data, ok
}

func (fs *FileStorage) Dump() error {
	var buf []byte

	fs.Mtx.RLock()
	for _, data := range fs.items {
		data = append(data, byte('\n'))
		buf = append(buf, data...)
	}
	fs.Mtx.RUnlock()

	return os.WriteFile(fs.filePath, buf, 0666)
}
