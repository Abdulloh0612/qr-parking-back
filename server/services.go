package server

import (
	"qr-parking/db/repositories"
	"qr-parking/services"
)

// Services aggregates all repositories and application services used by the HTTP server.
type Services struct {
	Essentials

	// Repositories accessed directly by handlers
	AdminRepo   repositories.AdminRepository
	UserRepo    repositories.UserRepository
	VehicleRepo repositories.VehicleRepository
	QRRepo      repositories.QRCodeRepository
	ScanRepo    repositories.ScanEventRepository
	MsgRepo     repositories.MessageRepository

	// Application services
	Auth    *services.AuthService
	QR      *services.QRService
	APISpec *services.APISpecService
	Message *services.MessageService
	OTP     *services.OTPService
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
	apiSpecSvc := services.NewAPISpecService(userRepo, vehicleRepo, qrRepo, socialRepo, scanRepo, tgRepo)
	msgSvc := services.NewMessageService(msgRepo, qrRepo, vehicleRepo, userRepo, tgRepo, e.Logger, e.Vars[TGBotTokenVar])
	otpSvc := services.NewOTPService(e.Redis, e.Logger)

	return Services{
		Essentials:  e,
		AdminRepo:   adminRepo,
		UserRepo:    userRepo,
		VehicleRepo: vehicleRepo,
		QRRepo:      qrRepo,
		ScanRepo:    scanRepo,
		MsgRepo:     msgRepo,
		Auth:       authSvc,
		QR:         qrSvc,
		APISpec:    apiSpecSvc,
		Message:    msgSvc,
		OTP:        otpSvc,
	}
}
