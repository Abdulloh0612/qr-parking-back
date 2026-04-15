package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"qr-parking/types"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{pool: pool}
}

// id=bigint, display_id=uuid
const userSelectCols = `id, display_id, phone, first_name, last_name, is_admin, device_id, avatar_url, created_at, updated_at`

func scanUser(row interface {
	Scan(dest ...any) error
}) (types.User, error) {
	var u types.User
	var deviceID sql.NullString
	var avatarURL sql.NullString
	err := row.Scan(
		&u.ID, &u.DisplayID,
		&u.Phone, &u.FirstName, &u.LastName,
		&u.IsAdmin, &deviceID, &avatarURL,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return u, err
	}
	if deviceID.Valid {
		s := deviceID.String
		u.DeviceID = &s
	}
	if avatarURL.Valid {
		s := avatarURL.String
		u.AvatarURL = &s
	}
	return u, nil
}

func (r *UserRepo) Create(ctx context.Context, user *types.User) error {
	// id is BIGSERIAL — auto-generated; display_id is the UUID
	query := `INSERT INTO users (display_id, phone, first_name, last_name, is_admin, device_id, avatar_url)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.pool.Exec(ctx, query,
		user.DisplayID, user.Phone, user.FirstName, user.LastName,
		user.IsAdmin, user.DeviceID, user.AvatarURL,
	)
	return err
}

// GetByID looks up by UUID display_id (the external-facing identifier).
func (r *UserRepo) GetByID(ctx context.Context, displayID uuid.UUID) (*types.User, error) {
	return r.GetByDisplayID(ctx, displayID)
}

func (r *UserRepo) GetByDisplayID(ctx context.Context, displayID uuid.UUID) (*types.User, error) {
	query := `SELECT ` + userSelectCols + ` FROM users WHERE display_id = $1`
	u, err := scanUser(r.pool.QueryRow(ctx, query, displayID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) GetByPhone(ctx context.Context, phone string) (*types.User, error) {
	query := `SELECT ` + userSelectCols + ` FROM users WHERE phone = $1`
	u, err := scanUser(r.pool.QueryRow(ctx, query, phone))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) GetByDeviceID(ctx context.Context, deviceID string) (*types.User, error) {
	query := `SELECT ` + userSelectCols + ` FROM users WHERE device_id = $1`
	u, err := scanUser(r.pool.QueryRow(ctx, query, deviceID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) GetByAdminUsername(ctx context.Context, username string) (*types.User, string, error) {
	query := `SELECT ` + userSelectCols + `, password_hash FROM users WHERE admin_username = $1 AND is_admin = TRUE`
	var u types.User
	var deviceID sql.NullString
	var avatarURL sql.NullString
	var hash sql.NullString
	err := r.pool.QueryRow(ctx, query, username).Scan(
		&u.ID, &u.DisplayID,
		&u.Phone, &u.FirstName, &u.LastName,
		&u.IsAdmin, &deviceID, &avatarURL,
		&u.CreatedAt, &u.UpdatedAt, &hash,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, "", nil
	}
	if err != nil {
		return nil, "", err
	}
	if deviceID.Valid {
		s := deviceID.String
		u.DeviceID = &s
	}
	if avatarURL.Valid {
		s := avatarURL.String
		u.AvatarURL = &s
	}
	if !hash.Valid || hash.String == "" {
		return &u, "", nil
	}
	return &u, hash.String, nil
}

// Update looks up by UUID display_id.
func (r *UserRepo) Update(ctx context.Context, displayID uuid.UUID, upd types.UserUpdate) (*types.User, error) {
	query := `UPDATE users SET
		first_name = COALESCE($2, first_name),
		last_name = COALESCE($3, last_name),
		phone = COALESCE($4, phone),
		avatar_url = COALESCE($5, avatar_url),
		updated_at = NOW()
		WHERE display_id = $1
		RETURNING ` + userSelectCols
	u, err := scanUser(r.pool.QueryRow(ctx, query, displayID, upd.FirstName, upd.LastName, upd.Phone, upd.AvatarURL))
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) List(ctx context.Context, offset, limit int) ([]types.User, int, error) {
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count users: %w", err)
	}

	rows, err := r.pool.Query(ctx,
		`SELECT `+userSelectCols+` FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []types.User
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, 0, err
		}
		users = append(users, u)
	}
	return users, total, nil
}

func (r *UserRepo) Block(ctx context.Context, displayID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM users WHERE display_id = $1`, displayID)
	return err
}
