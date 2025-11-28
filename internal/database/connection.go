package database

import (
	"context"
	"fmt"
	"log"
	"order-service/internal/database/db"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	*pgxpool.Pool
	Queries *db.Queries
}

func NewConnection(databaseURL string) (*DB, error) {
	pool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}
	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("✅ Database connected successfully")

	return &DB{
		Pool:    pool,
		Queries: db.New(pool),
	}, nil
}

func (db *DB) Close() {
	db.Pool.Close()
	log.Println("✅ Database connection closed")
}
