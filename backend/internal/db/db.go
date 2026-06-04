package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/relay/backend/internal/config"
	"github.com/relay/backend/internal/logger"
)

type DB struct {
	conn *sql.DB
	log  *logger.Logger
}

func New(cfg *config.Config, log *logger.Logger) (*DB, error) {
	dsn := cfg.DSN()

	// Open connection pool
	conn, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	conn.SetMaxOpenConns(cfg.Database.MaxConn)
	conn.SetMaxIdleConns(cfg.Database.MaxConn / 2)
	conn.SetConnMaxLifetime(time.Hour)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := conn.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{
		conn: conn,
		log:  log,
	}, nil
}

func (db *DB) Close() error {
	if db.conn != nil {
		return db.conn.Close()
	}
	return nil
}

func (db *DB) Ping(ctx context.Context) error {
	return db.conn.PingContext(ctx)
}
