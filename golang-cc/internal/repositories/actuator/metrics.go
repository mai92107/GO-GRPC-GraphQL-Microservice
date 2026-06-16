package actuator

import "github.com/jackc/pgx/v5/pgxpool"

type PoolStats struct {
	TotalConns        int32
	AcquiredConns     int32
	IdleConns         int32
	ConstructingConns int32
	MaxConns          int32
	AcquireSeconds    float64
}

type PoolStatsRepository struct {
	pool *pgxpool.Pool
}

func NewPoolStatsRepository(pool *pgxpool.Pool) *PoolStatsRepository {
	return &PoolStatsRepository{pool: pool}
}

func (r *PoolStatsRepository) Snapshot() PoolStats {
	stat := r.pool.Stat()
	return PoolStats{
		TotalConns:        stat.TotalConns(),
		AcquiredConns:     stat.AcquiredConns(),
		IdleConns:         stat.IdleConns(),
		ConstructingConns: stat.ConstructingConns(),
		MaxConns:          stat.MaxConns(),
		AcquireSeconds:    stat.AcquireDuration().Seconds(),
	}
}
