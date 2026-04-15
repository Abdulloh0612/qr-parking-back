package db

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var instance *pgxpool.Pool

// GetPoolInstance creates a pgxpool singleton.
// If an instance was already created, it is returned immediately.
func GetPoolInstance(ctx context.Context, databaseURL, maxConnsStr, minConnsStr string, attempts int) *pgxpool.Pool {
	if instance != nil {
		return instance
	}

	maxConns, err := strconv.Atoi(maxConnsStr)
	if err != nil || maxConns <= 0 {
		maxConns = 20
	}
	minConns, err := strconv.Atoi(minConnsStr)
	if err != nil || minConns < 0 {
		minConns = 5
	}

	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		log.Fatalf("parse db config: %v", err)
	}
	cfg.MaxConns = int32(maxConns)
	cfg.MinConns = int32(minConns)

	start := time.Now()
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		if attempts <= 20 {
			time.Sleep(time.Second)
			log.Printf("Retrying database connection, attempt %d: %v", attempts, err)
			return GetPoolInstance(ctx, databaseURL, maxConnsStr, minConnsStr, attempts+1)
		}
		log.Fatalf("Could not establish database connection after %d attempts: %v", attempts, err)
	}

	if err := pool.Ping(ctx); err != nil {
		if attempts <= 20 {
			pool.Close()
			time.Sleep(time.Second)
			log.Printf("Retrying database ping, attempt %d: %v", attempts, err)
			return GetPoolInstance(ctx, databaseURL, maxConnsStr, minConnsStr, attempts+1)
		}
		log.Fatalf("Could not ping database after %d attempts: %v", attempts, err)
	}

	log.Printf("Database connection established -- %dms", time.Since(start).Milliseconds())
	instance = pool
	return instance
}

// NewPool creates a new pgxpool (non-singleton, used when you need explicit lifecycle).
func NewPool(ctx context.Context, databaseURL string, maxConns, minConns int) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse db config: %w", err)
	}

	cfg.MaxConns = int32(maxConns)
	cfg.MinConns = int32(minConns)

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("connect to db: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}

	return pool, nil
}
