package server

import (
	"bufio"
	"bytes"
	ctx "context"
	"os"

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
	m, _ := fs.MemStorage.Put(ctx, met)
	if fs.SyncDump {
		if err := dump(ctx, fs.FilePath, fs); err != nil {
			return m, err
		}
	}
	return m, nil
}

func (fs *FileStorage) PutBatch(ctx ctx.Context, mets []s.Metrics) error {
	_ = fs.MemStorage.PutBatch(ctx, mets)
	if fs.SyncDump {
		if err := dump(ctx, fs.FilePath, fs); err != nil {
			return err
		}
	}
	return nil
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
