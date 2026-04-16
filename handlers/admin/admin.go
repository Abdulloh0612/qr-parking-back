package admin

import (
	"strconv"
	"time"

	"qr-parking/db/repositories"
	"qr-parking/middleware"
	"qr-parking/services"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type AdminHandler struct {
	adminRepo repositories.AdminRepository
	qrSvc     *services.QRService
	userRepo  repositories.UserRepository
	qrRepo    repositories.QRCodeRepository
	scanRepo  repositories.ScanEventRepository
	msgRepo   repositories.MessageRepository
	pool      *pgxpool.Pool
	redis     *redis.Client
	logger    *zap.Logger
}

func NewAdminHandler(
	adminRepo repositories.AdminRepository,
	qrSvc *services.QRService,
	userRepo repositories.UserRepository,
	qrRepo repositories.QRCodeRepository,
	scanRepo repositories.ScanEventRepository,
	msgRepo repositories.MessageRepository,
	pool *pgxpool.Pool,
	redis *redis.Client,
	logger *zap.Logger,
) *AdminHandler {
	return &AdminHandler{
		adminRepo: adminRepo,
		qrSvc:     qrSvc,
		userRepo:  userRepo,
		qrRepo:    qrRepo,
		scanRepo:  scanRepo,
		msgRepo:   msgRepo,
		pool:      pool,
		redis:     redis,
		logger:    logger,
	}
}

// ─── Admin management ─────────────────────────────────────────────────────────

// ListAdmins godoc
// @Summary List all admins (paginated)
// @Tags Admin
// @Security BearerAuth
// @Param page query int false "Page" default(1)
// @Param limit query int false "Limit" default(20)
// @Success 200 {object} PaginatedResponse
// @Router /admin/admins [get]
func (h *AdminHandler) ListAdmins(c *fiber.Ctx) error {
	page, limit, offset := parsePage(c)
	admins, total, err := h.adminRepo.List(c.Context(), offset, limit)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	return paginatedResponse(c, admins, total, page, limit)
}

// GetAdmin godoc
// @Summary Get admin by UUID
// @Tags Admin
// @Security BearerAuth
// @Param id path string true "Admin UUID"
// @Success 200 {object} DataResponse
// @Router /admin/admins/{id} [get]
func (h *AdminHandler) GetAdmin(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "invalid admin id")
	}
	admin, err := h.adminRepo.GetByDisplayID(c.Context(), id)
	if err != nil || admin == nil {
		return errorResponse(c, fiber.StatusNotFound, "admin not found")
	}
	return successResponse(c, admin)
}

// BlockAdmin godoc
// @Summary Remove an admin account
// @Tags Admin
// @Security BearerAuth
// @Param id path string true "Admin UUID"
// @Success 200 {object} MessageResponse
// @Router /admin/admins/{id}/block [patch]
func (h *AdminHandler) BlockAdmin(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "invalid admin id")
	}
	callerID := middleware.GetUserID(c)
	if callerID == id {
		return errorResponse(c, fiber.StatusBadRequest, "cannot remove yourself")
	}
	if err := h.adminRepo.Block(c.Context(), id); err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(MessageResponse{Message: "admin removed"})
}

// ─── Stats ───────────────────────────────────────────────────────────────────

type StatsResponse struct {
	TotalQRCodes   int `json:"total_qr_codes"`
	UnregisteredQR int `json:"unregistered_qr"`
	ActiveQR       int `json:"active_qr"`
	BlockedQR      int `json:"blocked_qr"`
	TotalUsers     int `json:"total_users"`
	ScansToday     int `json:"scans_today"`
	TotalMessages  int `json:"total_messages"`
}

// GetStats godoc
// @Summary Platform statistics
// @Tags Admin
// @Security BearerAuth
// @Success 200 {object} StatsResponse
// @Router /admin/stats [get]
func (h *AdminHandler) GetStats(c *fiber.Ctx) error {
	ctx := c.Context()

	qrCounts, _ := h.qrRepo.CountByStatus(ctx)
	scansToday, _ := h.scanRepo.CountToday(ctx)
	msgCount, _ := h.msgRepo.CountAll(ctx)
	_, userTotal, _ := h.userRepo.List(ctx, 0, 1)

	totalQR := 0
	for _, v := range qrCounts {
		totalQR += v
	}
	return c.JSON(StatsResponse{
		TotalQRCodes:   totalQR,
		UnregisteredQR: qrCounts["unregistered"],
		ActiveQR:       qrCounts["active"],
		BlockedQR:      qrCounts["blocked"],
		TotalUsers:     userTotal,
		ScansToday:     scansToday,
		TotalMessages:  msgCount,
	})
}

// ─── Users ───────────────────────────────────────────────────────────────────

// ListUsers godoc
// @Summary List all users (paginated)
// @Tags Admin
// @Security BearerAuth
// @Param page query int false "Page" default(1)
// @Param limit query int false "Limit" default(20)
// @Success 200 {object} PaginatedResponse
// @Router /admin/users [get]
func (h *AdminHandler) ListUsers(c *fiber.Ctx) error {
	page, limit, offset := parsePage(c)
	users, total, err := h.userRepo.List(c.Context(), offset, limit)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	return paginatedResponse(c, users, total, page, limit)
}

// GetUser godoc
// @Summary Get user by UUID
// @Tags Admin
// @Security BearerAuth
// @Param id path string true "User UUID"
// @Success 200 {object} DataResponse
// @Router /admin/users/{id} [get]
func (h *AdminHandler) GetUser(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "invalid user id")
	}
	user, err := h.userRepo.GetByID(c.Context(), id)
	if err != nil || user == nil {
		return errorResponse(c, fiber.StatusNotFound, "user not found")
	}
	return successResponse(c, user)
}

// BlockUser godoc
// @Summary Block (delete) a user
// @Tags Admin
// @Security BearerAuth
// @Param id path string true "User UUID"
// @Success 200 {object} MessageResponse
// @Router /admin/users/{id}/block [patch]
func (h *AdminHandler) BlockUser(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "invalid user id")
	}
	if err := h.userRepo.Block(c.Context(), id); err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(MessageResponse{Message: "user blocked"})
}

// ─── QR codes ────────────────────────────────────────────────────────────────

type GenerateQRRequest struct {
	Count int `json:"count" validate:"required,min=1,max=1000"`
}

// ListQRCodes godoc
// @Summary List QR codes (paginated, filterable by status)
// @Tags Admin
// @Security BearerAuth
// @Param status query string false "Filter: unregistered | active | blocked"
// @Param page query int false "Page" default(1)
// @Param limit query int false "Limit" default(20)
// @Success 200 {object} PaginatedResponse
// @Router /admin/qrcodes [get]
func (h *AdminHandler) ListQRCodes(c *fiber.Ctx) error {
	page, limit, offset := parsePage(c)
	codes, total, err := h.qrRepo.List(c.Context(), c.Query("status"), offset, limit)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	return paginatedResponse(c, codes, total, page, limit)
}

// GenerateQRCodes godoc
// @Summary Generate new QR codes (1–1000)
// @Tags Admin
// @Security BearerAuth
// @Param body body GenerateQRRequest true "Count"
// @Success 201 {object} DataResponse
// @Router /admin/qrcodes/generate [post]
func (h *AdminHandler) GenerateQRCodes(c *fiber.Ctx) error {
	var req GenerateQRRequest
	if err := c.BodyParser(&req); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "invalid request body")
	}
	if req.Count <= 0 {
		req.Count = 1
	}
	adminID := middleware.GetUserID(c)
	codes, err := h.qrSvc.GenerateQRCodes(c.Context(), req.Count, adminID)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": codes, "count": len(codes)})
}

// BlockQR godoc
// @Summary Block a QR code
// @Tags Admin
// @Security BearerAuth
// @Param id path string true "QR code string or numeric id"
// @Success 200 {object} MessageResponse
// @Router /admin/qrcodes/{id}/block [patch]
func (h *AdminHandler) BlockQR(c *fiber.Ctx) error {
	qrRef := c.Params("id")
	if qrRef == "" {
		return errorResponse(c, fiber.StatusBadRequest, "qr id is required")
	}
	if err := h.qrSvc.BlockQR(c.Context(), qrRef); err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(MessageResponse{Message: "QR code blocked"})
}

// ─── Messages ────────────────────────────────────────────────────────────────

// ListMessages godoc
// @Summary List all messages (paginated)
// @Tags Admin
// @Security BearerAuth
// @Param page query int false "Page" default(1)
// @Param limit query int false "Limit" default(20)
// @Success 200 {object} PaginatedResponse
// @Router /admin/messages [get]
func (h *AdminHandler) ListMessages(c *fiber.Ctx) error {
	page, limit, offset := parsePage(c)
	messages, total, err := h.msgRepo.List(c.Context(), offset, limit)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	return paginatedResponse(c, messages, total, page, limit)
}

// GetMessagesByUserID godoc
// @Summary Get messages by user UUID
// @Tags Admin
// @Security BearerAuth
// @Param user_id path string true "User UUID"
// @Param page query int false "Page" default(1)
// @Param limit query int false "Limit" default(20)
// @Success 200 {object} PaginatedResponse
// @Router /admin/messages/{user_id} [get]
func (h *AdminHandler) GetMessagesByUserID(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Params("user_id"))
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "invalid user id")
	}
	page, limit, offset := parsePage(c)
	messages, total, err := h.msgRepo.GetByUserID(c.Context(), userID, offset, limit)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	return paginatedResponse(c, messages, total, page, limit)
}

// ─── Monitoring ──────────────────────────────────────────────────────────────

// GetMonitoringHealth godoc
// @Summary Health check (DB + Redis)
// @Tags Admin
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /admin/monitoring/health [get]
func (h *AdminHandler) GetMonitoringHealth(c *fiber.Ctx) error {
	ctx := c.Context()

	dbStatus := "ok"
	if err := h.pool.Ping(ctx); err != nil {
		dbStatus = "error: " + err.Error()
	}
	redisStatus := "ok"
	if err := h.redis.Ping(ctx).Err(); err != nil {
		redisStatus = "error: " + err.Error()
	}
	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"database": dbStatus,
			"redis":    redisStatus,
			"time":     time.Now().UTC(),
		},
	})
}

// GetMonitoringMetrics godoc
// @Summary Platform metrics + DB pool stats
// @Tags Admin
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /admin/monitoring/metrics [get]
func (h *AdminHandler) GetMonitoringMetrics(c *fiber.Ctx) error {
	ctx := c.Context()

	qrCounts, _ := h.qrRepo.CountByStatus(ctx)
	scansToday, _ := h.scanRepo.CountToday(ctx)
	msgCount, _ := h.msgRepo.CountAll(ctx)
	_, userTotal, _ := h.userRepo.List(ctx, 0, 1)

	totalQR := 0
	for _, v := range qrCounts {
		totalQR += v
	}
	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"total_qr_codes":   totalQR,
			"unregistered_qr":  qrCounts["unregistered"],
			"active_qr":        qrCounts["active"],
			"blocked_qr":       qrCounts["blocked"],
			"total_users":      userTotal,
			"scans_today":      scansToday,
			"total_messages":   msgCount,
			"db_pool_acquired": h.pool.Stat().AcquiredConns(),
			"db_pool_idle":     h.pool.Stat().IdleConns(),
		},
	})
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func parsePage(c *fiber.Ctx) (page, limit, offset int) {
	page, _ = strconv.Atoi(c.Query("page", "1"))
	limit, _ = strconv.Atoi(c.Query("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 200 {
		limit = 20
	}
	offset = (page - 1) * limit
	return
}
