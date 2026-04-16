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
	ID           int64      `json:"id" db:"id"`
	Code         string     `json:"code" db:"code"`
	VehicleID    *uuid.UUID `json:"vehicle_id,omitempty" db:"vehicle_id"`
	Status       QRStatus   `json:"status" db:"status"`
	CreatedBy    *uuid.UUID `json:"created_by,omitempty" db:"created_by"`
	RegisteredAt *time.Time `json:"registered_at,omitempty" db:"registered_at"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
}
