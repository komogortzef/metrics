package server

import (
	"fmt"
	"metrics/internal/logger"
	m "metrics/internal/models"
	"os"
	"sync"
	"time"

	"github.com/tidwall/gjson"
	"go.uber.org/zap"
)

const metricsNumber = 29

var (
	storeInterval int
	fileStorePath string
	restore       bool
)

type (
	MemStorage struct {
		items map[string][]byte
		Mtx   *sync.RWMutex
	}

	FileStorage struct {
		Repository
		file *os.File
	}
)

func (ms *MemStorage) Write(input []byte) (int, error) {
	logger.Info("Mem Write...")
	var newMet m.Metrics
	newMet.UnmarshalJSON(input)
	fmt.Println("passed:", newMet.String())

	mtype, name := getInfo(input)
	ms.Mtx.Lock()
	if mtype == m.Counter {
		logger.Info("is counter...")
		if old, ok := ms.items[name]; ok {
			logger.Info("counter exists")
			new := addCounter(old, input)
			input = input[:0]
			input = new
		}
	}

	ms.items[name] = input
	lenItems := len(ms.items)
	ms.Mtx.Unlock()

	logger.Info("storage length", zap.Int("len", lenItems))
	return lenItems, nil
}

func (ms *MemStorage) Read(output *[]byte) (int, error) {
	logger.Info("Mem Read ...")

	name := gjson.GetBytes(*output, m.Id)

	var err error
	ms.Mtx.RLock()
	data, ok := ms.items[name.String()]
	ms.Mtx.RUnlock()
	if !ok {
		logger.Warn("NO VALUE!!")
		err = m.ErrNoVal
	}

	*output = make([]byte, len(data))
	copy(*output, data)

	return len(*output), err
}

func (fs *FileStorage) Write(output []byte) (int, error) {

	time.AfterFunc(time.Duration(storeInterval)*time.Second, func() {
		list := getList(fs.Repository)
		for _, bytes := range list {
			fs.file.Write(bytes)
		}
	})

	return 0, nil
}
