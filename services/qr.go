package services

import (
	"context"
	"fmt"

	"qr-parking/db/repositories"
	"qr-parking/pkg/qrgen"
	"qr-parking/types"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type QRService struct {
	qrRepo  repositories.QRCodeRepository
	logger  *zap.Logger
	baseURL string
}

func NewQRService(
	qrRepo repositories.QRCodeRepository,
	logger *zap.Logger,
	baseURL string,
) *QRService {
	return &QRService{
		qrRepo:  qrRepo,
		logger:  logger,
		baseURL: baseURL,
	}
}

// GenerateQRCodes creates N new unregistered QR codes and saves them to the DB.
func (s *QRService) GenerateQRCodes(ctx context.Context, count int, adminID uuid.UUID) ([]types.QRCode, error) {
	if count <= 0 || count > 1000 {
		return nil, fmt.Errorf("count must be between 1 and 1000")
	}

	codes := make([]types.QRCode, 0, count)
	for i := 0; i < count; i++ {
		code, err := qrgen.GenerateCode()
		if err != nil {
			return nil, fmt.Errorf("generate code: %w", err)
		}

		qr := types.QRCode{
			DisplayID: uuid.New(),
			Code:      code,
			Status:    types.QRStatusUnregistered,
			CreatedBy: &adminID,
		}
		if err := s.qrRepo.Create(ctx, &qr); err != nil {
			return nil, fmt.Errorf("save qr code: %w", err)
		}
		codes = append(codes, qr)
	}
	return codes, nil
}

// GetQRImage returns a PNG image of the QR code that encodes the public scan URL.
func (s *QRService) GetQRImage(ctx context.Context, qrRef string, size int) ([]byte, error) {
	qr, err := s.qrRepo.GetByCodeOrID(ctx, qrRef)
	if err != nil {
		return nil, fmt.Errorf("get qr: %w", err)
	}

	code := qrRef
	if qr != nil {
		code = qr.Code
	}

	url := fmt.Sprintf("%s/qr/%s", s.baseURL, code)
	return qrgen.GenerateQRImageURL(url, size)
}

// BlockQR sets a QR code's status to blocked. qrRef can be the code string or numeric id.
func (s *QRService) BlockQR(ctx context.Context, qrRef string) error {
	return s.qrRepo.Block(ctx, qrRef)
}
