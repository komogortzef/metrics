package server

import (
	ctx "context"
	"os"

	log "metrics/internal/logger"
	s "metrics/internal/service"

	"github.com/pquerna/ffjson/ffjson"
)

type FileStorage struct {
	MemStorage
	filePath string
	syncDump bool
}

func NewFileStore(path string, interval int) *FileStorage {
	sync := interval <= 0
	return &FileStorage{
		MemStorage: *NewMemStore(),
		filePath:   path,
		syncDump:   sync,
	}
}

func (fs *FileStorage) Put(ctx ctx.Context, met *s.Metrics) (*s.Metrics, error) {
	m, _ := fs.MemStorage.Put(ctx, met)
	if fs.syncDump {
		if err := dump(ctx, fs.filePath, fs); err != nil {
			return m, err
		}
	}
	return m, nil
}

func (fs *FileStorage) PutBatch(ctx ctx.Context, mets []*s.Metrics) error {
	_ = fs.MemStorage.PutBatch(ctx, mets)
	if fs.syncDump {
		if err := dump(ctx, fs.filePath, fs); err != nil {
			return err
		}
	}
	return nil
}

func (fs *FileStorage) restoreFromFile(ctx ctx.Context) error {
	log.Debug("Restore from file...")

	b, err := os.ReadFile(fs.filePath)
	if err != nil {
		log.Warn("couldn't read from file. Try three more times...")
		if err = s.Retry(ctx, func() error {
			b, err = os.ReadFile(fs.filePath)
			return err
		}); err != nil {
			return err
		}
	}
	var mets []*s.Metrics
	if err = ffjson.Unmarshal(b, &mets); err != nil {
		return err
	}
	for _, m := range mets {
		_, _ = fs.MemStorage.Put(ctx, m)
	}

	return err
}
