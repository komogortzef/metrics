package conf

import (
	"sync"
	"time"

	m "metrics/internal/models"
	"metrics/internal/server"
)

func (cfg *serverConfig) setStorage(manager *server.MetricsManager) {
	if cfg.DBAddress != "" {
		manager.Store, _ = server.NewDataBase(cfg.DBAddress)
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
