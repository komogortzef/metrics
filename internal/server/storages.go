package server

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	l "metrics/internal/logger"

	"github.com/jackc/pgx/v5/pgxpool"
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
		Restore  bool
	}

	DataBase struct {
		*pgxpool.Pool
		Addr string
	}
)

func (ms *MemStorage) Put(name string, input []byte, helps ...helper) (int, error) {
	var err error
	ms.Mtx.Lock()
	old, exists := ms.Items[name]
	for _, helper := range helps {
		if exists {
			input, err = helper(old, input)
		}
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

func (ms *MemStorage) Delete(name string) error {
	return nil
}

func (fs *FileStorage) Put(name string, data []byte, helps ...helper) (int, error) {
	len, err := fs.MemStorage.Put(name, data, helps...)

	if fs.Interval == 0 {
		if err = fs.dump(); err != nil {
			return len, err
		}
	}

	return len, err
}

func (ms *MemStorage) listFromMem() [][]byte {
	metrics := make([][]byte, ms.len)
	i := 0
	ms.Mtx.RLock()
	for _, met := range ms.Items {
		metrics[i] = met
		i++
	}
	ms.Mtx.RUnlock()
	return metrics
}

func (fs *FileStorage) dump() error {
	l.Info("Dump starts")
	var buf []byte

	fs.Mtx.RLock()
	for _, data := range fs.Items {
		data = append(data, byte('\n'))
		buf = append(buf, data...)
	}
	fs.Mtx.RUnlock()

	return os.WriteFile(fs.FilePath, buf, 0666)
}

func (fs *FileStorage) startTicker() {
	l.Warn("fs.startTicker()...")
	ticker := time.NewTicker(fs.Interval * time.Second)

	go func() {
		for {
			<-ticker.C
			if err := fs.dump(); err != nil {
				l.Warn("fs.startTicker(): Couldn't save data to file")
				return
			}
		}
	}()
}

func (db *DataBase) Put(name string, data []byte, helps ...helper) (int, error) {
	return 0, nil
}

func (db *DataBase) Get(key string) ([]byte, bool) {
	return nil, false
}

func (db *DataBase) Delete(name string) error {
	return nil
}

func (db *DataBase) Ping(ctx context.Context) error {
	return db.Pool.Ping(ctx)
}

func NewDataBase(addr string) (*DataBase, error) {
	config, err := pgxpool.ParseConfig(addr)
	if err != nil {
		return nil, fmt.Errorf("unable to parse connection string: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	db := &DataBase{
		Pool: pool,
		Addr: addr,
	}

	return db, nil
}
