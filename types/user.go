package types

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        int64     `json:"id" db:"id"`             // BIGINT internal PK (admin display)
	DisplayID uuid.UUID `json:"display_id" db:"display_id"` // UUID for external/API use
	Phone     string    `json:"phone" db:"phone"`
	FirstName string    `json:"first_name" db:"first_name"`
	LastName  string    `json:"last_name" db:"last_name"`
	IsAdmin   bool      `json:"is_admin" db:"is_admin"`
	DeviceID  *string   `json:"device_id,omitempty" db:"device_id"`
	AvatarURL *string   `json:"avatar_url,omitempty" db:"avatar_url"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type UserUpdate struct {
	FirstName *string `json:"first_name"`
	LastName  *string `json:"last_name"`
	Phone     *string `json:"phone"`
	AvatarURL *string `json:"avatar_url"`
}
