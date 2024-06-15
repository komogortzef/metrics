package server

import (
	"bufio"
	"bytes"
	"os"
	"time"

	log "metrics/internal/logger"
	m "metrics/internal/models"

	"github.com/tidwall/gjson"
	"go.uber.org/zap"
)

type FileStorage struct {
	MemStorage
	FilePath string
	Interval time.Duration
}

func (fs *FileStorage) Put(name string, data []byte, helps ...helper) (int, error) {
	len, err := fs.MemStorage.Put(name, data, helps...)
	if fs.Interval == 0 { // синхронная запись в файл при поступлении данных
		if err = fs.dump(); err != nil {
			return len, err
		}
	}
	return len, err
}

func (fs *FileStorage) restoreFromFile() (len int) {
	b, err := os.ReadFile(fs.FilePath)
	if err != nil {
		log.Warn("No file to restore!")
		return
	}
	buff := bytes.NewBuffer(b)
	scanner := bufio.NewScanner(buff)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		bytes := scanner.Bytes()
		name := gjson.GetBytes(bytes, m.ID).String()
		len, _ = fs.Put(name, bytes)
	}
	log.Info("number of metrics recovered from the file", zap.Int("len", len))
	return
}

func (fs *FileStorage) dump() error {
	log.Info("Dump starts")
	var buf []byte
	// объединение всех метрик в один байтовый срез(разделение с помощью '\n'):
	fs.Mtx.RLock()
	for _, data := range fs.Items {
		data = append(data, byte('\n'))
		buf = append(buf, data...)
	}
	fs.Mtx.RUnlock()

	return os.WriteFile(fs.FilePath, buf, 0666)
}

func (fs *FileStorage) dumpWithInterval() {
	log.Info("fs.dumpWithInterval run...")
	ticker := time.NewTicker(fs.Interval * time.Second)

	go func() {
		for {
			<-ticker.C
			if err := fs.dump(); err != nil {
				log.Warn("fs.dumpWithInterval(): Couldn't save data to file")
				return
			}
		}
	}()
}
