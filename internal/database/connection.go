package database

import (
	"order-service/internal/database/db"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	*pgxpool.Pool
	Queries *db.Queries
}
