package server

import (
	"bufio"
	"bytes"
	ctx "context"
	"os"
	"time"

	log "metrics/internal/logger"
	s "metrics/internal/service"
)

type FileStorage struct {
	MemStorage
	FilePath string
	SyncDump bool
}

func NewFileStore(path string) *FileStorage {
	return &FileStorage{
		MemStorage: *NewMemStore(),
		FilePath:   path,
		SyncDump:   true,
	}
}

func (fs *FileStorage) Put(ctx ctx.Context, met s.Metrics) (s.Metrics, error) {
	m, err := fs.MemStorage.Put(ctx, met)
	if fs.SyncDump {
		if err = dump(ctx, fs.FilePath, fs); err != nil {
			return m, err
		}
	}
	return m, err
}

func (fs *FileStorage) RestoreFromFile(ctx ctx.Context) error {
	log.Debug("Restore from file...")
	b, err := os.ReadFile(fs.FilePath)
	if err != nil {
		return err
	}
	buff := bytes.NewBuffer(b)
	scanner := bufio.NewScanner(buff)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		var met s.Metrics
		bytes := scanner.Bytes()
		_ = met.UnmarshalJSON(bytes)
		_, _ = fs.MemStorage.Put(ctx, met)
	}
	return err
}

func dump(ctx ctx.Context, path string, store Storage) error {
	log.Debug("Dump to file...")
	var allMetBytes []byte
	var metBytes []byte

	// объединение всех метрик в один байтовый срез(разделение с помощью '\n'):
	items, _ := store.List(ctx)
	for _, metric := range items {
		metBytes, _ = metric.MarshalJSON()
		metBytes = append(metBytes, byte('\n'))
		allMetBytes = append(allMetBytes, metBytes...)
	}

	return os.WriteFile(path, allMetBytes, 0666)
}

func dumpWait(ctx ctx.Context, store Storage, path string, interval int) {
	log.Debug("fs.DumpWait run...")
	ticker := time.NewTicker(time.Duration(interval) * time.Second)

	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := dump(ctx, path, store); err != nil {
					log.Warn("fs.dumpWithinterval(): Couldn't save data to file")
					return
				}
			case <-ctx.Done():
				log.Info("DumpWait end...")
				return
			}
		}
	}()
}
