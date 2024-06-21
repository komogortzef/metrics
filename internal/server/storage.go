package server

import (
	ctx "context"
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

func dump(ctx ctx.Context, path string, store Storage) error {
	log.Debug("Dump to file...")
	items, _ := store.List(ctx)
	metBytes, err := ffjson.Marshal(items)
	if err != nil {
		log.Warn("couldn't marshal", zap.Error(err))
	}

	return os.WriteFile(path, metBytes, 0666)
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
