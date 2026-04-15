package types

import (
	"time"

	"github.com/google/uuid"
)

type Rating struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	VehicleID uuid.UUID  `json:"vehicle_id" db:"vehicle_id"`
	QRCodeID  *uuid.UUID `json:"qr_code_id,omitempty" db:"qr_code_id"`
	Rating    int        `json:"rating" db:"rating"`
	Comment   *string    `json:"comment,omitempty" db:"comment"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
}
