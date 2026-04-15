package repositories

import (
	"context"
	"errors"

	"qr-parking/types"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type VehicleRepo struct {
	pool *pgxpool.Pool
}

func NewVehicleRepo(pool *pgxpool.Pool) *VehicleRepo {
	return &VehicleRepo{pool: pool}
}

// id=bigint, display_id=uuid, user_id=uuid FK→users.display_id
const vehicleSelectCols = `id, display_id, user_id, plate_number, car_model, is_public, photo_url, reviews_enabled, telegram_enabled, created_at, updated_at`

func scanVehicle(row interface {
	Scan(dest ...any) error
}) (types.Vehicle, error) {
	var v types.Vehicle
	err := row.Scan(
		&v.ID, &v.DisplayID,
		&v.UserID, &v.PlateNumber, &v.CarModel, &v.IsPublic,
		&v.PhotoURL, &v.ReviewsEnabled, &v.TelegramEnabled,
		&v.CreatedAt, &v.UpdatedAt,
	)
	return v, err
}

func (r *VehicleRepo) Create(ctx context.Context, v *types.Vehicle) error {
	// id is BIGSERIAL — auto-generated; display_id is the UUID
	query := `INSERT INTO vehicles (display_id, user_id, plate_number, car_model, is_public, photo_url, reviews_enabled, telegram_enabled)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := r.pool.Exec(ctx, query,
		v.DisplayID, v.UserID, v.PlateNumber, v.CarModel, v.IsPublic,
		v.PhotoURL, v.ReviewsEnabled, v.TelegramEnabled,
	)
	return err
}

// GetByID looks up by UUID display_id (the external-facing identifier).
func (r *VehicleRepo) GetByID(ctx context.Context, displayID uuid.UUID) (*types.Vehicle, error) {
	return r.GetByDisplayID(ctx, displayID)
}

func (r *VehicleRepo) GetByDisplayID(ctx context.Context, displayID uuid.UUID) (*types.Vehicle, error) {
	query := `SELECT ` + vehicleSelectCols + ` FROM vehicles WHERE display_id = $1`
	v, err := scanVehicle(r.pool.QueryRow(ctx, query, displayID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *VehicleRepo) GetByUserID(ctx context.Context, userID uuid.UUID) ([]types.Vehicle, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+vehicleSelectCols+` FROM vehicles WHERE user_id = $1 ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var vehicles []types.Vehicle
	for rows.Next() {
		v, err := scanVehicle(rows)
		if err != nil {
			return nil, err
		}
		vehicles = append(vehicles, v)
	}
	return vehicles, nil
}

// Update looks up by UUID display_id.
func (r *VehicleRepo) Update(ctx context.Context, displayID uuid.UUID, upd types.VehicleUpdate) (*types.Vehicle, error) {
	query := `UPDATE vehicles SET
		plate_number = COALESCE($2, plate_number),
		car_model = COALESCE($3, car_model),
		photo_url = COALESCE($4, photo_url),
		reviews_enabled = COALESCE($5, reviews_enabled),
		telegram_enabled = COALESCE($6, telegram_enabled),
		updated_at = NOW()
		WHERE display_id = $1
		RETURNING ` + vehicleSelectCols
	v, err := scanVehicle(r.pool.QueryRow(ctx, query, displayID, upd.PlateNumber, upd.CarModel, upd.PhotoURL, upd.ReviewsEnabled, upd.TelegramEnabled))
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *VehicleRepo) SetPrivacy(ctx context.Context, displayID uuid.UUID, isPublic bool) error {
	_, err := r.pool.Exec(ctx, `UPDATE vehicles SET is_public = $2, updated_at = NOW() WHERE display_id = $1`, displayID, isPublic)
	return err
}
