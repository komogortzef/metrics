package server

import (
	ctx "context"
	"fmt"
	"os"
	"time"

	log "metrics/internal/logger"
	s "metrics/internal/service"

	"github.com/pquerna/ffjson/ffjson"
	"go.uber.org/zap"
)

type Storage interface {
	Put(ctx.Context, *s.Metrics) (*s.Metrics, error)
	Get(ctx.Context, *s.Metrics) (*s.Metrics, error)
	List(ctx.Context) ([]*s.Metrics, error)
	PutBatch(ctx.Context, []*s.Metrics) error
	Ping(ctx.Context) error
	Close()
}

func NewStorage(cx ctx.Context, m *MetricManager) (err error) {
	switch {
	case m.DBAddress != "":
		if m.Store, err = NewDB(cx, m.DBAddress); err != nil {
			return err
		}
		m.FileStoragePath = s.NoStorage
	case m.FileStoragePath != "":
		store := NewFileStore(m.FileStoragePath, m.StoreInterval)
		if m.Restore {
			if err := store.restoreFromFile(cx); err != nil {
				log.Warn("restore from file error", zap.Error(err))
			}
		}
		if !store.syncDump {
			dumpWait(cx, store, m.FileStoragePath, m.StoreInterval)
		}
		m.Store = store
	default:
		m.Store = NewMemStore()
	}
	return nil
}

func dump(cx ctx.Context, path string, store Storage) error {
	items, _ := store.List(cx)
	metBytes, err := ffjson.Marshal(items)
	if err != nil {
		return fmt.Errorf("dump err: %w", err)
	}
	err = os.WriteFile(path, metBytes, 0o666)
	if err != nil && !os.IsPermission(err) {
		log.Warn("couldn't write to file. Try three more times...")
		err = s.Retry(cx, func() error {
			return os.WriteFile(path, metBytes, 0o666)
		})
	}
	return err
}

func dumpWait(cx ctx.Context, store Storage, path string, interval int) {
	log.Debug("fs.DumpWait run...")
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := dump(cx, path, store); err != nil {
					log.Warn("fs.dumpWithinterval(): Couldn't save data to file")
					return
				}
			case <-cx.Done():
				log.Info("DumpWait is stopped")
				return
			}
		}
	}()
}
