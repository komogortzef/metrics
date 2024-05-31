package server

import (
	"errors"
	"io"
	"os"
	"sync"
	"time"

	"metrics/internal/logger"
	"metrics/internal/models"

	"go.uber.org/zap"
)

var (
	storeInterval int
	fileStorePath string
	restore       bool
)

func SetCond(inteval int, path string, restore bool) {
	storeInterval = inteval
	fileStorePath = path
	restore = restore
}

func SetStorage(ots string) {
	switch ots {
	case "file":
		fileStor, err := newFileStorage(fileStorePath, restore)
		if err != nil {
			logger.Warn("Problem with file storage", zap.Error(err))
		}
		storage = fileStor
	default:
		storage = newMemStorage()
	}
}

type ( // хранилища
	MemStorage struct {
		Items map[string]string
		Mtx   *sync.RWMutex
	}

	FileStorage struct {
		Repository
		File *os.File
	}
)

func (ms *MemStorage) Update(key string, newVal string) error {
	if isCounter(key) { // если тип counter, то суммируем при сохранении
		old, ok := storage.Get(key)
		if ok {
			newVal = counterAdd(old, newVal)
		}
	}

	ms.Mtx.Lock()
	ms.Items[key] = newVal
	ms.Mtx.Unlock()

	logger.Info("value saved")
	return nil
}

func (ms *MemStorage) Get(key string) (string, bool) {
	ms.Mtx.RLock()
	val, ok := ms.Items[key]
	ms.Mtx.RUnlock()

	return val, ok
}

func (fs *FileStorage) Read(bytes []byte) (int, error) {
	logger.Info("Read mem storage starts")

	for _, name := range accounter.list() {
		val, ok := fs.Repository.Get(name)
		if !ok {
			logger.Warn("No data")
			return 0, errors.New("No data")
		}

		metric := models.NewMetric(name, val)
		bytesJson, err := metric.MarshalJSON()
		if err != nil {
			logger.Warn("Couldn't marshal metric")
		}
		bytes = append(bytes, bytesJson...)
		bytes = append(bytes, []byte("\n")...)
	}

	return len(bytes), nil
}

func (fs *FileStorage) Update(key string, newVal string) error { // переопределении для FileStorage
	logger.Info("file store update starts...")

	// по истечении интервала сохраняем в файл
	time.AfterFunc(time.Duration(storeInterval)*time.Second, func() {
		_, err := io.Copy(fs.File, fs)
		if err != nil {
			logger.Warn("Couldn't write to file")
		}
	})

	err := fs.Repository.Update(key, newVal)

	return err
}
