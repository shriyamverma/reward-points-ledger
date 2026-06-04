package repository

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// InitDBPool handles the lifecycle setup, configuration parsing, and retry logic for the PostgreSQL cluster.
func InitDBPool(dbURL string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		return nil, fmt.Errorf("unable to parse database URL: %w", err)
	}

	// Dynamic connection pooling defaults
	config.MaxConns = 25
	config.MinConns = 5

	var pool *pgxpool.Pool

	for i := 1; i <= 10; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

		pool, err = pgxpool.NewWithConfig(context.Background(), config)
		if err == nil {
			err = pool.Ping(ctx)
		}

		cancel() // Release context resources immediately

		if err == nil {
			slog.Info("Successfully established secure connection pool with PostgreSQL cluster.")
			return pool, nil // Success! Return the active pool
		}

		// Clean up memory leaks if pool was allocated but ping failed
		if pool != nil {
			pool.Close()
		}

		slog.Warn("Database unavailable, retrying in 2s", "attempt", i, "max_attempts", 10, "error", err)
		time.Sleep(2 * time.Second)
	}

	return nil, fmt.Errorf("database pool verification failed after 10 attempts: %w", err)
}
