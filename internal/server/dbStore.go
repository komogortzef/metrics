package server

import (
	"context"
	"fmt"

	log "metrics/internal/logger"
	m "metrics/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tidwall/gjson"
	"go.uber.org/zap"
)

type DataBase struct {
	*pgxpool.Pool
}

func (db *DataBase) Put(_ string, data []byte, _ ...helper) (int, error) {
	log.Info("DB Put ...")

	conn, err := db.Acquire(context.TODO())
	if err != nil {
		return 0, fmt.Errorf("failed to acquire: %w", err)
	}
	defer conn.Release()

	if gjson.GetBytes(data, "type").String() == m.Counter {
		_, err = conn.Exec(context.TODO(), "counterQuery", data)
	} else {
		_, err = conn.Exec(context.TODO(), "gaugeQuery", data)
	}

	return 1, err
}

func (db *DataBase) Get(key string) ([]byte, error) {
	log.Info("DB Get...")
	var data []byte

	conn, err := db.Acquire(context.TODO())
	if err != nil {
		return data, fmt.Errorf("failed to acquire: %w", err)
	}
	defer conn.Release()

	if err = conn.QueryRow(
		context.TODO(),
		"selectMetric",
		key).Scan(&data); err != nil {
		return data, fmt.Errorf("failed to execute query: %w", err)
	}
	log.Info("data in db get", zap.String("data", string(data)))

	return data, err
}

func (db *DataBase) List() (metrics [][]byte, err error) {
	log.Info("DB List...")
	conn, err := db.Acquire(context.TODO())
	if err != nil {
		return metrics, fmt.Errorf("failed to acquire: %w", err)
	}
	defer conn.Release()

	rows, err := conn.Query(context.TODO(), "selectAll")
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
	query :=
		`CREATE TABLE IF NOT EXISTS metrics(
	id SERIAL PRIMARY KEY,
	data JSONB NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS unique_metric_id ON metrics ((data->>'id'));
CREATE UNIQUE INDEX IF NOT EXISTS unique_metric_delta ON metrics ((data->>'delta'))`

	if _, err := db.Exec(ctx, query); err != nil {
		return fmt.Errorf("couldn't create tables: %w", err)
	}
	log.Info("Success creating tables!")
	return nil
}

func (db *DataBase) insertBatch(ctx context.Context, data []byte) error {
	log.Info("batch sending...")

	conn, err := db.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("failed to acuire connection: %w", err)
	}
	defer conn.Release()

	tx, err := conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed transaction beginning: %w", err)
	}
	defer tx.Rollback(ctx)

	batch := &pgx.Batch{}

	log.Info("get bytes from insertBatch")
	gjson.ParseBytes(data).ForEach(func(key, value gjson.Result) bool {
		if value.Get(m.Mtype).String() == m.Counter {
			batch.Queue("counterQuery", []byte(value.Raw))
		} else {
			batch.Queue("gaugeQuery", []byte(value.Raw))
		}
		return true
	})

	br := tx.SendBatch(ctx, batch)

	if _, err := br.Exec(); err != nil {
		br.Close()
		return fmt.Errorf("batch exec failed: %w", err)
	}

	if err := br.Close(); err != nil {
		return fmt.Errorf("failed to close batch res: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (db *DataBase) prepareQueries(ctx context.Context) error {
	conn, err := db.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("coulnd't to acquire a connection: %w", err)
	}
	defer conn.Release()

	queries := map[string]string{
		"gaugeQuery": `INSERT INTO metrics (data) VALUES ($1)
	ON CONFLICT ((data->>'id'))
	DO UPDATE SET data = EXCLUDED.data`,

		"counterQuery": `INSERT INTO metrics (data)
VALUES ($1)
ON CONFLICT ((data->>'id')) DO UPDATE
SET data = jsonb_set(
    metrics.data,
    '{delta}',
    ((metrics.data->>'delta')::numeric + (EXCLUDED.data->>'delta')::numeric)::text::jsonb
)
RETURNING *`,

		"selectAll":    `SELECT data FROM metrics`,
		"selectMetric": `SELECT data FROM metrics WHERE (data->>'id') = $1`,
	}

	for name, query := range queries {
		if _, err = conn.Conn().Prepare(ctx, name, query); err != nil {
			log.Info("name query", zap.String("name", name))
			return fmt.Errorf("failed to prepare queries: %w", err)
		}
	}

	return nil
}
