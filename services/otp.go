package services

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const otpTTL = 5 * time.Minute

// OTPService handles SMS one-time password generation and verification via Redis.
type OTPService struct {
	redis  *redis.Client
	logger *zap.Logger
}

func NewOTPService(rdb *redis.Client, logger *zap.Logger) *OTPService {
	return &OTPService{redis: rdb, logger: logger}
}

// Send generates a 6-digit OTP, stores it in Redis and sends it via SMS.
// For development it logs the code to stdout.
func (s *OTPService) Send(ctx context.Context, phone string) error {
	otp := fmt.Sprintf("%06d", rand.Intn(1_000_000)) //nolint:gosec
	key := fmt.Sprintf("otp:%s", phone)

	if err := s.redis.Set(ctx, key, otp, otpTTL).Err(); err != nil {
		return fmt.Errorf("store otp: %w", err)
	}

	// TODO: replace with a real SMS gateway (e.g. ESKIZ, PlayMobile)
	s.logger.Info("OTP sent", zap.String("phone", phone), zap.String("otp", otp))
	return nil
}

// Verify checks the OTP submitted by the user against the stored value.
// Returns true and deletes the code on success; false on mismatch or expiry.
func (s *OTPService) Verify(ctx context.Context, phone, otp string) (bool, error) {
	key := fmt.Sprintf("otp:%s", phone)
	stored, err := s.redis.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("get otp: %w", err)
	}
	if stored != otp {
		return false, nil
	}
	_ = s.redis.Del(ctx, key)
	return true, nil
}
