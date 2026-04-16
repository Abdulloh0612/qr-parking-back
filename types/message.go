package types

import (
	"time"

	"github.com/google/uuid"
)

type Message struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	QRCodeID    *string    `json:"qr_code_id,omitempty" db:"qr_code_id"` // = qr_codes.code (nullable after QR deletion)
	VehicleID   uuid.UUID  `json:"vehicle_id" db:"vehicle_id"`
	Content     string     `json:"content" db:"content"`
	SenderName  *string    `json:"sender_name,omitempty" db:"sender_name"`
	SenderPhone *string    `json:"sender_phone,omitempty" db:"sender_phone"`
	IsDelivered bool       `json:"is_delivered" db:"is_delivered"`
	DeliveredAt *time.Time `json:"delivered_at,omitempty" db:"delivered_at"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
}

type MessageCreate struct {
	Content     string  `json:"content" validate:"required,min=1,max=1000"`
	SenderName  *string `json:"sender_name" validate:"omitempty,max=100"`
	SenderPhone *string `json:"sender_phone" validate:"omitempty,max=20"`
}
