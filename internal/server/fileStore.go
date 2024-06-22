package server

import (
	ctx "context"
	"fmt"
	"os"
	"time"

	log "metrics/internal/logger"
	s "metrics/internal/service"

	"github.com/pquerna/ffjson/ffjson"
)

type FileStorage struct {
	MemStorage
	FilePath string
	interval int
}

func NewFileStore(path string, interval int) *FileStorage {
	return &FileStorage{
		MemStorage: *NewMemStore(),
		FilePath:   path,
		interval:   interval,
	}
}

func (fs *FileStorage) Put(cx ctx.Context, met *s.Metrics) (*s.Metrics, error) {
	m, _ := fs.MemStorage.Put(cx, met)
	if fs.interval <= 0 {
		if err := fs.dump(cx); err != nil {
			return m, err
		}
	}
	return m, nil
}

func (fs *FileStorage) PutBatch(cx ctx.Context, mets []*s.Metrics) error {
	_ = fs.MemStorage.PutBatch(cx, mets)
	if fs.interval <= 0 {
		if err := fs.dump(cx); err != nil {
			return err
		}
	}
	return nil
}

func (fs *FileStorage) RestoreFromFile(cx ctx.Context) {
	b, err := os.ReadFile(fs.FilePath)
	if err != nil && !os.IsPermission(err) && !os.IsNotExist(err) {
		if err := s.Retry(cx, func() error {
			b, err = os.ReadFile(fs.FilePath)
			return err
		}); err != nil {
			return
		}
	}
	if err != nil {
		return
	}
	var mets []*s.Metrics
	if err := ffjson.Unmarshal(b, &mets); err != nil {
		return
	}
	for _, m := range mets {
		_, _ = fs.MemStorage.Put(cx, m)
	}
	log.Debug("success restore from file!")
}

func (fs *FileStorage) dump(cx ctx.Context) error {
	items, _ := fs.List(cx)
	metBytes, err := ffjson.Marshal(items)
	if err != nil {
		return fmt.Errorf("dump err: %w", err)
	}
	err = os.WriteFile(fs.FilePath, metBytes, 0o666)
	if err != nil && !os.IsPermission(err) {
		err = s.Retry(cx, func() error {
			return os.WriteFile(fs.FilePath, metBytes, 0o666)
		})
	}
	if err != nil {
		return fmt.Errorf("dump error: %w", err)
	}
	log.Debug("success dump!")
	return nil
}

func (fs *FileStorage) dumpWait(cx ctx.Context) {
	if fs.interval <= 0 {
		return
	}
	ticker := time.NewTicker(time.Duration(fs.interval) * time.Second)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := fs.dump(cx); err != nil {
					log.Warn("fs.dumpWithinterval(): Couldn't save data to file")
					return
				}
			case <-cx.Done():
				return
			}
		}
	}()
}
