package server

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	adminhandler "qr-parking/handlers/admin"
	clienthandler "qr-parking/handlers/client"
	"qr-parking/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	fiberlogger "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/swagger"
)

// @title QR-Parking API
// @version 1.0
// @description Scan a QR code to view vehicle owner info and send them a message.
// @host qr-parking-back.abdullokh.com
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter "Bearer {token}"

type Server struct {
	app      *fiber.App
	services Services
}

func New() (*Server, error) {
	e := NewEssentials()
	applySwaggerFromBaseURL(e.Vars[AppBaseURLVar])
	s := NewServices(e)

	app := fiber.New(fiber.Config{
		ReadTimeout:  parseDurationOrDefault(e.Vars[ServerReadTimeoutVar], 10*time.Second),
		WriteTimeout: parseDurationOrDefault(e.Vars[ServerWriteTimeoutVar], 10*time.Second),
		AppName:      "QR-Parking API",
	})

	app.Use(recover.New())
	app.Use(fiberlogger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, PUT, PATCH, DELETE, OPTIONS",
	}))

	app.Get("/swagger/*", swagger.New(swagger.Config{InstanceName: "swagger"}))

	qrH := clienthandler.NewQRHandler(s.QR, s.APISpec, s.Message, s.JWTMgr, s.OTP)
	meH := clienthandler.NewMeHandler(s.APISpec)
	adminH := adminhandler.NewAdminHandler(
		s.AdminRepo, s.QR, s.UserRepo, s.QRRepo, s.ScanRepo, s.MsgRepo,
		s.Pool, s.Redis, s.Logger,
	)
	authH := adminhandler.NewAuthHandler(s.Auth)

	api := app.Group("/api/v1")

	// ── Public QR endpoints ───────────────────────────────────────────────────
	api.Get("/qr/:qr_id", qrH.GetQR)
	api.Post("/qr/:qr_id", qrH.RegisterQR)
	api.Post("/qr-verify/:qr_id", qrH.VerifyQR)
	api.Post("/qr-message/:qr_id", qrH.PostQRMessage)

	// ── Owner self-management (JWT required) ──────────────────────────────────
	me := api.Group("/me", middleware.JWTAuth(s.JWTMgr))
	me.Get("", meH.GetMe)
	me.Patch("", meH.PatchMe)
	me.Get("/vehicles", meH.GetVehicles)
	me.Patch("/vehicles/:id", meH.PatchVehicle)

	// ── Admin endpoints ───────────────────────────────────────────────────────
	api.Post("/admin/login", authH.AdminLogin)

	admin := api.Group("/admin", middleware.JWTAuth(s.JWTMgr), middleware.AdminOnly())
	admin.Get("/admins", adminH.ListAdmins)
	admin.Get("/admins/:id", adminH.GetAdmin)
	admin.Patch("/admins/:id/block", adminH.BlockAdmin)
	admin.Get("/users", adminH.ListUsers)
	admin.Get("/users/:id", adminH.GetUser)
	admin.Patch("/users/:id/block", adminH.BlockUser)
	admin.Post("/qrcodes/generate", adminH.GenerateQRCodes)
	admin.Patch("/qrcodes/:id/block", adminH.BlockQR)
	admin.Get("/messages", adminH.ListMessages)
	admin.Get("/messages/:user_id", adminH.GetMessagesByUserID)

	return &Server{app: app, services: s}, nil
}

func (s *Server) Run() error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	port := s.services.Vars[ServerPortVar]
	go func() {
		if err := s.app.Listen(fmt.Sprintf(":%s", port)); err != nil {
			log.Fatalf("server listen: %v", err)
		}
	}()

	s.services.Logger.Sugar().Infof("Server started on :%s", port)
	<-quit
	s.services.Logger.Info("Shutting down...")
	return s.app.Shutdown()
}

func (s *Server) Close() {
	if s.services.Pool != nil {
		s.services.Pool.Close()
	}
	if s.services.Redis != nil {
		_ = s.services.Redis.Close()
	}
	if s.services.Logger != nil {
		_ = s.services.Logger.Sync()
	}
}

func parseDurationOrDefault(s string, def time.Duration) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil || d <= 0 {
		return def
	}
	return d
}
