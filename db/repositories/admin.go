package repositories

import (
	"context"
	"errors"

	"qr-parking/types"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AdminRepo struct {
	pool *pgxpool.Pool
}

func NewAdminRepo(pool *pgxpool.Pool) *AdminRepo {
	return &AdminRepo{pool: pool}
}

func (r *AdminRepo) GetByUsername(ctx context.Context, username string) (*types.Admin, error) {
	var a types.Admin
	err := r.pool.QueryRow(ctx,
		`SELECT id, display_id, username, password_hash, created_at, updated_at
		 FROM admins WHERE username = $1`,
		username,
	).Scan(&a.ID, &a.DisplayID, &a.Username, &a.PasswordHash, &a.CreatedAt, &a.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *AdminRepo) GetByDisplayID(ctx context.Context, displayID uuid.UUID) (*types.Admin, error) {
	var a types.Admin
	err := r.pool.QueryRow(ctx,
		`SELECT id, display_id, username, password_hash, created_at, updated_at
		 FROM admins WHERE display_id = $1`,
		displayID,
	).Scan(&a.ID, &a.DisplayID, &a.Username, &a.PasswordHash, &a.CreatedAt, &a.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *AdminRepo) Create(ctx context.Context, username, passwordHash string) (*types.Admin, error) {
	var a types.Admin
	err := r.pool.QueryRow(ctx,
		`INSERT INTO admins (username, password_hash)
		 VALUES ($1, $2)
		 RETURNING id, display_id, username, password_hash, created_at, updated_at`,
		username, passwordHash,
	).Scan(&a.ID, &a.DisplayID, &a.Username, &a.PasswordHash, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &a, nil
}
