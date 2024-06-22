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
	SyncDump bool
}

func NewFileStore(path string, interval int) *FileStorage {
	sync := interval <= 0
	return &FileStorage{
		MemStorage: *NewMemStore(),
		FilePath:   path,
		SyncDump:   sync,
	}
}

func (fs *FileStorage) Put(cx ctx.Context, met *s.Metrics) (*s.Metrics, error) {
	m, _ := fs.MemStorage.Put(cx, met)
	if fs.SyncDump {
		if err := fs.dump(cx); err != nil {
			return m, err
		}
	}
	return m, nil
}

func (fs *FileStorage) PutBatch(cx ctx.Context, mets []*s.Metrics) error {
	_ = fs.MemStorage.PutBatch(cx, mets)
	if fs.SyncDump {
		if err := fs.dump(cx); err != nil {
			return err
		}
	}
	return nil
}

func (fs *FileStorage) RestoreFromFile(cx ctx.Context) error {
	b, err := os.ReadFile(fs.FilePath)
	if err != nil && !os.IsPermission(err) && !os.IsNotExist(err) {
		if err := s.Retry(cx, func() error {
			b, err = os.ReadFile(fs.FilePath)
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
	log.Info("success restore from file!")
	return nil
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
	log.Info("success dump!")
	return err
}

func (fs *FileStorage) DumpWait(cx ctx.Context, interval int) {
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
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
