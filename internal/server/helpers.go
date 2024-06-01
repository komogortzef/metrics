package server

import (
	"metrics/internal/logger"
	m "metrics/internal/models"
	"sync"

	"github.com/tidwall/gjson"
	"go.uber.org/zap"
)

func newMemStorage() Repository {
	return &MemStorage{
		items: make(map[string][]byte, metricsNumber),
		Mtx:   &sync.RWMutex{},
	}
}

// func newFileStorage(path string, resotre bool) (Repository, error) {
// 	logger.Info("New file storage ...")
// 	storage := FileStorage{
// 		Repository: newMemStorage(),
// 	}

// 	var file *os.File
// 	var err error
// 	if restore {
// 		file, err = os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0666)
// 		logger.Info("restore from file")
// 		restoreFromFile(file, storage.Repository)
// 	} else {
// 		file, err = os.OpenFile(path, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
// 	}

// 	storage.File = file

// 	return &storage, err
// }

func addCounter(old []byte, input []byte) (new []byte) {
	logger.Info("addCounter ...")
	var newStruct m.Metrics
	err := newStruct.UnmarshalJSON(old)
	if err != nil {
		logger.Warn("UNMARSHAL problems", zap.Error(err))
	}
	numBytes := gjson.GetBytes(input, m.Delta)
	logger.Info("adding delta...")
	*newStruct.Delta += numBytes.Int()

	if new, err = newStruct.MarshalJSON(); err != nil {
		logger.Warn("marshal problems", zap.Error(err))
	}

	return
}

func getList(storage Repository) []m.Metrics {
	metrics := make([]m.Metrics, 0, metricsNumber)

	switch s := storage.(type) {
	case *MemStorage:
		var metric m.Metrics
		for _, met := range s.items {
			_ = metric.UnmarshalJSON(met)
			metrics = append(metrics, metric)
		}
	}

	return metrics
}

func SetCond(inteval int, path string, restore bool) {
	storeInterval = inteval
	fileStorePath = path
	restore = restore
}

func SetStorage(ots string) {
	logger.Info("Set storage ...")
	switch ots {
	case "file":
		// fileStor, err := newFileStorage(fileStorePath, restore)
		// if err != nil {
		// 	logger.Warn("Problem with file storage", zap.Error(err))
		// }
		// storage = fileStor
	default:
		storage = newMemStorage()
	}
}
