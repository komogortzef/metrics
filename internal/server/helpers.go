package server

import (
	"io"
	"metrics/internal/logger"
	"os"
	"strconv"
	"sync"
)

func counterAdd(val string, delta string) string {
	numVal, _ := strconv.ParseInt(val, 10, 64)
	numDelta, _ := strconv.ParseInt(delta, 10, 64)
	numVal += numDelta

	return strconv.FormatInt(numVal, 10)
}

func newMemStorage() Repository {
	return &MemStorage{
		Items: make(map[string]string, metricsNumber),
		Mtx:   &sync.RWMutex{},
	}
}

func newFileStorage(path string, resotre bool) (Repository, error) {
	logger.Info("New file storage ...")
	storage := FileStorage{
		Repository: newMemStorage(),
	}

	var file *os.File
	var err error
	if restore {
		file, err = os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0666)
		logger.Info("restore from file")
		restoreFromFile(file, storage.Repository)
	} else {
		file, err = os.OpenFile(path, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
	}

	storage.File = file

	return &storage, err
}

func saveToFile(buf io.Reader, file io.Writer) error {

	return nil
}
