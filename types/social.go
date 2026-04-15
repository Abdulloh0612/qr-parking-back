package types

import (
	"time"

	"github.com/google/uuid"
)

type SocialPlatform string

const (
	PlatformTelegram  SocialPlatform = "telegram"
	PlatformInstagram SocialPlatform = "instagram"
	PlatformVK        SocialPlatform = "vk"
	PlatformFacebook  SocialPlatform = "facebook"
	PlatformWhatsApp  SocialPlatform = "whatsapp"
)

type SocialProfile struct {
	ID        uuid.UUID      `json:"id" db:"id"`
	UserID    uuid.UUID      `json:"user_id" db:"user_id"`
	Platform  SocialPlatform `json:"platform" db:"platform"`
	Handle    string         `json:"handle" db:"handle"`
	IsPublic  bool           `json:"is_public" db:"is_public"`
	CreatedAt time.Time      `json:"created_at" db:"created_at"`
}

type SocialProfileCreate struct {
	Platform SocialPlatform `json:"platform" validate:"required,oneof=telegram instagram vk facebook whatsapp"`
	Handle   string         `json:"handle" validate:"required,min=1,max=255"`
	IsPublic bool           `json:"is_public"`
}

type SocialProfileUpdate struct {
	Handle   *string `json:"handle" validate:"omitempty,min=1,max=255"`
	IsPublic *bool   `json:"is_public"`
}
