package server

import (
	"context"
	"fmt"

	log "metrics/internal/logger"
	m "metrics/internal/service"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tidwall/gjson"
)

const (
	counterQuery = "counterQuery"
	gaugeQuery   = "gaugeQuery"
	selectAll    = "selectAll"
	selectMetric = "selectMetric"
)

type DataBase struct {
	*pgxpool.Pool
}

func NewDB(ctx context.Context, addr string) (*DataBase, error) {
	log.Info("DB storage creating")
	config, err := pgxpool.ParseConfig(addr)
	if err != nil {
		return nil, fmt.Errorf("unable to parse connection string: %w", err)
	}
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}
	if err = createTables(ctx, pool); err != nil {
		return nil, err
	}
	if err = prepareQueries(ctx, pool); err != nil {
		return nil, err
	}

	return &DataBase{Pool: pool}, nil
}

func (db *DataBase) Put(ctx context.Context, _ string, data []byte, _ ...helper) error {
	log.Debug("DB Put ...")
	return m.Retry(ctx, func() error {
		conn, err := db.Acquire(ctx)
		if err != nil {
			return err
		}
		defer conn.Release()

		queryName := gaugeQuery
		if gjson.GetBytes(data, "type").String() == m.Counter {
			queryName = counterQuery
		}
		_, err = conn.Exec(ctx, queryName, data)
		return err
	})
}

func (db *DataBase) Get(ctx context.Context, key string) ([]byte, error) {
	log.Info("DB Get...")

	var data []byte
	err := m.Retry(ctx, func() error {
		conn, err := db.Acquire(ctx)
		if err != nil {
			return fmt.Errorf("failed to acquire: %w", err)
		}
		defer conn.Release()

		err = conn.QueryRow(ctx, selectMetric, key).Scan(&data)
		if err != nil {
			return fmt.Errorf("failed to execute query: %w", err)
		}
		return nil
	})

	return data, err
}

func (db *DataBase) List(ctx context.Context) (metrics [][]byte, err error) {
	log.Info("DB List...")
	err = m.Retry(ctx, func() error {
		conn, err := db.Acquire(ctx)
		if err != nil {
			return fmt.Errorf("failed to acquire: %w", err)
		}
		defer conn.Release()

		rows, err := conn.Query(ctx, selectAll)
		if err != nil {
			return err
		}
		for rows.Next() {
			var metric []byte
			if err := rows.Scan(&metric); err != nil {
				return err
			}
			metrics = append(metrics, metric)
		}
		return nil
	})

	return metrics, err
}

func (db *DataBase) insertBatch(ctx context.Context, data []byte) error {
	log.Info("batch sending...")
	return m.Retry(ctx, func() error {
		conn, err := db.Acquire(ctx)
		if err != nil {
			return fmt.Errorf("failed to acquire connection: %w", err)
		}
		defer conn.Release()

		tx, err := conn.Begin(ctx)
		if err != nil {
			return fmt.Errorf("failed transaction beginning: %w", err)
		}
		defer tx.Rollback(ctx)

		batch := &pgx.Batch{}
		gjson.ParseBytes(data).ForEach(func(key, value gjson.Result) bool {
			queryName := gaugeQuery
			if value.Get(m.Mtype).String() == m.Counter {
				queryName = counterQuery
			}
			batch.Queue(queryName, []byte(value.Raw))
			return true
		})

		br := tx.SendBatch(ctx, batch)
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("batch exec failed: %w", err)
		}
		if err := br.Close(); err != nil {
			return fmt.Errorf("failed to close batch res: %w", err)
		}
		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}

		return nil
	})
}

func prepareQueries(ctx context.Context, pool *pgxpool.Pool) error {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("couldn't acquire a connection: %w", err)
	}
	defer conn.Release()

	queries := map[string]string{
		gaugeQuery: `INSERT INTO metrics (data) VALUES ($1)
			           ON CONFLICT ((data->>'id'))
			           DO UPDATE SET data = EXCLUDED.data`,

		counterQuery: `INSERT INTO metrics (data) VALUES ($1) 
			             ON CONFLICT ((data->>'id'))
			             DO UPDATE SET data = jsonb_set(
			             	metrics.data,
			             	'{delta}',
			             	((metrics.data->>'delta')::numeric + (EXCLUDED.data->>'delta')::numeric)::text::jsonb
			             )
			             RETURNING *`,

		selectAll: `SELECT data FROM metrics`,

		selectMetric: `SELECT data FROM metrics WHERE (data->>'id') = $1`,
	}
	for name, query := range queries {
		if _, err = conn.Conn().Prepare(ctx, name, query); err != nil {
			return fmt.Errorf("failed to prepare query %s: %w", name, err)
		}
	}
	return nil
}

func createTables(ctx context.Context, pool *pgxpool.Pool) error {
	log.Debug("Creating tables!")
	query := `
	CREATE TABLE IF NOT EXISTS metrics (
		id SERIAL PRIMARY KEY,
		data JSONB NOT NULL
	);
	CREATE UNIQUE INDEX IF NOT EXISTS unique_metric_id ON metrics ((data->>'id'));
	CREATE UNIQUE INDEX IF NOT EXISTS unique_metric_delta ON metrics ((data->>'delta'));
	`
	if _, err := pool.Exec(ctx, query); err != nil {
		return fmt.Errorf("couldn't create tables: %w", err)
	}

	return nil
}
