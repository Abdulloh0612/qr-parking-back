package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"qr-parking/db/repositories"
	"qr-parking/types"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ErrMessageQRUnavailable is returned when the QR is missing, blocked, or not linked to a vehicle.
var ErrMessageQRUnavailable = errors.New("qr unavailable for messaging")

type MessageService struct {
	messageRepo repositories.MessageRepository
	qrRepo      repositories.QRCodeRepository
	vehicleRepo repositories.VehicleRepository
	userRepo    repositories.UserRepository
	tgRepo      repositories.TelegramRepository
	logger      *zap.Logger
	botToken    string
}

func NewMessageService(
	messageRepo repositories.MessageRepository,
	qrRepo repositories.QRCodeRepository,
	vehicleRepo repositories.VehicleRepository,
	userRepo repositories.UserRepository,
	tgRepo repositories.TelegramRepository,
	logger *zap.Logger,
	botToken string,
) *MessageService {
	return &MessageService{
		messageRepo: messageRepo,
		qrRepo:      qrRepo,
		vehicleRepo: vehicleRepo,
		userRepo:    userRepo,
		tgRepo:      tgRepo,
		logger:      logger,
		botToken:    botToken,
	}
}

// SendMessageByQRRef saves a message for an active QR and notifies the owner via Telegram if linked.
func (s *MessageService) SendMessageByQRRef(ctx context.Context, qrRef, content string, mediaURL *string) (*types.Message, error) {
	qr, err := s.qrRepo.GetByCodeOrID(ctx, qrRef)
	if err != nil {
		return nil, fmt.Errorf("get qr: %w", err)
	}
	if qr == nil || qr.Status != types.QRStatusActive || qr.VehicleID == nil {
		return nil, ErrMessageQRUnavailable
	}

	code := qr.Code
	msg := &types.Message{
		ID:        uuid.New(),
		QRCodeID:  &code,
		VehicleID: *qr.VehicleID,
		Content:   content,
		MediaURL:  mediaURL,
	}
	if err := s.messageRepo.Create(ctx, msg); err != nil {
		return nil, fmt.Errorf("save message: %w", err)
	}

	vehicleID := qr.VehicleID
	go s.notifyOwner(vehicleID, content, mediaURL)

	return msg, nil
}

// notifyOwner finds the vehicle owner's linked Telegram account and sends them a message.
// Runs in a goroutine — errors are logged, not propagated.
func (s *MessageService) notifyOwner(vehicleID *uuid.UUID, text string, mediaURL *string) {
	if vehicleID == nil || s.botToken == "" {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	veh, err := s.vehicleRepo.GetByID(ctx, *vehicleID)
	if err != nil || veh == nil {
		return
	}
	user, err := s.userRepo.GetByID(ctx, veh.UserID)
	if err != nil || user == nil {
		return
	}
	tgAcc, err := s.tgRepo.GetByUserID(ctx, user.DisplayID)
	if err != nil || tgAcc == nil {
		return
	}

	notification := fmt.Sprintf(
		"📩 Вам написали о вашем автомобиле\n\n%s %s\n\n%s",
		veh.CarModel, veh.PlateNumber, text,
	)
	if mediaURL != nil && *mediaURL != "" {
		notification += fmt.Sprintf("\n\n🖼 Прикреплён файл: %s", *mediaURL)
	}
	if err := sendTelegramMessage(s.botToken, tgAcc.TgUserID, notification); err != nil {
		s.logger.Warn("telegram notification failed", zap.Error(err))
	}
}

func sendTelegramMessage(botToken string, tgUserID int64, text string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)
	body, _ := json.Marshal(map[string]any{
		"chat_id": tgUserID,
		"text":    text,
	})
	resp, err := http.Post(url, "application/json", bytes.NewReader(body)) //nolint:noctx
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram api status %d", resp.StatusCode)
	}
	return nil
}
