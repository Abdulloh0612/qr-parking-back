package repositories

import (
	"context"
	"errors"

	"qr-parking/types"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RatingRepo struct {
	pool *pgxpool.Pool
}

func NewRatingRepo(pool *pgxpool.Pool) *RatingRepo {
	return &RatingRepo{pool: pool}
}

func (r *RatingRepo) Create(ctx context.Context, rt *types.Rating) error {
	if rt.ID == uuid.Nil {
		rt.ID = uuid.New()
	}
	query := `INSERT INTO ratings (id, vehicle_id, qr_code_id, rating, comment) VALUES ($1, $2, $3, $4, $5)`
	_, err := r.pool.Exec(ctx, query, rt.ID, rt.VehicleID, rt.QRCodeID, rt.Rating, rt.Comment)
	return err
}

func (r *RatingRepo) StatsByVehicleID(ctx context.Context, vehicleID uuid.UUID) (avg float64, count int, err error) {
	query := `SELECT COALESCE(AVG(rating::double precision), 0), COUNT(*)::int FROM ratings WHERE vehicle_id = $1`
	err = r.pool.QueryRow(ctx, query, vehicleID).Scan(&avg, &count)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, 0, nil
	}
	return avg, count, err
}

func (r *RatingRepo) ListByVehicleID(ctx context.Context, vehicleID uuid.UUID, limit int) ([]types.Rating, error) {
	if limit <= 0 || limit > 1000 {
		limit = 100
	}
	query := `SELECT id, vehicle_id, qr_code_id, rating, comment, created_at 
		FROM ratings WHERE vehicle_id = $1 ORDER BY created_at DESC LIMIT $2`
	rows, err := r.pool.Query(ctx, query, vehicleID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []types.Rating
	for rows.Next() {
		var rt types.Rating
		if err := rows.Scan(&rt.ID, &rt.VehicleID, &rt.QRCodeID, &rt.Rating, &rt.Comment, &rt.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, rt)
	}
	return list, nil
}
