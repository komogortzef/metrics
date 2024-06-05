package server

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"sync"
	"time"

	"metrics/internal/logger"
	l "metrics/internal/logger"
	m "metrics/internal/models"

	"github.com/tidwall/gjson"
	"go.uber.org/zap"
)

func NewMemStorage() MemStorage {
	return MemStorage{
		items: make(map[string][]byte, m.MetricsNumber),
		Mtx:   &sync.RWMutex{},
	}
}

func NewFileStorage(interval int, path string, restore bool) (*FileStorage, error) {
	l.Debug("New file storage ...")
	fileStorage := FileStorage{
		MemStorage: NewMemStorage(),
		filePath:   path,
	}

	if restore {
		b, err := os.ReadFile(fileStorage.filePath)
		if err != nil {
			return nil, m.ErrRestoreFile
		}
		buff := bytes.NewBuffer(b)
		scanner := bufio.NewScanner(buff)
		scanner.Split(bufio.ScanLines)
		var len int
		for scanner.Scan() {
			len, err = fileStorage.MemStorage.Write(scanner.Bytes())
			if err != nil {
				return nil, fmt.Errorf("write to mem from file error: %w", err)
			}
		}
		l.Info("saved items number", zap.Int("items", len))
	}

	time.AfterFunc(time.Duration(interval)*time.Second, func() {
		if err := fileStorage.Dump(); err != nil {
			logger.Warn("Couldn't dump data!!")
		}
	})

	return &fileStorage, nil
}

func addCounter(old []byte, input []byte) ([]byte, error) {
	var oldStruct m.Metrics
	err := oldStruct.UnmarshalJSON(old)
	if err != nil {
		return nil, fmt.Errorf("Unmarshal error: %w", err)
	}
	numBytes := gjson.GetBytes(input, m.Delta)
	*oldStruct.Delta += numBytes.Int()

	return oldStruct.MarshalJSON()
}

func getList(storage Repository) [][]byte {
	l.Info("Get list...")

	switch s := storage.(type) {
	case *MemStorage:
		return listFromMem(s)
	case *FileStorage:
		return listFromMem(&s.MemStorage)
	default:
		return nil
	}
}

func listFromMem(ms *MemStorage) [][]byte {
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
