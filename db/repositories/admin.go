package repositories

import (
	"context"
	"errors"
	"fmt"

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

const adminSelectCols = `id, display_id, username, password_hash, created_at, updated_at`

func scanAdmin(dest *types.Admin, scan func(...any) error) error {
	return scan(&dest.ID, &dest.DisplayID, &dest.Username, &dest.PasswordHash, &dest.CreatedAt, &dest.UpdatedAt)
}

func (r *AdminRepo) GetByUsername(ctx context.Context, username string) (*types.Admin, error) {
	var a types.Admin
	err := scanAdmin(&a, r.pool.QueryRow(ctx,
		`SELECT `+adminSelectCols+` FROM admins WHERE username = $1`,
		username,
	).Scan)
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
	err := scanAdmin(&a, r.pool.QueryRow(ctx,
		`SELECT `+adminSelectCols+` FROM admins WHERE display_id = $1`,
		displayID,
	).Scan)
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
	err := scanAdmin(&a, r.pool.QueryRow(ctx,
		`INSERT INTO admins (username, password_hash)
		 VALUES ($1, $2)
		 RETURNING `+adminSelectCols,
		username, passwordHash,
	).Scan)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *AdminRepo) List(ctx context.Context, offset, limit int) ([]types.Admin, int, error) {
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM admins`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count admins: %w", err)
	}

	rows, err := r.pool.Query(ctx,
		`SELECT `+adminSelectCols+` FROM admins ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var admins []types.Admin
	for rows.Next() {
		var a types.Admin
		if err := scanAdmin(&a, rows.Scan); err != nil {
			return nil, 0, err
		}
		admins = append(admins, a)
	}
	return admins, total, nil
}

func (r *AdminRepo) Block(ctx context.Context, displayID uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM admins WHERE display_id = $1`, displayID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("admin not found")
	}
	return nil
}
