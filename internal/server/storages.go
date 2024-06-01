package server

import (
	"fmt"
	"metrics/internal/logger"
	m "metrics/internal/models"
	"os"
	"sync"

	"github.com/tidwall/gjson"
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
		File *os.File
	}
)

func (ms *MemStorage) Write(input []byte) (int, error) {
	logger.Info("Mem Write...")
	typeBytes := gjson.GetBytes(input, m.Mtype)
	fmt.Println("type:", typeBytes.String())
	nameBytes := gjson.GetBytes(input, m.Id)
	name := nameBytes.String()

	ms.Mtx.Lock()
	if typeBytes.String() == m.Counter {
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

	fmt.Println("len storage:", lenItems)
	return lenItems, nil
}

func (ms *MemStorage) Read(output []byte) (int, error) {

	return 0, nil
}

func (ms *MemStorage) Put(name string, input []byte) error {
	return nil
}

func (ms *MemStorage) Get(name string) ([]byte, error) {

	return []byte("ok"), nil
}
