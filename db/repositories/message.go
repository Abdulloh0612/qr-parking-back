package repositories

import (
	"context"
	"fmt"
	"time"

	"qr-parking/types"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MessageRepo struct {
	pool *pgxpool.Pool
}

func NewMessageRepo(pool *pgxpool.Pool) *MessageRepo {
	return &MessageRepo{pool: pool}
}

const msgSelectCols = `id, qr_code_id, vehicle_id, content, sender_name, sender_phone, is_delivered, delivered_at, created_at`

func scanMessage(dest *types.Message, scan func(...any) error) error {
	return scan(
		&dest.ID, &dest.QRCodeID, &dest.VehicleID,
		&dest.Content, &dest.SenderName, &dest.SenderPhone,
		&dest.IsDelivered, &dest.DeliveredAt, &dest.CreatedAt,
	)
}

func (r *MessageRepo) Create(ctx context.Context, msg *types.Message) error {
	query := `INSERT INTO messages (id, qr_code_id, vehicle_id, content, sender_name, sender_phone)
		VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.pool.Exec(ctx, query, msg.ID, msg.QRCodeID, msg.VehicleID, msg.Content, msg.SenderName, msg.SenderPhone)
	return err
}

func (r *MessageRepo) GetByVehicleID(ctx context.Context, vehicleID uuid.UUID, offset, limit int) ([]types.Message, error) {
	query := `SELECT ` + msgSelectCols + `
		FROM messages WHERE vehicle_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	rows, err := r.pool.Query(ctx, query, vehicleID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []types.Message
	for rows.Next() {
		var m types.Message
		if err := scanMessage(&m, rows.Scan); err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}
	return messages, nil
}

// GetByUserID returns paginated messages for all vehicles owned by the given user.
func (r *MessageRepo) GetByUserID(ctx context.Context, userID uuid.UUID, offset, limit int) ([]types.Message, int, error) {
	var total int
	if err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM messages m
		 JOIN vehicles v ON m.vehicle_id = v.display_id
		 WHERE v.user_id = $1`,
		userID,
	).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count messages by user: %w", err)
	}

	query := `SELECT m.` + msgSelectCols + `
		FROM messages m
		JOIN vehicles v ON m.vehicle_id = v.display_id
		WHERE v.user_id = $1
		ORDER BY m.created_at DESC LIMIT $2 OFFSET $3`
	rows, err := r.pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var messages []types.Message
	for rows.Next() {
		var m types.Message
		if err := scanMessage(&m, rows.Scan); err != nil {
			return nil, 0, err
		}
		messages = append(messages, m)
	}
	return messages, total, nil
}

func (r *MessageRepo) MarkDelivered(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	_, err := r.pool.Exec(ctx, `UPDATE messages SET is_delivered = TRUE, delivered_at = $2 WHERE id = $1`, id, now)
	return err
}

func (r *MessageRepo) CountAll(ctx context.Context) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM messages`).Scan(&count)
	return count, err
}

func (r *MessageRepo) List(ctx context.Context, offset, limit int) ([]types.Message, int, error) {
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM messages`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count messages: %w", err)
	}

	query := `SELECT ` + msgSelectCols + ` FROM messages ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var messages []types.Message
	for rows.Next() {
		var m types.Message
		if err := scanMessage(&m, rows.Scan); err != nil {
			return nil, 0, err
		}
		messages = append(messages, m)
	}
	return messages, total, nil
}
