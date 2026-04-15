package types

import (
	"time"

	"github.com/google/uuid"
)

type Vehicle struct {
	ID              int64     `json:"id" db:"id"`                 // BIGINT internal PK (admin display)
	DisplayID       uuid.UUID `json:"display_id" db:"display_id"` // UUID for external/API use
	UserID          uuid.UUID `json:"user_id" db:"user_id"`       // FK → users.display_id
	PlateNumber     string    `json:"plate_number" db:"plate_number"`
	CarModel        string    `json:"car_model" db:"car_model"`
	IsPublic        bool      `json:"is_public" db:"is_public"`
	PhotoURL        *string   `json:"photo_url,omitempty" db:"photo_url"`
	ReviewsEnabled  bool      `json:"reviews_enabled" db:"reviews_enabled"`
	TelegramEnabled bool      `json:"telegram_enabled" db:"telegram_enabled"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

type VehicleCreate struct {
	PlateNumber string `json:"plate_number" validate:"required,min=3,max=20"`
	CarModel    string `json:"car_model" validate:"required,min=2,max=100"`
}

type VehicleUpdate struct {
	PlateNumber     *string `json:"plate_number"`
	CarModel        *string `json:"car_model"`
	PhotoURL        *string `json:"photo_url"`
	ReviewsEnabled  *bool   `json:"reviews_enabled"`
	TelegramEnabled *bool   `json:"telegram_enabled"`
}
