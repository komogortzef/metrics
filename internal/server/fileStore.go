package server

import (
	"bufio"
	"bytes"
	"context"
	"os"

	log "metrics/internal/logger"
	"metrics/internal/service"

	"github.com/tidwall/gjson"
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

func (fs *FileStorage) Put(ctx context.Context,
	name string, data []byte, helps ...helper) error {
	err := fs.MemStorage.Put(ctx, name, data, helps...)
	if fs.SyncDump {
		if err = dump(ctx, fs.FilePath, fs); err != nil {
			return err
		}
	}
	return err
}

func (fs *FileStorage) RestoreFromFile(ctx context.Context) error {
	log.Debug("Restore from file...")
	b, err := os.ReadFile(fs.FilePath)
	if err != nil {
		return err
	}
	buff := bytes.NewBuffer(b)
	scanner := bufio.NewScanner(buff)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		bytes := scanner.Bytes()
		name := gjson.GetBytes(bytes, service.ID).String()
		_ = fs.MemStorage.Put(ctx, name, bytes) // ошибка не может здесь возникнуть(addCount не задействован)
	}
	return err
}
