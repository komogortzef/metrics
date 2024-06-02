package server

import (
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
