package server

import (
	"bufio"
	"bytes"
	"os"
	"time"

	log "metrics/internal/logger"
	m "metrics/internal/models"

	"github.com/tidwall/gjson"
)

type FileStorage struct {
	MemStorage
	filePath string
	syncDump bool
}

func NewFileStore(path string) *FileStorage {
	return &FileStorage{
		MemStorage: *NewMemStore(),
		filePath:   path,
		syncDump:   true,
	}
}

func (fs *FileStorage) Put(name string, data []byte, helps ...helper) error {
	err := fs.MemStorage.Put(name, data, helps...)
	if fs.syncDump {
		if err = fs.dump(); err != nil {
			return err
		}
	}
	return err
}

func (fs *FileStorage) restoreFromFile() error {
	b, err := os.ReadFile(fs.filePath)
	if err != nil {
		return err
	}
	buff := bytes.NewBuffer(b)
	scanner := bufio.NewScanner(buff)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		bytes := scanner.Bytes()
		name := gjson.GetBytes(bytes, m.ID).String()
		_ = fs.Put(name, bytes) // ошибка не может здесь возникнуть(addCount не задействован)
	}
	return err
}

func (fs *FileStorage) dump() error {
	log.Info("Dump starts")
	var buf []byte
	// объединение всех метрик в один байтовый срез(разделение с помощью '\n'):
	fs.mtx.RLock()
	for _, data := range fs.items {
		data = append(data, byte('\n'))
		buf = append(buf, data...)
	}
	fs.mtx.RUnlock()

	return os.WriteFile(fs.filePath, buf, 0666)
}

func (fs *FileStorage) dumpWithInterval(interval int) {
	log.Info("fs.dumpWithinterval run...")
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	go func() {
		for {
			<-ticker.C
			if err := fs.dump(); err != nil {
				log.Warn("fs.dumpWithinterval(): Couldn't save data to file")
				return
			}
		}
	}()
}
