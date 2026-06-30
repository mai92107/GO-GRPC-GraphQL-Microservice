package public

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"gorm.io/gorm"
)

type Repository struct {
	pool *pgxpool.Pool
	db   *gorm.DB
}

func New(pool *pgxpool.Pool, db *gorm.DB) *Repository {
	return &Repository{pool: pool, db: db}
}
