package types

import (
	"time"

	"github.com/google/uuid"
)

type TelegramAccount struct {
	ID         uuid.UUID `json:"id" db:"id"`
	UserID     uuid.UUID `json:"user_id" db:"user_id"`
	TgUserID   int64     `json:"tg_user_id" db:"tg_user_id"`
	TgUsername *string   `json:"tg_username,omitempty" db:"tg_username"`
	LinkedAt   time.Time `json:"linked_at" db:"linked_at"`
}
