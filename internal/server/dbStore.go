package server

import (
	ctx "context"
	"fmt"

	log "metrics/internal/logger"
	s "metrics/internal/service"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type dbOperation uint8

const (
	insertMetric dbOperation = iota
	selectMetric
)
const (
	insertCounter = "insertCounter"
	insertGauge   = "insertGauge"
	selectGauge   = "selectGauge"
	selectCounter = "selectCounter"
	selectAll     = "selectAll"
)

type DataBase struct {
	*pgxpool.Pool
}

func NewDB(ctx ctx.Context, addr string) (*DataBase, error) {
	log.Debug("DB Put ...")
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

func (db *DataBase) Put(ctx ctx.Context, m s.Metrics) (s.Metrics, error) {
	log.Debug("DB Put ...")
	err := s.Retry(ctx, func() error {
		conn, err := db.Acquire(ctx)
		if err != nil {
			return err
		}
		defer conn.Release()

		var val any
		query := getQuery(insertMetric, m.MType)
		err = conn.QueryRow(ctx, query, m.ToSlice()...).Scan(&val)
		if v, ok := val.(int64); ok {
			m.Delta = &v
		} else {
			v, _ := val.(float64)
			m.Value = &v
		}
		return err
	})

	return m, err
}

func (db *DataBase) Get(ctx ctx.Context, m s.Metrics) (s.Metrics, error) {
	log.Debug("DB Get...")
	err := s.Retry(ctx, func() error {
		conn, err := db.Acquire(ctx)
		if err != nil {
			return fmt.Errorf("failed to acquire: %w", err)
		}
		defer conn.Release()

		query := getQuery(selectMetric, m.MType)
		var val any
		if err = conn.QueryRow(ctx, query, m.ID).Scan(&val); err != nil {
			return fmt.Errorf("failed to execute query: %w", err)
		}
		if v, ok := val.(int64); ok {
			m.Delta = &v
		} else {
			v, _ := val.(float64)
			m.Value = &v
		}
		return nil
	})

	return m, err
}

func (db *DataBase) List(ctx ctx.Context) (metrics []s.Metrics, err error) {
	log.Debug("DB List...")
	err = s.Retry(ctx, func() error {
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
			var met s.Metrics
			var val any
			err := rows.Scan(&met.ID, &val)
			if err != nil {
				return err
			}
			if v, ok := val.(int64); ok {
				met.Delta = &v
			} else {
				v, _ := val.(float64)
				met.Value = &v
			}
			metrics = append(metrics, met)
		}
		return nil
	})
	return metrics, err
}

func (db *DataBase) PutBatch(ctx ctx.Context, mets []s.Metrics) error {
	log.Debug("Batch sending...")
	return s.Retry(ctx, func() error {
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
		for _, m := range mets {
			batch.Queue(getQuery(insertMetric, m.MType), m.ToSlice()...)
		}
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

func (db *DataBase) Ping(ctx ctx.Context) error {
	if err := db.Pool.Ping(ctx); err != nil {
		log.Warn("PingHandler(): There is no connection to data base")
		return err
	}
	log.Info("DataBase storage is working")
	return nil
}

func (db *DataBase) Close() {
	db.Pool.Close()
	log.Info("DataBase connection is closed")
}

func prepareQueries(ctx ctx.Context, pool *pgxpool.Pool) error {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("couldn't acquire a connection: %w", err)
	}
	defer conn.Release()

	queries := map[string]string{
		insertGauge: `INSERT INTO gauge(id, value) VALUES($1, $2) 
			          ON CONFLICT(id) 
				      DO UPDATE SET value = EXCLUDED.value
				      RETURNING value`,

		insertCounter: `INSERT INTO counter(id, value) VALUES($1, $2) 
			            ON CONFLICT(id) 
				        DO UPDATE SET value = counter.value + excluded.value
				        RETURNING value`,

		selectGauge: `SELECT value FROM gauge WHERE id = $1`,

		selectCounter: `SELECT value FROM counter WHERE id = $1`,

		selectAll: `SELECT id, value FROM gauge
			        UNION ALL
			        SELECT id, value FROM counter;`,
	}
	for name, query := range queries {
		if _, err = conn.Conn().Prepare(ctx, name, query); err != nil {
			return fmt.Errorf("failed to prepare query %s: %w", name, err)
		}
	}
	return nil
}

func createTables(ctx ctx.Context, pool *pgxpool.Pool) error {
	log.Debug("Creating tables!")
	query := `
	CREATE TABLE IF NOT EXISTS counter(
		id VARCHAR(255) PRIMARY KEY,
		value BIGINT NOT NULL
	);
	CREATE TABLE IF NOT EXISTS gauge(
	   id VARCHAR(255) PRIMARY KEY,
	   value DOUBLE PRECISION NOT NULL
	);`
	if _, err := pool.Exec(ctx, query); err != nil {
		return fmt.Errorf("couldn't create tables: %w", err)
	}

	return nil
}
