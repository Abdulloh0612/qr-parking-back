package types

import (
	"time"

	"github.com/google/uuid"
)

type QRStatus string

const (
	QRStatusUnregistered QRStatus = "unregistered"
	QRStatusActive       QRStatus = "active"
	QRStatusBlocked      QRStatus = "blocked"
)

type QRCode struct {
	ID           int64      `json:"id" db:"id"`          // BIGINT internal PK
	DisplayID    uuid.UUID  `json:"-" db:"display_id"`   // UUID used only for internal FK relationships
	Code         string     `json:"code" db:"code"`      // External identifier — used in QR URLs and all API calls
	VehicleID    *uuid.UUID `json:"vehicle_id,omitempty" db:"vehicle_id"` // FK → vehicles.display_id
	Status       QRStatus   `json:"status" db:"status"`
	CreatedBy    *uuid.UUID `json:"created_by,omitempty" db:"created_by"` // FK → users.display_id
	RegisteredAt *time.Time `json:"registered_at,omitempty" db:"registered_at"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
}
