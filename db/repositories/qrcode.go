package repositories

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"qr-parking/types"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type QRCodeRepo struct {
	pool *pgxpool.Pool
}

func NewQRCodeRepo(pool *pgxpool.Pool) *QRCodeRepo {
	return &QRCodeRepo{pool: pool}
}

// id=bigint, display_id=uuid, vehicle_id=uuid FK→vehicles.display_id, created_by=uuid FK→users.display_id
const qrSelectCols = `id, display_id, code, vehicle_id, status, created_by, registered_at, created_at`

func scanQR(row interface {
	Scan(dest ...any) error
}) (types.QRCode, error) {
	var qr types.QRCode
	err := row.Scan(
		&qr.ID, &qr.DisplayID,
		&qr.Code, &qr.VehicleID, &qr.Status,
		&qr.CreatedBy, &qr.RegisteredAt, &qr.CreatedAt,
	)
	return qr, err
}

func (r *QRCodeRepo) Create(ctx context.Context, qr *types.QRCode) error {
	// id is BIGSERIAL — auto-generated; display_id is the UUID
	query := `INSERT INTO qr_codes (display_id, code, status, created_by) VALUES ($1, $2, $3, $4)`
	_, err := r.pool.Exec(ctx, query, qr.DisplayID, qr.Code, qr.Status, qr.CreatedBy)
	return err
}

func (r *QRCodeRepo) GetByCode(ctx context.Context, code string) (*types.QRCode, error) {
	qr, err := scanQR(r.pool.QueryRow(ctx, `SELECT `+qrSelectCols+` FROM qr_codes WHERE code = $1`, code))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &qr, nil
}

// GetByCodeOrID tries the code field first; if not found, parses qrRef as a
// bigint and falls back to looking up by the internal numeric id.
func (r *QRCodeRepo) GetByCodeOrID(ctx context.Context, qrRef string) (*types.QRCode, error) {
	qr, err := r.GetByCode(ctx, qrRef)
	if err != nil {
		return nil, err
	}
	if qr != nil {
		return qr, nil
	}
	numID, parseErr := strconv.ParseInt(qrRef, 10, 64)
	if parseErr != nil {
		return nil, nil
	}
	return r.getByNumericID(ctx, numID)
}

func (r *QRCodeRepo) getByNumericID(ctx context.Context, id int64) (*types.QRCode, error) {
	qr, err := scanQR(r.pool.QueryRow(ctx, `SELECT `+qrSelectCols+` FROM qr_codes WHERE id = $1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &qr, nil
}

// getByDisplayID is used internally for FK-related operations (e.g. scan events).
func (r *QRCodeRepo) getByDisplayID(ctx context.Context, displayID uuid.UUID) (*types.QRCode, error) {
	qr, err := scanQR(r.pool.QueryRow(ctx, `SELECT `+qrSelectCols+` FROM qr_codes WHERE display_id = $1`, displayID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &qr, nil
}

func (r *QRCodeRepo) GetByVehicleID(ctx context.Context, vehicleDisplayID uuid.UUID) ([]types.QRCode, error) {
	rows, err := r.pool.Query(ctx, `SELECT `+qrSelectCols+` FROM qr_codes WHERE vehicle_id = $1`, vehicleDisplayID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var codes []types.QRCode
	for rows.Next() {
		qr, err := scanQR(rows)
		if err != nil {
			return nil, err
		}
		codes = append(codes, qr)
	}
	return codes, nil
}

func (r *QRCodeRepo) GetByUserID(ctx context.Context, userDisplayID uuid.UUID) ([]types.QRCode, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT q.`+qrSelectCols+` FROM qr_codes q
		 JOIN vehicles v ON q.vehicle_id = v.display_id
		 WHERE v.user_id = $1`,
		userDisplayID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var codes []types.QRCode
	for rows.Next() {
		qr, err := scanQR(rows)
		if err != nil {
			return nil, err
		}
		codes = append(codes, qr)
	}
	return codes, nil
}

// Register sets a QR code as active and links it to a vehicle (by vehicle display_id UUID).
func (r *QRCodeRepo) Register(ctx context.Context, code string, vehicleDisplayID uuid.UUID) error {
	now := time.Now()
	tag, err := r.pool.Exec(ctx,
		`UPDATE qr_codes SET vehicle_id = $2, status = 'active', registered_at = $3 WHERE code = $1 AND status = 'unregistered'`,
		code, vehicleDisplayID, now,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("qr code not found or already registered")
	}
	return nil
}

// Block blocks a QR code by code string or numeric id string.
func (r *QRCodeRepo) Block(ctx context.Context, qrRef string) error {
	qr, err := r.GetByCodeOrID(ctx, qrRef)
	if err != nil {
		return err
	}
	if qr == nil {
		return fmt.Errorf("qr code not found")
	}
	_, err = r.pool.Exec(ctx, `UPDATE qr_codes SET status = 'blocked' WHERE id = $1`, qr.ID)
	return err
}

func (r *QRCodeRepo) List(ctx context.Context, status string, offset, limit int) ([]types.QRCode, int, error) {
	var total int
	if status != "" {
		if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM qr_codes WHERE status = $1`, status).Scan(&total); err != nil {
			return nil, 0, err
		}
	} else {
		if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM qr_codes`).Scan(&total); err != nil {
			return nil, 0, err
		}
	}

	var rows interface {
		Next() bool
		Scan(...any) error
		Close()
	}
	var err error
	if status != "" {
		rows, err = r.pool.Query(ctx,
			`SELECT `+qrSelectCols+` FROM qr_codes WHERE status = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
			status, limit, offset,
		)
	} else {
		rows, err = r.pool.Query(ctx,
			`SELECT `+qrSelectCols+` FROM qr_codes ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
			limit, offset,
		)
	}
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var codes []types.QRCode
	for rows.Next() {
		qr, e := scanQR(rows)
		if e != nil {
			return nil, 0, e
		}
		codes = append(codes, qr)
	}
	return codes, total, nil
}

func (r *QRCodeRepo) CountByStatus(ctx context.Context) (map[string]int, error) {
	rows, err := r.pool.Query(ctx, `SELECT status, COUNT(*) FROM qr_codes GROUP BY status`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counts := make(map[string]int)
	for rows.Next() {
		var st string
		var count int
		if err := rows.Scan(&st, &count); err != nil {
			return nil, err
		}
		counts[st] = count
	}
	return counts, nil
}
