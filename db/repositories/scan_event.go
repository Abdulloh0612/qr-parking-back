package repositories

import (
	"context"
	"fmt"

	"qr-parking/types"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ScanEventRepo struct {
	pool *pgxpool.Pool
}

func NewScanEventRepo(pool *pgxpool.Pool) *ScanEventRepo {
	return &ScanEventRepo{pool: pool}
}

func (r *ScanEventRepo) Create(ctx context.Context, event *types.ScanEvent) error {
	query := `INSERT INTO scan_events (id, qr_code_id, ip_address, user_agent) VALUES ($1, $2, $3, $4)`
	_, err := r.pool.Exec(ctx, query, event.ID, event.QRCodeID, event.IPAddress, event.UserAgent)
	return err
}

func (r *ScanEventRepo) GetByQRCodeID(ctx context.Context, qrCodeID uuid.UUID, offset, limit int) ([]types.ScanEvent, error) {
	query := `SELECT id, qr_code_id, ip_address, user_agent, scanned_at 
		FROM scan_events WHERE qr_code_id = $1 ORDER BY scanned_at DESC LIMIT $2 OFFSET $3`
	rows, err := r.pool.Query(ctx, query, qrCodeID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []types.ScanEvent
	for rows.Next() {
		var e types.ScanEvent
		if err := rows.Scan(&e.ID, &e.QRCodeID, &e.IPAddress, &e.UserAgent, &e.ScannedAt); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, nil
}

func (r *ScanEventRepo) CountByQRCodeID(ctx context.Context, qrCodeID uuid.UUID) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM scan_events WHERE qr_code_id = $1`, qrCodeID).Scan(&count)
	return count, err
}

func (r *ScanEventRepo) CountToday(ctx context.Context) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM scan_events WHERE scanned_at >= CURRENT_DATE`).Scan(&count)
	return count, err
}

func (r *ScanEventRepo) List(ctx context.Context, offset, limit int) ([]types.ScanEvent, int, error) {
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM scan_events`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count scans: %w", err)
	}

	query := `SELECT id, qr_code_id, ip_address, user_agent, scanned_at 
		FROM scan_events ORDER BY scanned_at DESC LIMIT $1 OFFSET $2`
	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var events []types.ScanEvent
	for rows.Next() {
		var e types.ScanEvent
		if err := rows.Scan(&e.ID, &e.QRCodeID, &e.IPAddress, &e.UserAgent, &e.ScannedAt); err != nil {
			return nil, 0, err
		}
		events = append(events, e)
	}
	return events, total, nil
}
