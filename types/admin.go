package types

import (
	"time"

	"github.com/google/uuid"
)

type Admin struct {
	ID           int64     `json:"id" db:"id"`
	DisplayID    uuid.UUID `json:"display_id" db:"display_id"`
	Username     string    `json:"username" db:"username"`
	Role         string    `json:"role" db:"role"`
	PasswordHash string    `json:"-" db:"password_hash"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}
