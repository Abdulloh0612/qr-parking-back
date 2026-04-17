package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"qr-parking/db/repositories"
	"qr-parking/types"

	"github.com/google/uuid"
)

// ─── Input types ──────────────────────────────────────────────────────────────

// QRRegisterInput contains the minimum data needed to claim a QR code.
type QRRegisterInput struct {
	QRID  string
	Phone string
}

// OwnerProfileInput is used to update user-level fields (name, socials).
// Nil pointer = "not provided, leave as-is".
// Pointer to empty string = "delete this value".
type OwnerProfileInput struct {
	FirstName *string
	LastName  *string
	AvatarURL *string
	// Socials: nil = skip, "" = delete, "value" = upsert
	WhatsApp  *string
	Instagram *string
	Telegram  *string
	VK        *string
	Facebook  *string
}

// ─── Output types ─────────────────────────────────────────────────────────────

// QROwnerOut is the public profile shown when anyone scans a registered QR.
type QROwnerOut struct {
	DisplayID         uuid.UUID `json:"id"`
	Phone             string    `json:"phone"`
	FirstName         string    `json:"first_name"`
	LastName          string    `json:"last_name"`
	AvatarURL         *string   `json:"avatar_url,omitempty"`
	VehicleNumber     string    `json:"vehicle_number"`
	VehicleBrand      string    `json:"vehicle_brand"`
	VehiclePhotoURL   *string   `json:"vehicle_photo_url,omitempty"`
	WhatsApp          *string   `json:"whatsapp,omitempty"`
	Instagram         *string   `json:"instagram,omitempty"`
	Telegram          *string   `json:"telegram,omitempty"`
	VK                *string   `json:"vk,omitempty"`
	Facebook          *string   `json:"facebook,omitempty"`
	TelegramEnabled   bool      `json:"telegram_enabled"`
	TelegramBotLinked bool      `json:"telegram_bot_linked"`
	ScanCount         int       `json:"scan_count"`
}

// QRInfoResult wraps the GetQRInfo response.
type QRInfoResult struct {
	Registered bool        `json:"registered"`
	Owner      *QROwnerOut `json:"owner,omitempty"`
}

// OwnerVehicleOut is a vehicle entry in the owner's private profile.
type OwnerVehicleOut struct {
	ID              uuid.UUID `json:"id"`
	PlateNumber     string    `json:"plate_number"`
	CarModel        string    `json:"car_model"`
	PhotoURL        *string   `json:"photo_url,omitempty"`
	TelegramEnabled bool      `json:"telegram_enabled"`
	QRCode          string    `json:"qr_code,omitempty"`
}

// OwnerDataOut is the full private profile returned to the authenticated owner.
type OwnerDataOut struct {
	ID        uuid.UUID         `json:"id"`
	Phone     string            `json:"phone"`
	FirstName string            `json:"first_name"`
	LastName  string            `json:"last_name"`
	AvatarURL *string           `json:"avatar_url,omitempty"`
	Vehicles  []OwnerVehicleOut `json:"vehicles"`
	WhatsApp  *string           `json:"whatsapp,omitempty"`
	Instagram *string           `json:"instagram,omitempty"`
	Telegram  *string           `json:"telegram,omitempty"`
	VK        *string           `json:"vk,omitempty"`
	Facebook  *string           `json:"facebook,omitempty"`
}

// ─── Service ──────────────────────────────────────────────────────────────────

type APISpecService struct {
	userRepo    repositories.UserRepository
	vehicleRepo repositories.VehicleRepository
	qrRepo      repositories.QRCodeRepository
	socialRepo  repositories.SocialProfileRepository
	scanRepo    repositories.ScanEventRepository
	tgRepo      repositories.TelegramRepository
}

func NewAPISpecService(
	userRepo repositories.UserRepository,
	vehicleRepo repositories.VehicleRepository,
	qrRepo repositories.QRCodeRepository,
	socialRepo repositories.SocialProfileRepository,
	scanRepo repositories.ScanEventRepository,
	tgRepo repositories.TelegramRepository,
) *APISpecService {
	return &APISpecService{
		userRepo:    userRepo,
		vehicleRepo: vehicleRepo,
		qrRepo:      qrRepo,
		socialRepo:  socialRepo,
		scanRepo:    scanRepo,
		tgRepo:      tgRepo,
	}
}

// ─── Public scan ──────────────────────────────────────────────────────────────

// GetQRInfo resolves a QR code for anonymous scanners.
// Returns {registered:false} if unregistered; full owner profile if active.
// A scan event is recorded for active QRs.
func (s *APISpecService) GetQRInfo(ctx context.Context, qrRef, ip, userAgent string) (*QRInfoResult, error) {
	qr, err := s.qrRepo.GetByCodeOrID(ctx, qrRef)
	if err != nil {
		return nil, fmt.Errorf("lookup qr: %w", err)
	}
	if qr == nil || qr.Status == types.QRStatusBlocked {
		return nil, errNotFound("QR код не найден")
	}
	if qr.Status != types.QRStatusActive || qr.VehicleID == nil {
		return &QRInfoResult{Registered: false}, nil
	}

	_ = s.scanRepo.Create(ctx, &types.ScanEvent{
		ID:        uuid.New(),
		QRCodeID:  qr.Code,
		IPAddress: ip,
		UserAgent: userAgent,
	})

	veh, err := s.vehicleRepo.GetByID(ctx, *qr.VehicleID)
	if err != nil {
		return nil, fmt.Errorf("lookup vehicle: %w", err)
	}
	if veh == nil {
		return nil, errNotFound("QR код не найден")
	}

	user, err := s.userRepo.GetByID(ctx, veh.UserID)
	if err != nil {
		return nil, fmt.Errorf("lookup user: %w", err)
	}
	if user == nil {
		return nil, errNotFound("QR код не найден")
	}

	socials, _ := s.socialRepo.GetPublicByUserID(ctx, user.DisplayID)
	scans, _ := s.scanRepo.CountByQRCodeID(ctx, qr.Code)
	tgAcc, _ := s.tgRepo.GetByUserID(ctx, user.DisplayID)

	owner := &QROwnerOut{
		DisplayID:         user.DisplayID,
		Phone:             user.Phone,
		FirstName:         user.FirstName,
		LastName:          user.LastName,
		AvatarURL:         user.AvatarURL,
		VehicleNumber:     veh.PlateNumber,
		VehicleBrand:      veh.CarModel,
		VehiclePhotoURL:   veh.PhotoURL,
		TelegramEnabled:   veh.TelegramEnabled,
		TelegramBotLinked: tgAcc != nil,
		ScanCount:         scans,
		WhatsApp:          pickSocial(socials, types.PlatformWhatsApp),
		Instagram:         pickSocial(socials, types.PlatformInstagram),
		Telegram:          pickSocial(socials, types.PlatformTelegram),
		VK:                pickSocial(socials, types.PlatformVK),
		Facebook:          pickSocial(socials, types.PlatformFacebook),
	}
	return &QRInfoResult{Registered: true, Owner: owner}, nil
}

// ─── QR registration ─────────────────────────────────────────────────────────

// loginExistingQROwner handles VerifyQR when QR is already registered (active).
// 1) Phone matches QR owner → token for owner.
// 2) Phone differs but exists in DB → token for that account (same user as phone).
// 3) Otherwise → validation error (unknown OTP phone for this QR).
func (s *APISpecService) loginExistingQROwner(ctx context.Context, qr *types.QRCode, phone string) (uuid.UUID, bool, error) {
	if qr.VehicleID == nil {
		return uuid.Nil, false, errValidation("QR код уже зарегистрирован", map[string]string{"qr_id": "Уже зарегистрирован"})
	}
	veh, err := s.vehicleRepo.GetByID(ctx, *qr.VehicleID)
	if err != nil {
		return uuid.Nil, false, fmt.Errorf("lookup vehicle: %w", err)
	}
	if veh == nil {
		return uuid.Nil, false, errNotFound("QR код не найден")
	}
	owner, err := s.userRepo.GetByID(ctx, veh.UserID)
	if err != nil {
		return uuid.Nil, false, fmt.Errorf("lookup user: %w", err)
	}
	if owner == nil {
		return uuid.Nil, false, errNotFound("Пользователь не найден")
	}
	p := strings.TrimSpace(phone)
	if strings.TrimSpace(owner.Phone) == p {
		return owner.DisplayID, false, nil
	}
	byPhone, err := s.userRepo.GetByPhone(ctx, phone)
	if err != nil {
		return uuid.Nil, false, fmt.Errorf("lookup user by phone: %w", err)
	}
	if byPhone != nil {
		return byPhone.DisplayID, false, nil
	}
	return uuid.Nil, false, errValidation(
		"Нет пользователя с таким номером. Укажите номер владельца наклейки или зарегистрируйтесь.",
		map[string]string{"phone": "Номер не найден в системе"},
	)
}

// Register claims a QR code for the given phone number.
// If the user does not exist, they are created with the phone only.
// A blank vehicle is created and linked to the QR so the owner can fill in
// details afterwards via UpdateOwnerProfile / UpdateOwnerVehicle.
// If the QR is already registered (active), resolves user id: owner phone, or any existing account with the same phone.
func (s *APISpecService) Register(ctx context.Context, in QRRegisterInput) (userID uuid.UUID, isNewUser bool, err error) {
	phone := strings.TrimSpace(in.Phone)
	if phone == "" {
		return uuid.Nil, false, errValidation("Номер телефона обязателен", map[string]string{"phone": "Обязательное поле"})
	}

	qr, err := s.qrRepo.GetByCodeOrID(ctx, in.QRID)
	if err != nil {
		return uuid.Nil, false, fmt.Errorf("lookup qr: %w", err)
	}
	if qr == nil || qr.Status == types.QRStatusBlocked {
		return uuid.Nil, false, errNotFound("QR код не найден")
	}
	// Уже привязан к машине — после OTP выдаём тот же ответ, что и при первой регистрации (токен + user_id).
	if qr.Status == types.QRStatusActive && qr.VehicleID != nil {
		return s.loginExistingQROwner(ctx, qr, phone)
	}
	if qr.Status != types.QRStatusUnregistered || qr.VehicleID != nil {
		return uuid.Nil, false, errValidation("QR код уже зарегистрирован", map[string]string{"qr_id": "Уже зарегистрирован"})
	}

	existing, err := s.userRepo.GetByPhone(ctx, phone)
	if err != nil {
		return uuid.Nil, false, fmt.Errorf("lookup user: %w", err)
	}

	if existing != nil {
		userID = existing.DisplayID
		isNewUser = false
	} else {
		userID = uuid.New()
		if err := s.userRepo.Create(ctx, &types.User{
			DisplayID: userID,
			Phone:     phone,
		}); err != nil {
			return uuid.Nil, false, fmt.Errorf("create user: %w", err)
		}
		isNewUser = true
	}

	// Create a blank vehicle — owner fills in plate/model via PATCH /me/vehicles/:id
	vehID := uuid.New()
	if err := s.vehicleRepo.Create(ctx, &types.Vehicle{
		DisplayID:      vehID,
		UserID:         userID,
		IsPublic:       true,
		ReviewsEnabled: true,
	}); err != nil {
		return uuid.Nil, false, fmt.Errorf("create vehicle: %w", err)
	}

	if err := s.qrRepo.Register(ctx, qr.Code, vehID); err != nil {
		return uuid.Nil, false, fmt.Errorf("register qr: %w", err)
	}

	return userID, isNewUser, nil
}

// ─── Owner self-management ────────────────────────────────────────────────────

// GetOwnerData returns the full private profile for the authenticated user.
func (s *APISpecService) GetOwnerData(ctx context.Context, userID uuid.UUID) (*OwnerDataOut, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("lookup user: %w", err)
	}
	if user == nil {
		return nil, errNotFound("Пользователь не найден")
	}

	vehicles, _ := s.vehicleRepo.GetByUserID(ctx, userID)
	vehicleOuts := make([]OwnerVehicleOut, 0, len(vehicles))
	for _, v := range vehicles {
		vOut := OwnerVehicleOut{
			ID:              v.DisplayID,
			PlateNumber:     v.PlateNumber,
			CarModel:        v.CarModel,
			PhotoURL:        v.PhotoURL,
			TelegramEnabled: v.TelegramEnabled,
		}
		if codes, _ := s.qrRepo.GetByVehicleID(ctx, v.DisplayID); len(codes) > 0 {
			vOut.QRCode = codes[0].Code
		}
		vehicleOuts = append(vehicleOuts, vOut)
	}

	socials, _ := s.socialRepo.GetByUserID(ctx, userID)

	return &OwnerDataOut{
		ID:        user.DisplayID,
		Phone:     user.Phone,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		AvatarURL: user.AvatarURL,
		Vehicles:  vehicleOuts,
		WhatsApp:  pickSocial(socials, types.PlatformWhatsApp),
		Instagram: pickSocial(socials, types.PlatformInstagram),
		Telegram:  pickSocial(socials, types.PlatformTelegram),
		VK:        pickSocial(socials, types.PlatformVK),
		Facebook:  pickSocial(socials, types.PlatformFacebook),
	}, nil
}

// UpdateOwnerProfile updates the user's name, avatar and social profiles.
func (s *APISpecService) UpdateOwnerProfile(ctx context.Context, userID uuid.UUID, in OwnerProfileInput) (*OwnerDataOut, error) {
	if _, err := s.userRepo.Update(ctx, userID, types.UserUpdate{
		FirstName: in.FirstName,
		LastName:  in.LastName,
		AvatarURL: in.AvatarURL,
	}); err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}

	if err := s.upsertSocials(ctx, userID, []socialPatch{
		{types.PlatformWhatsApp, in.WhatsApp},
		{types.PlatformInstagram, in.Instagram},
		{types.PlatformTelegram, in.Telegram},
		{types.PlatformVK, in.VK},
		{types.PlatformFacebook, in.Facebook},
	}); err != nil {
		return nil, err
	}

	return s.GetOwnerData(ctx, userID)
}

// UpdateOwnerVehicle updates a vehicle that belongs to the given user.
func (s *APISpecService) UpdateOwnerVehicle(ctx context.Context, userID, vehicleID uuid.UUID, upd types.VehicleUpdate) (*OwnerVehicleOut, error) {
	veh, err := s.vehicleRepo.GetByID(ctx, vehicleID)
	if err != nil {
		return nil, fmt.Errorf("lookup vehicle: %w", err)
	}
	if veh == nil {
		return nil, errNotFound("Автомобиль не найден")
	}
	if veh.UserID != userID {
		return nil, errValidation("Нет доступа к этому автомобилю", nil)
	}

	updated, err := s.vehicleRepo.Update(ctx, vehicleID, upd)
	if err != nil {
		return nil, fmt.Errorf("update vehicle: %w", err)
	}

	out := &OwnerVehicleOut{
		ID:              updated.DisplayID,
		PlateNumber:     updated.PlateNumber,
		CarModel:        updated.CarModel,
		PhotoURL:        updated.PhotoURL,
		TelegramEnabled: updated.TelegramEnabled,
	}
	if codes, _ := s.qrRepo.GetByVehicleID(ctx, updated.DisplayID); len(codes) > 0 {
		out.QRCode = codes[0].Code
	}
	return out, nil
}

// ─── Internal helpers ─────────────────────────────────────────────────────────

type socialPatch struct {
	platform types.SocialPlatform
	handle   *string
}

// upsertSocials inserts/updates socials when handle is non-empty, deletes when handle is "".
func (s *APISpecService) upsertSocials(ctx context.Context, userID uuid.UUID, patches []socialPatch) error {
	for _, p := range patches {
		if p.handle == nil {
			continue
		}
		h := strings.TrimSpace(*p.handle)
		if h == "" {
			if existing, _ := s.socialRepo.GetByUserIDAndPlatform(ctx, userID, p.platform); existing != nil {
				_ = s.socialRepo.Delete(ctx, existing.ID)
			}
		} else {
			if err := s.socialRepo.UpsertByPlatform(ctx, userID, p.platform, h, true); err != nil {
				return fmt.Errorf("upsert %s: %w", p.platform, err)
			}
		}
	}
	return nil
}

func pickSocial(profiles []types.SocialProfile, platform types.SocialPlatform) *string {
	for i := range profiles {
		if profiles[i].Platform == platform {
			h := profiles[i].Handle
			return &h
		}
	}
	return nil
}

// ─── Error helpers ────────────────────────────────────────────────────────────

type specErrKind int

const (
	kindNotFound   specErrKind = iota
	kindValidation specErrKind = iota
)

type specError struct {
	kind    specErrKind
	message string
	fields  map[string]string
}

func (e *specError) Error() string { return e.message }

func errNotFound(msg string) error {
	return &specError{kind: kindNotFound, message: msg}
}

func errValidation(msg string, fields map[string]string) error {
	return &specError{kind: kindValidation, message: msg, fields: fields}
}

func IsSpecNotFound(err error) bool {
	var se *specError
	return errors.As(err, &se) && se.kind == kindNotFound
}

func IsSpecValidation(err error) bool {
	var se *specError
	return errors.As(err, &se) && se.kind == kindValidation
}

func SpecErrorFields(err error) map[string]string {
	var se *specError
	if errors.As(err, &se) {
		return se.fields
	}
	return nil
}

// Unused but kept for time.Time zero value reference
var _ = time.Time{}
