package server

import (
	"fmt"
	"os"
	"sync"
	"time"

	l "metrics/internal/logger"

	"github.com/jackc/pgx/v5/pgxpool"
)

type (
	MemStorage struct {
		Items map[string][]byte
		len   int
		Mtx   *sync.RWMutex
	}

	FileStorage struct {
		MemStorage
		FilePath string
		Interval time.Duration
		Restore  bool
	}

	DataBase struct {
		*pgxpool.Pool
		Addr string
	}
)

func (ms *MemStorage) Put(name string, input []byte, helpers ...helper) (int, error) {
	var err error
	ms.Mtx.Lock()
	old, exists := ms.Items[name]
	for _, helper := range helpers { // применение доп логики перед сохранением
		if exists {
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

func (ms *MemStorage) Get(name string) ([]byte, bool) {
	ms.Mtx.RLock()
	data, ok := ms.Items[name]
	ms.Mtx.RUnlock()

	return data, ok
}

func (ms *MemStorage) Pop(name string) ([]byte, error) {
	data, exists := ms.Items[name]
	if !exists {
		return nil, fmt.Errorf("no such metric")
	}
	delete(ms.Items, name)
	return data, nil
}

func (fs *FileStorage) Put(name string, data []byte, helpers ...helper) (int, error) {
	len, err := fs.MemStorage.Put(name, data, helpers...)

	if fs.Interval == 0 { // синхронная запись в файл при поступлении данных
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
	for _, met := range ms.Items {
		metrics[i] = met
		i++
	}
	ms.Mtx.RUnlock()
	return metrics
}

func (fs *FileStorage) dump() error {
	l.Info("Dump starts")
	var buf []byte

	// объединение всех метрик в один байтовый срез(разделение с помощью '\n')
	fs.Mtx.RLock()
	for _, data := range fs.Items {
		data = append(data, byte('\n'))
		buf = append(buf, data...)
	}
	fs.Mtx.RUnlock()

	return os.WriteFile(fs.FilePath, buf, 0666)
}

func (fs *FileStorage) startTicker() {
	l.Warn("fs.startTicker()...")
	ticker := time.NewTicker(fs.Interval * time.Second)

	go func() {
		for {
			<-ticker.C
			if err := fs.dump(); err != nil {
				l.Warn("fs.startTicker(): Couldn't save data to file")
				return
			}
		}
	}()
}

func (db *DataBase) Put(name string, data []byte, helps ...helper) (int, error) {
	return 0, nil
}

func (db *DataBase) Get(key string) ([]byte, bool) {
	return nil, false
}

func (db *DataBase) Pop(name string) ([]byte, error) {
	return nil, nil
}
