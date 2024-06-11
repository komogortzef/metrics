package server

import (
	"context"
	"fmt"
	"sync"

	log "metrics/internal/logger"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DataBase struct {
	*pgxpool.Pool
	Counters map[string][]byte
	Mtx      *sync.RWMutex
}

func (db *DataBase) Put(id string, data []byte, helps ...helper) (int, error) {
	log.Info("DB Put ...")
	helper := helps[0]
	var err error

	query := `INSERT INTO metrics (id, data) VALUES ($1, $2)
	ON CONFLICT (id)
	DO UPDATE SET data = EXCLUDED.data;`

	db.Mtx.RLock()
	oldData, exists := db.Counters[id]
	if !exists && helper != nil {
		if oldData, err = db.recoverCount(id); err != nil {
			db.Counters[id] = data
		} else {
			data, _ = helper(oldData, data)
		}
	}
	db.Mtx.RUnlock()

	if helper != nil && exists {
		data, _ = helper(oldData, data)
		db.Mtx.Lock()
		db.Counters[id] = data
		db.Mtx.Unlock()
	}
	_, err = db.Exec(context.TODO(), query, id, data)

	return 1, err
}

func (db *DataBase) Get(key string) ([]byte, error) {
	log.Info("DB Get...")
	var data []byte
	query := `SELECT data FROM metrics WHERE id=$1`
	err := db.QueryRow(context.TODO(), query, key).Scan(&data)

	return data, err
}

func (db *DataBase) List() (metrics [][]byte, err error) {
	log.Info("DB List...")
	query := `SELECT data FROM metrics`
	db.Mtx.RLock()
	rows, err := db.Query(context.TODO(), query)
	db.Mtx.RUnlock()
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var metric []byte
		if err = rows.Scan(&metric); err != nil {
			log.Fatal("unable to scan row")
		}
		metrics = append(metrics, metric)
	}

	return
}

func (db *DataBase) createTables(ctx context.Context) error {
	query := `
CREATE TABLE IF NOT EXISTS metrics(
	id VARCHAR(255) PRIMARY KEY,
	data JSONB NOT NULL
);`
	if _, err := db.Exec(ctx, query); err != nil {
		return fmt.Errorf("couldn't create gauges table: %w", err)
	}

	log.Info("Success creating tables!")
	return nil
}

func (db *DataBase) recoverCount(id string) ([]byte, error) {
	countData, err := db.Get(id)
	if err != nil {
		return nil, err
	}
	db.Counters[id] = countData

	return countData, err
}
