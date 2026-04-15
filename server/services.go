package server

import (
	"qr-parking/db/repositories"
	"qr-parking/services"
)

// Services aggregates all repositories and application services used by the HTTP server.
type Services struct {
	Essentials

	// Repositories accessed directly by admin handlers
	UserRepo repositories.UserRepository
	QRRepo   repositories.QRCodeRepository
	ScanRepo repositories.ScanEventRepository
	MsgRepo  repositories.MessageRepository

	// Application services
	Auth    *services.AuthService
	QR      *services.QRService
	APISpec *services.APISpecService
	Message *services.MessageService
}

func NewServices(e Essentials) Services {
	adminRepo := repositories.NewAdminRepo(e.Pool)
	userRepo := repositories.NewUserRepo(e.Pool)
	vehicleRepo := repositories.NewVehicleRepo(e.Pool)
	qrRepo := repositories.NewQRCodeRepo(e.Pool)
	scanRepo := repositories.NewScanEventRepo(e.Pool)
	msgRepo := repositories.NewMessageRepo(e.Pool)
	socialRepo := repositories.NewSocialProfileRepo(e.Pool)
	tgRepo := repositories.NewTelegramRepo(e.Pool)

	authSvc := services.NewAuthService(adminRepo, e.JWTMgr)
	qrSvc := services.NewQRService(qrRepo, e.Logger, e.Vars[AppBaseURLVar])
	apiSpecSvc := services.NewAPISpecService(userRepo, vehicleRepo, qrRepo, socialRepo, scanRepo)
	msgSvc := services.NewMessageService(msgRepo, qrRepo, vehicleRepo, userRepo, tgRepo, e.Logger, e.Vars[TGBotTokenVar])

	return Services{
		Essentials: e,
		UserRepo:   userRepo,
		QRRepo:     qrRepo,
		ScanRepo:   scanRepo,
		MsgRepo:    msgRepo,
		Auth:       authSvc,
		QR:         qrSvc,
		APISpec:    apiSpecSvc,
		Message:    msgSvc,
	}
}
