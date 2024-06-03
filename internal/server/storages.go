package server

import (
	"fmt"
	"sync"
	"time"

	l "metrics/internal/logger"
	m "metrics/internal/models"
)

type (
	MemStorage struct {
		items map[string][]byte
		len   int
		Mtx   *sync.RWMutex
	}

	FileStorage struct {
		Repository
		filePath      string
		storeInterval int
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

func (fs *FileStorage) Write(input []byte) (int, error) {
	l.Info("File Store write...")
	n, err := fs.Write(input)
	if err != nil {
		return n, fmt.Errorf("save to memory: %w", err)
	}

	time.AfterFunc(time.Duration(fs.storeInterval)*time.Second, func() {
		err = dump(fs.filePath, fs.Repository)
	})

	return 0, nil
}
