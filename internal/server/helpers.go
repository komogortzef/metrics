package server

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"sync"
	"time"

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

func NewFileStorage(interv int, path string, restore bool) (*FileStorage, error) {
	l.Debug("New file storage ...")
	fileStorage := FileStorage{
		MemStorage: NewMemStorage(),
		filePath:   path,
		interval:   time.Duration(interv),
	}

	var len int
	if restore {
		b, err := os.ReadFile(fileStorage.filePath)
		if err != nil {
			l.Warn("No file to restore!")
			goto returnWithoutRestore
		}
		buff := bytes.NewBuffer(b)
		scanner := bufio.NewScanner(buff)
		scanner.Split(bufio.ScanLines)
		for scanner.Scan() {
			len, err = fileStorage.Write(scanner.Bytes())
			if err != nil {
				return nil, fmt.Errorf("write to mem from file error: %w", err)
			}
		}
	}

returnWithoutRestore:
	l.Info("saved items number", zap.Int("items", len))
	return &fileStorage, nil
}

func addCounter(old []byte, input []byte) ([]byte, error) {
	var oldStruct m.Metrics
	err := oldStruct.UnmarshalJSON(old)
	if err != nil {
		return nil, fmt.Errorf("unmarshal error: %w", err)
	}
	numBytes := gjson.GetBytes(input, m.Delta)
	*oldStruct.Delta += numBytes.Int()

	return oldStruct.MarshalJSON()
}

func getList(storage Repository) [][]byte {
	l.Info("Get list...")

	switch s := storage.(type) {
	case *MemStorage:
		return s.listFromMem()
	case *FileStorage:
		return s.listFromMem()
	default:
		return nil
	}
}
