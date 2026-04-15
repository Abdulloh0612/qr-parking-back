package types

import (
	"time"

	"github.com/google/uuid"
)

type ScanEvent struct {
	ID        uuid.UUID `json:"id" db:"id"`
	QRCodeID  uuid.UUID `json:"qr_code_id" db:"qr_code_id"`
	IPAddress string    `json:"ip_address" db:"ip_address"`
	UserAgent string    `json:"user_agent" db:"user_agent"`
	ScannedAt time.Time `json:"scanned_at" db:"scanned_at"`
}
