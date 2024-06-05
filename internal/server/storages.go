package server

import (
	"os"
	"sync"
	"time"

	l "metrics/internal/logger"
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
		interval time.Duration
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

func (fs *FileStorage) Write(input []byte) (int, error) {
	l.Info("File Store Write...")
	len, err := fs.MemStorage.Write(input)

	if fs.interval == 0 {
		if err = fs.dump(); err != nil {
			return len, err
		}
	}

	return len, err
}

func (ms *MemStorage) listFromMem() [][]byte {
	metrics := make([][]byte, ms.len)
	i := 0
	ms.Mtx.RLock()
	for _, met := range ms.items {
		metrics[i] = met
		i++
	}
	ms.Mtx.RUnlock()
	return metrics
}

func (fs *FileStorage) dump() error {
	l.Info("Dump starts")
	var buf []byte

	fs.Mtx.RLock()
	for _, data := range fs.items {
		data = append(data, byte('\n'))
		buf = append(buf, data...)
	}
	fs.Mtx.RUnlock()

	return os.WriteFile(fs.filePath, buf, 0666)
}

func (fs *FileStorage) StartTicker() {
	ticker := time.NewTicker(fs.interval * time.Second)

	go func() {
		for {
			<-ticker.C
			if err := fs.dump(); err != nil {
				l.Warn("Coulnd't save data to file")
				return
			}
		}
	}()
}
