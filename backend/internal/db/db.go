package db

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
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

	conn, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	conn.SetMaxOpenConns(cfg.Database.MaxConn)
	conn.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	conn.SetConnMaxIdleTime(time.Duration(cfg.Database.ConnMaxIdleTimeSeconds) * time.Second)
	conn.SetConnMaxLifetime(time.Duration(cfg.Database.ConnMaxLifetimeSeconds) * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := conn.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Info("connected to database",
		slog.String("host", cfg.Database.Host),
		slog.String("database", cfg.Database.DBName),
	)

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

func (db *DB) Conn() *sql.DB {
	return db.conn
}

func (db *DB) Ping(ctx context.Context) error {
	return db.conn.PingContext(ctx)
}
