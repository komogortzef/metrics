package server

import (
	ctx "context"
	"fmt"
	"os"

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

func (fs *FileStorage) Put(cx ctx.Context, met *s.Metrics) (*s.Metrics, error) {
	m, _ := fs.MemStorage.Put(cx, met)
	if fs.syncDump {
		if err := dump(cx, fs.filePath, fs); err != nil {
			return m, err
		}
	}
	return m, nil
}

func (fs *FileStorage) PutBatch(cx ctx.Context, mets []*s.Metrics) error {
	_ = fs.MemStorage.PutBatch(cx, mets)
	if fs.syncDump {
		if err := dump(cx, fs.filePath, fs); err != nil {
			return err
		}
	}
	return nil
}

func (fs *FileStorage) restoreFromFile(cx ctx.Context) error {
	b, err := os.ReadFile(fs.filePath)
	if err != nil && !os.IsPermission(err) && !os.IsNotExist(err) {
		if err := s.Retry(cx, func() error {
			b, err = os.ReadFile(fs.filePath)
			return err
		}); err != nil {
			return fmt.Errorf("restoreFromFile failed retry: %w", err)
		}
	}
	if err != nil {
		return fmt.Errorf("restoreFromFile err: %w", err)
	}
	var mets []*s.Metrics
	if err := ffjson.Unmarshal(b, &mets); err != nil {
		return fmt.Errorf("restoreFromFile unmarshal err: %w", err)
	}
	for _, m := range mets {
		_, _ = fs.MemStorage.Put(cx, m)
	}
	return nil
}
