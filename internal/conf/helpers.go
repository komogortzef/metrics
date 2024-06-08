package conf

import (
	"sync"
	"time"

	m "metrics/internal/models"
	"metrics/internal/server"

	"github.com/jackc/pgx/v5/pgxpool"
)

func (cfg *serverConfig) setStorage(manager *server.MetricsManager) {
	if cfg.dbAddress != "" {
		manager.Store = &server.DataBase{
			Pool: &pgxpool.Pool{},
			Addr: cfg.dbAddress,
		}
	} else if cfg.FileStoragePath != "" {
		manager.Store = &server.FileStorage{
			MemStorage: server.MemStorage{
				Items: make(map[string][]byte, m.MetricsNumber),
				Mtx:   &sync.RWMutex{},
			},
			FilePath: cfg.FileStoragePath,
			Interval: time.Duration(cfg.StoreInterval),
			Restore:  cfg.Restore,
		}
	} else {
		manager.Store = &server.MemStorage{
			Items: make(map[string][]byte, m.MetricsNumber),
			Mtx:   &sync.RWMutex{},
		}
	}
}
