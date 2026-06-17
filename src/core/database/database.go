package database

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPool creates a new PostgreSQL connection pool using pgx.
func NewPool(ctx context.Context, databaseURL string) *pgxpool.Pool {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		log.Fatalf("FATAL: unable to parse database URL: %v", err)
	}

	config.MaxConns = 20
	config.MinConns = 2

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		log.Fatalf("FATAL: unable to create database pool: %v", err)
	}

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("FATAL: unable to ping database: %v", err)
	}

	fmt.Println("✅ Database connection pool established")
	return pool
}
