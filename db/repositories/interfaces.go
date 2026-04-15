package repositories

import (
	"context"

	"qr-parking/types"

	"github.com/google/uuid"
)

type AdminRepository interface {
	GetByUsername(ctx context.Context, username string) (*types.Admin, error)
	GetByDisplayID(ctx context.Context, displayID uuid.UUID) (*types.Admin, error)
	Create(ctx context.Context, username, passwordHash string) (*types.Admin, error)
}

type UserRepository interface {
	Create(ctx context.Context, user *types.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*types.User, error)
	GetByDisplayID(ctx context.Context, displayID uuid.UUID) (*types.User, error)
	GetByPhone(ctx context.Context, phone string) (*types.User, error)
	GetByDeviceID(ctx context.Context, deviceID string) (*types.User, error)
	GetByAdminUsername(ctx context.Context, username string) (*types.User, string, error)
	Update(ctx context.Context, id uuid.UUID, upd types.UserUpdate) (*types.User, error)
	List(ctx context.Context, offset, limit int) ([]types.User, int, error)
	Block(ctx context.Context, id uuid.UUID) error
}

type VehicleRepository interface {
	Create(ctx context.Context, vehicle *types.Vehicle) error
	GetByID(ctx context.Context, id uuid.UUID) (*types.Vehicle, error)
	GetByDisplayID(ctx context.Context, displayID uuid.UUID) (*types.Vehicle, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]types.Vehicle, error)
	Update(ctx context.Context, id uuid.UUID, upd types.VehicleUpdate) (*types.Vehicle, error)
	SetPrivacy(ctx context.Context, id uuid.UUID, isPublic bool) error
}

type QRCodeRepository interface {
	Create(ctx context.Context, qr *types.QRCode) error
	// GetByCode looks up exactly by the code field.
	GetByCode(ctx context.Context, code string) (*types.QRCode, error)
	// GetByCodeOrID tries code first; if not found, tries the numeric bigint id.
	GetByCodeOrID(ctx context.Context, qrRef string) (*types.QRCode, error)
	GetByVehicleID(ctx context.Context, vehicleID uuid.UUID) ([]types.QRCode, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]types.QRCode, error)
	Register(ctx context.Context, code string, vehicleID uuid.UUID) error
	// Block accepts the QR code string or numeric id string.
	Block(ctx context.Context, qrRef string) error
	List(ctx context.Context, status string, offset, limit int) ([]types.QRCode, int, error)
	CountByStatus(ctx context.Context) (map[string]int, error)
}

type ScanEventRepository interface {
	Create(ctx context.Context, event *types.ScanEvent) error
	GetByQRCodeID(ctx context.Context, qrCodeID uuid.UUID, offset, limit int) ([]types.ScanEvent, error)
	CountByQRCodeID(ctx context.Context, qrCodeID uuid.UUID) (int, error)
	CountToday(ctx context.Context) (int, error)
	List(ctx context.Context, offset, limit int) ([]types.ScanEvent, int, error)
}

type MessageRepository interface {
	Create(ctx context.Context, msg *types.Message) error
	GetByVehicleID(ctx context.Context, vehicleID uuid.UUID, offset, limit int) ([]types.Message, error)
	MarkDelivered(ctx context.Context, id uuid.UUID) error
	CountAll(ctx context.Context) (int, error)
	List(ctx context.Context, offset, limit int) ([]types.Message, int, error)
}

type SocialProfileRepository interface {
	Create(ctx context.Context, sp *types.SocialProfile) error
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]types.SocialProfile, error)
	GetPublicByUserID(ctx context.Context, userID uuid.UUID) ([]types.SocialProfile, error)
	GetByUserIDAndPlatform(ctx context.Context, userID uuid.UUID, platform types.SocialPlatform) (*types.SocialProfile, error)
	UpsertByPlatform(ctx context.Context, userID uuid.UUID, platform types.SocialPlatform, handle string, isPublic bool) error
	Update(ctx context.Context, id uuid.UUID, upd types.SocialProfileUpdate) (*types.SocialProfile, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type RatingRepository interface {
	Create(ctx context.Context, r *types.Rating) error
	StatsByVehicleID(ctx context.Context, vehicleID uuid.UUID) (avg float64, count int, err error)
	ListByVehicleID(ctx context.Context, vehicleID uuid.UUID, limit int) ([]types.Rating, error)
}

type TelegramRepository interface {
	Create(ctx context.Context, acc *types.TelegramAccount) error
	GetByUserID(ctx context.Context, userID uuid.UUID) (*types.TelegramAccount, error)
	GetByTgUserID(ctx context.Context, tgUserID int64) (*types.TelegramAccount, error)
	Delete(ctx context.Context, userID uuid.UUID) error
}
