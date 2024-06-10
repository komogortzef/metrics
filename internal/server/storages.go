package server

import (
	"bufio"
	"bytes"
	"os"
	"sync"
	"time"

	l "metrics/internal/logger"
	m "metrics/internal/models"

	"github.com/tidwall/gjson"
	"go.uber.org/zap"
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
)

func (ms *MemStorage) Put(name string, input []byte, help helper) (int, error) {
	var err error
	ms.Mtx.Lock()
	old, exists := ms.Items[name]
	if exists && help != nil {
		input, err = help(old, input)
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

func (ms *MemStorage) List() [][]byte {
	i := 0
	ms.Mtx.RLock()
	metrics := make([][]byte, ms.len)
	for _, met := range ms.Items {
		metrics[i] = met
		i++
	}
	ms.Mtx.RUnlock()

	return metrics
}

func (fs *FileStorage) Put(name string, data []byte, help helper) (int, error) {
	len, err := fs.MemStorage.Put(name, data, help)

	if fs.Interval == 0 { // синхронная запись в файл при поступлении данных
		if err = fs.dump(); err != nil {
			return len, err
		}
	}

	return len, err
}

func (fs *FileStorage) dump() error {
	l.Info("Dump...")
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

func (fs *FileStorage) restoreFromFile() (len int) {
	b, err := os.ReadFile(fs.FilePath)
	if err != nil {
		l.Warn("No file to restore!")
		return
	}
	buff := bytes.NewBuffer(b)
	scanner := bufio.NewScanner(buff)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		bytes := scanner.Bytes()
		name := gjson.GetBytes(bytes, m.ID).String()
		len, _ = fs.Put(name, bytes, nil)
	}
	l.Info("number of metrics recovered from the file", zap.Int("len", len))
	return

}
