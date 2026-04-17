package admin

import (
	"errors"
	"strconv"
	"time"

	"qr-parking/db/repositories"
	"qr-parking/middleware"
	"qr-parking/pkg/validator"
	"qr-parking/services"
	"qr-parking/types"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type AdminHandler struct {
	adminRepo   repositories.AdminRepository
	qrSvc       *services.QRService
	userRepo    repositories.UserRepository
	vehicleRepo repositories.VehicleRepository
	qrRepo      repositories.QRCodeRepository
	scanRepo    repositories.ScanEventRepository
	msgRepo     repositories.MessageRepository
	pool        *pgxpool.Pool
	redis       *redis.Client
	logger      *zap.Logger
}

func NewAdminHandler(
	adminRepo repositories.AdminRepository,
	qrSvc *services.QRService,
	userRepo repositories.UserRepository,
	vehicleRepo repositories.VehicleRepository,
	qrRepo repositories.QRCodeRepository,
	scanRepo repositories.ScanEventRepository,
	msgRepo repositories.MessageRepository,
	pool *pgxpool.Pool,
	redis *redis.Client,
	logger *zap.Logger,
) *AdminHandler {
	return &AdminHandler{
		adminRepo:   adminRepo,
		qrSvc:       qrSvc,
		userRepo:    userRepo,
		vehicleRepo: vehicleRepo,
		qrRepo:      qrRepo,
		scanRepo:    scanRepo,
		msgRepo:     msgRepo,
		pool:        pool,
		redis:       redis,
		logger:      logger,
	}
}

// ─── Admin management ─────────────────────────────────────────────────────────

// GetAdminMe godoc
// @Summary Current admin session (username, role)
// @Tags Admin
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /admin/me [get]
func (h *AdminHandler) GetAdminMe(c *fiber.Ctx) error {
	id := middleware.GetUserID(c)
	a, err := h.adminRepo.GetByDisplayID(c.Context(), id)
	if err != nil || a == nil {
		return errorResponse(c, fiber.StatusNotFound, "admin not found")
	}
	role := a.Role
	if role == "" {
		role = "admin"
	}
	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"username":   a.Username,
			"role":       role,
			"display_id": a.DisplayID.String(),
		},
	})
}

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

type createAdminRequest struct {
	Username string `json:"username" validate:"required,min=2,max=64"`
	Password string `json:"password" validate:"required,min=8"`
	Role     string `json:"role" validate:"omitempty,oneof=admin super_admin"`
}

// CreateAdmin godoc
// @Summary Create admin user (super_admin only)
// @Tags Admin
// @Security BearerAuth
// @Router /admin/admins [post]
func (h *AdminHandler) CreateAdmin(c *fiber.Ctx) error {
	var req createAdminRequest
	if err := c.BodyParser(&req); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "invalid request body")
	}
	if err := validator.ValidateStruct(req); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	role := req.Role
	if role == "" {
		role = "admin"
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "hash password failed")
	}
	a, err := h.adminRepo.Create(c.Context(), req.Username, string(hash), role)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return errorResponse(c, fiber.StatusConflict, "username already taken")
		}
		h.logger.Sugar().Errorw("create admin", "err", err)
		return errorResponse(c, fiber.StatusInternalServerError, "create admin failed")
	}
	return successResponse(c, a)
}

type patchAdminBody struct {
	Password *string `json:"password"`
	Role     *string `json:"role"`
}

// PatchAdmin godoc
// @Summary Update admin password and/or role (super_admin only)
// @Tags Admin
// @Security BearerAuth
// @Router /admin/admins/{id} [patch]
func (h *AdminHandler) PatchAdmin(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "invalid admin id")
	}
	var body patchAdminBody
	if err := c.BodyParser(&body); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "invalid request body")
	}
	if body.Password == nil && body.Role == nil {
		return errorResponse(c, fiber.StatusBadRequest, "nothing to update")
	}
	ctx := c.Context()
	target, err := h.adminRepo.GetByDisplayID(ctx, id)
	if err != nil || target == nil {
		return errorResponse(c, fiber.StatusNotFound, "admin not found")
	}
	if body.Role != nil {
		r := *body.Role
		if r != "admin" && r != "super_admin" {
			return errorResponse(c, fiber.StatusBadRequest, "invalid role")
		}
		if target.Role == "super_admin" && r == "admin" {
			n, err := h.adminRepo.CountByRole(ctx, "super_admin")
			if err != nil {
				return errorResponse(c, fiber.StatusInternalServerError, err.Error())
			}
			if n <= 1 {
				return errorResponse(c, fiber.StatusBadRequest, "cannot demote last super admin")
			}
		}
	}
	var hash *string
	if body.Password != nil && *body.Password != "" {
		if len(*body.Password) < 8 {
			return errorResponse(c, fiber.StatusBadRequest, "password min 8 characters")
		}
		b, err := bcrypt.GenerateFromPassword([]byte(*body.Password), bcrypt.DefaultCost)
		if err != nil {
			return errorResponse(c, fiber.StatusInternalServerError, "hash password failed")
		}
		hs := string(b)
		hash = &hs
	}
	var rolePtr *string
	if body.Role != nil {
		rolePtr = body.Role
	}
	updated, err := h.adminRepo.UpdateByDisplayID(ctx, id, hash, rolePtr)
	if err != nil || updated == nil {
		return errorResponse(c, fiber.StatusInternalServerError, "update failed")
	}
	return successResponse(c, updated)
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
	ctx := c.Context()
	target, err := h.adminRepo.GetByDisplayID(ctx, id)
	if err != nil || target == nil {
		return errorResponse(c, fiber.StatusNotFound, "admin not found")
	}
	if target.Role == "super_admin" {
		n, err := h.adminRepo.CountByRole(ctx, "super_admin")
		if err != nil {
			return errorResponse(c, fiber.StatusInternalServerError, err.Error())
		}
		if n <= 1 {
			return errorResponse(c, fiber.StatusBadRequest, "cannot remove last super admin")
		}
	}
	if err := h.adminRepo.Block(ctx, id); err != nil {
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

// ─── Dashboard ───────────────────────────────────────────────────────────────

type DashboardGrowthPoint struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

type DashboardResponse struct {
	TotalUsers    int                    `json:"total_users"`
	TotalQRCodes  int                    `json:"total_qr_codes"`
	ActiveQR      int                    `json:"active_qr"`
	UnregisteredQR int                   `json:"unregistered_qr"`
	BlockedQR     int                    `json:"blocked_qr"`
	TotalMessages int                    `json:"total_messages"`
	ScansToday    int                    `json:"scans_today"`
	UserGrowth    []DashboardGrowthPoint `json:"user_growth"`
}

// GetDashboard godoc
// @Summary Dashboard statistics with user growth chart
// @Tags Admin
// @Security BearerAuth
// @Param days query int false "Days for growth chart" default(30)
// @Success 200 {object} DashboardResponse
// @Router /admin/dashboard [get]
func (h *AdminHandler) GetDashboard(c *fiber.Ctx) error {
	ctx := c.Context()

	days, _ := strconv.Atoi(c.Query("days", "30"))
	if days < 7 {
		days = 7
	}
	if days > 90 {
		days = 90
	}

	qrCounts, _ := h.qrRepo.CountByStatus(ctx)
	scansToday, _ := h.scanRepo.CountToday(ctx)
	msgCount, _ := h.msgRepo.CountAll(ctx)
	_, userTotal, _ := h.userRepo.List(ctx, 0, 1)

	totalQR := 0
	for _, v := range qrCounts {
		totalQR += v
	}

	// User growth per day
	rows, err := h.pool.Query(ctx,
		`SELECT DATE(created_at)::text AS day, COUNT(*)::int
		 FROM users
		 WHERE created_at >= NOW() - ($1 || ' days')::interval
		 GROUP BY day ORDER BY day`,
		strconv.Itoa(days),
	)
	var growth []DashboardGrowthPoint
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var p DashboardGrowthPoint
			_ = rows.Scan(&p.Date, &p.Count)
			growth = append(growth, p)
		}
	}
	if growth == nil {
		growth = []DashboardGrowthPoint{}
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": DashboardResponse{
			TotalUsers:     userTotal,
			TotalQRCodes:   totalQR,
			ActiveQR:       qrCounts["active"],
			UnregisteredQR: qrCounts["unregistered"],
			BlockedQR:      qrCounts["blocked"],
			TotalMessages:  msgCount,
			ScansToday:     scansToday,
			UserGrowth:     growth,
		},
	})
}

// DashboardChartPoint is one day in an admin chart series.
type DashboardChartPoint struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

// DashboardChartResponse is returned by GET /admin/dashboard/chart.
type DashboardChartResponse struct {
	Series string                 `json:"series"`
	Days   int                    `json:"days"`
	Points []DashboardChartPoint `json:"points"`
}

// GetDashboardChart godoc
// @Summary Time series for dashboard chart (clients / QR created / QR activated per day)
// @Description series: clients — new users per day; qr_created — qr_codes.created_at; qr_activated — qr_codes.registered_at (active claim).
// @Tags Admin
// @Security BearerAuth
// @Param days query int false "Window length in days" default(30)
// @Param series query string false "clients | qr_created | qr_activated" default(clients)
// @Success 200 {object} DashboardChartResponse
// @Router /admin/dashboard/chart [get]
func (h *AdminHandler) GetDashboardChart(c *fiber.Ctx) error {
	ctx := c.Context()

	days, _ := strconv.Atoi(c.Query("days", "30"))
	if days < 7 {
		days = 7
	}
	if days > 90 {
		days = 90
	}
	series := c.Query("series", "clients")

	var q string
	switch series {
	case "clients":
		q = `
WITH b AS (SELECT (CURRENT_DATE - ($1::int - 1))::date AS start_d, CURRENT_DATE::date AS end_d),
day_list AS (SELECT generate_series(b.start_d, b.end_d, '1 day'::interval)::date AS d FROM b),
agg AS (
  SELECT (created_at::date) AS d, COUNT(*)::bigint AS cnt
  FROM users
  WHERE created_at::date >= (SELECT start_d FROM b)
    AND created_at::date <= (SELECT end_d FROM b)
  GROUP BY 1
)
SELECT day_list.d::text, COALESCE(agg.cnt, 0)::int FROM day_list
LEFT JOIN agg ON agg.d = day_list.d ORDER BY day_list.d`
	case "qr_created":
		q = `
WITH b AS (SELECT (CURRENT_DATE - ($1::int - 1))::date AS start_d, CURRENT_DATE::date AS end_d),
day_list AS (SELECT generate_series(b.start_d, b.end_d, '1 day'::interval)::date AS d FROM b),
agg AS (
  SELECT (created_at::date) AS d, COUNT(*)::bigint AS cnt
  FROM qr_codes
  WHERE created_at::date >= (SELECT start_d FROM b)
    AND created_at::date <= (SELECT end_d FROM b)
  GROUP BY 1
)
SELECT day_list.d::text, COALESCE(agg.cnt, 0)::int FROM day_list
LEFT JOIN agg ON agg.d = day_list.d ORDER BY day_list.d`
	case "qr_activated":
		q = `
WITH b AS (SELECT (CURRENT_DATE - ($1::int - 1))::date AS start_d, CURRENT_DATE::date AS end_d),
day_list AS (SELECT generate_series(b.start_d, b.end_d, '1 day'::interval)::date AS d FROM b),
agg AS (
  SELECT (registered_at::date) AS d, COUNT(*)::bigint AS cnt
  FROM qr_codes
  WHERE registered_at IS NOT NULL
    AND registered_at::date >= (SELECT start_d FROM b)
    AND registered_at::date <= (SELECT end_d FROM b)
  GROUP BY 1
)
SELECT day_list.d::text, COALESCE(agg.cnt, 0)::int FROM day_list
LEFT JOIN agg ON agg.d = day_list.d ORDER BY day_list.d`
	default:
		return errorResponse(c, fiber.StatusBadRequest, "series must be clients, qr_created or qr_activated")
	}

	rows, err := h.pool.Query(ctx, q, days)
	if err != nil {
		h.logger.Sugar().Errorw("dashboard chart query", "err", err, "series", series)
		return errorResponse(c, fiber.StatusInternalServerError, "chart query failed")
	}
	defer rows.Close()

	var points []DashboardChartPoint
	for rows.Next() {
		var p DashboardChartPoint
		if err := rows.Scan(&p.Date, &p.Count); err != nil {
			return errorResponse(c, fiber.StatusInternalServerError, "chart scan failed")
		}
		points = append(points, p)
	}
	if points == nil {
		points = []DashboardChartPoint{}
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": DashboardChartResponse{
			Series: series,
			Days:   days,
			Points: points,
		},
	})
}

// ─── User detail ─────────────────────────────────────────────────────────────

type UserDetailResponse struct {
	types.User
	Vehicles     []types.Vehicle  `json:"vehicles"`
	Messages     []types.Message  `json:"messages"`
	QRCodes      []types.QRCode   `json:"qr_codes"`
	VehicleCount int              `json:"vehicle_count"`
}

// GetUserDetail godoc
// @Summary Get full user details with vehicles, messages, qr codes
// @Tags Admin
// @Security BearerAuth
// @Param id path string true "User UUID"
// @Success 200 {object} UserDetailResponse
// @Router /admin/users/{id}/detail [get]
func (h *AdminHandler) GetUserDetail(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "invalid user id")
	}
	ctx := c.Context()

	user, err := h.userRepo.GetByID(ctx, id)
	if err != nil || user == nil {
		return errorResponse(c, fiber.StatusNotFound, "user not found")
	}

	vehicles, _ := h.vehicleRepo.GetByUserID(ctx, id)
	if vehicles == nil {
		vehicles = []types.Vehicle{}
	}

	messages, _, _ := h.msgRepo.GetByUserID(ctx, id, 0, 50)
	if messages == nil {
		messages = []types.Message{}
	}

	qrCodes, _ := h.qrRepo.GetByUserID(ctx, id)
	if qrCodes == nil {
		qrCodes = []types.QRCode{}
	}

	return successResponse(c, UserDetailResponse{
		User:         *user,
		Vehicles:     vehicles,
		Messages:     messages,
		QRCodes:      qrCodes,
		VehicleCount: len(vehicles),
	})
}

// UpdateUser godoc
// @Summary Update user data
// @Tags Admin
// @Security BearerAuth
// @Param id path string true "User UUID"
// @Param body body types.UserUpdate true "User update"
// @Success 200 {object} DataResponse
// @Router /admin/users/{id} [patch]
func (h *AdminHandler) UpdateUser(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "invalid user id")
	}
	var upd types.UserUpdate
	if err := c.BodyParser(&upd); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "invalid request body")
	}
	user, err := h.userRepo.Update(c.Context(), id, upd)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	return successResponse(c, user)
}

// ListUserVehicles godoc
// @Summary List vehicles for a user
// @Tags Admin
// @Security BearerAuth
// @Param id path string true "User UUID"
// @Success 200 {object} DataResponse
// @Router /admin/users/{id}/vehicles [get]
func (h *AdminHandler) ListUserVehicles(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "invalid user id")
	}
	vehicles, err := h.vehicleRepo.GetByUserID(c.Context(), id)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	if vehicles == nil {
		vehicles = []types.Vehicle{}
	}
	return successResponse(c, vehicles)
}

// ListUserQRCodes godoc
// @Summary List QR codes for a user
// @Tags Admin
// @Security BearerAuth
// @Param id path string true "User UUID"
// @Success 200 {object} DataResponse
// @Router /admin/users/{id}/qrcodes [get]
func (h *AdminHandler) ListUserQRCodes(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "invalid user id")
	}
	codes, err := h.qrRepo.GetByUserID(c.Context(), id)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	if codes == nil {
		codes = []types.QRCode{}
	}
	return successResponse(c, codes)
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
