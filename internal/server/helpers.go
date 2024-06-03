package server

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"sync"

	l "metrics/internal/logger"
	m "metrics/internal/models"

	"github.com/tidwall/gjson"
	"go.uber.org/zap"
)

func newMemStorage() Repository {
	return &MemStorage{
		items: make(map[string][]byte, m.MetricsNumber),
		Mtx:   &sync.RWMutex{},
	}
}

func NewFileStorage(interval int, path string, restore bool) error {
	l.Debug("New file storage ...")
	fileStorage := FileStorage{
		Repository:    newMemStorage(),
		filePath:      path,
		storeInterval: interval,
	}

	if restore {
		b, err := os.ReadFile(fileStorage.filePath)
		if err != nil {
			return fmt.Errorf("read file error: %w", err)
		}
		buff := bytes.NewBuffer(b)
		scanner := bufio.NewScanner(buff)
		scanner.Split(bufio.ScanLines)
		for scanner.Scan() {
			_, err = fileStorage.Repository.Write(scanner.Bytes())
			if err != nil {
				return fmt.Errorf("write to mem error: %w", err)
			}
		}
	}

	storage = &fileStorage
	return nil
}

func addCounter(old []byte, input []byte) []byte {
	var oldStruct m.Metrics
	err := oldStruct.UnmarshalJSON(old)
	if err != nil {
		l.Warn("UNMARSHAL problems", zap.Error(err))
	}
	numBytes := gjson.GetBytes(input, m.Delta)
	*oldStruct.Delta += numBytes.Int()
	bytes, err := oldStruct.MarshalJSON()
	if err != nil {
		l.Warn("marshal problems", zap.Error(err))
	}

	return bytes
}

func getList(storage Repository) [][]byte {
	l.Info("Get list...")

	switch s := storage.(type) {
	case *MemStorage:
		metrics := make([][]byte, s.len)
		i := 0
		s.Mtx.RLock()
		for _, met := range s.items {
			metrics[i] = met
			i++
		}
		s.Mtx.RUnlock()
		return metrics
	default:
		return nil
	}
}

func SetStorage(ots string) {
	l.Info("Set storage ...")
	switch ots {
	case "file":
	default:
		storage = newMemStorage()
	}
}

func getInfo(input []byte) (string, string) {
	mtype := gjson.GetBytes(input, m.Mtype).String()
	name := gjson.GetBytes(input, m.ID).String()

	return mtype, name
}

func dump(path string, rep Repository) error {
	metrics := getList(rep)
	var err error
	for _, metric := range metrics {
		metric = append(metric, byte('\n'))
		err = os.WriteFile(path, metric, 0666)
	}

	return err
}
