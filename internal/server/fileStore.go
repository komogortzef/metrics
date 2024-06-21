package server

import (
	ctx "context"
	"os"

	log "metrics/internal/logger"
	s "metrics/internal/service"

	"github.com/pquerna/ffjson/ffjson"
	"go.uber.org/zap"
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

func (fs *FileStorage) Put(ctx ctx.Context, met *s.Metrics) (*s.Metrics, error) {
	m, _ := fs.MemStorage.Put(ctx, met)
	if fs.SyncDump {
		if err := dump(ctx, fs.FilePath, fs); err != nil {
			return m, err
		}
	}
	return m, nil
}

func (fs *FileStorage) PutBatch(ctx ctx.Context, mets []*s.Metrics) error {
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
	var mets []*s.Metrics
	if err = ffjson.Unmarshal(b, &mets); err != nil {
		log.Warn("unmarshal error", zap.Error(err))
	}
	for _, m := range mets {
		_, _ = fs.MemStorage.Put(ctx, m)
	}

	return err
}
