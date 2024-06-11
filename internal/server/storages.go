package server

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	log "metrics/internal/logger"
	m "metrics/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tidwall/gjson"
	"go.uber.org/zap"
)

type (
	MemStorage struct {
		Items map[string][]byte
		len   int
		Mtx   *sync.RWMutex
	}

	FileStorage struct {
		MemStorage
		FilePath string
		Interval time.Duration
	}

	DataBase struct {
		*pgxpool.Pool
	}
)

func (ms *MemStorage) Put(name string, input []byte, help helper) (int, error) {
	var err error
	ms.Mtx.Lock()
	old, exists := ms.Items[name]
	if exists && help != nil {
		input, err = help(old, input)
	}
	ms.Items[name] = input
	if !exists {
		ms.len++
	}
	ms.Mtx.Unlock()

	return ms.len, err
}

func (ms *MemStorage) Get(name string) ([]byte, bool) {
	ms.Mtx.RLock()
	data, ok := ms.Items[name]
	ms.Mtx.RUnlock()

	return data, ok
}

func (ms *MemStorage) List() [][]byte {
	i := 0
	ms.Mtx.RLock()
	metrics := make([][]byte, ms.len)
	for _, met := range ms.Items {
		metrics[i] = met
		i++
	}
	ms.Mtx.RUnlock()
	return metrics
}

func (fs *FileStorage) Put(name string, data []byte, help helper) (int, error) {
	len, err := fs.MemStorage.Put(name, data, help)

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
		len, _ = fs.Put(name, bytes, nil)
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

func (db *DataBase) Put(name string, data []byte, help helper) (int, error) {
	return 0, nil
}

func (db *DataBase) Get(key string) ([]byte, bool) {
	return nil, false
}

func (db *DataBase) List() [][]byte {
	return nil
}

func (db *DataBase) createTables(ctx context.Context) error {
	if err := db.Ping(ctx); err != nil {
		return fmt.Errorf("db connection error: %w", err)
	}

	gaugeTable := `
CREATE TABLE IF NOT EXISTS gauge(
	id VARCHAR(255) PRIMARY KEY,
	value DOUBLE PRECISION NOT NULL
);`

	if _, err := db.Exec(ctx, gaugeTable); err != nil {
		return fmt.Errorf("couldn't create gauges table: %w", err)
	}

	counterTable := `
CREATE TABLE IF NOT EXISTS counter(
	id VARCHAR(255) PRIMARY KEY,
	value BIGINT NOT NULL
);`

	if _, err := db.Exec(ctx, counterTable); err != nil {
		return fmt.Errorf("couldn't create counter table: %w", err)
	}

	log.Info("Success creating tables!")
	return nil
}
