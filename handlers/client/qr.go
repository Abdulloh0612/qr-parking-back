package client

import (
	"strconv"
	"strings"
	"time"

	"qr-parking/middleware"
	jwtpkg "qr-parking/pkg/jwt"
	"qr-parking/services"

	"github.com/gofiber/fiber/v2"
)

type QRHandler struct {
	qrSvc  *services.QRService
	apiSvc *services.APISpecService
	msgSvc *services.MessageService
	jwtMgr *jwtpkg.Manager
	otpSvc *services.OTPService
}

func NewQRHandler(
	qrSvc *services.QRService,
	apiSvc *services.APISpecService,
	msgSvc *services.MessageService,
	jwtMgr *jwtpkg.Manager,
	otpSvc *services.OTPService,
) *QRHandler {
	return &QRHandler{
		qrSvc:  qrSvc,
		apiSvc: apiSvc,
		msgSvc: msgSvc,
		jwtMgr: jwtMgr,
		otpSvc: otpSvc,
	}
}

// GetQR godoc
// @Summary Get QR code info
// @Description Returns the full owner profile and vehicle linked to this QR code.
// @Description Returns {"registered":false} if the code is not yet claimed.
// @Tags QR
// @Produce json
// @Param qr_id path string true "QR code"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /qr/{qr_id} [get]
func (h *QRHandler) GetQR(c *fiber.Ctx) error {
	result, err := h.apiSvc.GetQRInfo(c.Context(), c.Params("qr_id"), c.IP(), c.Get("User-Agent"))
	if err != nil {
		if services.IsSpecNotFound(err) {
			return specError(c, fiber.StatusNotFound, "NOT_FOUND", err.Error(), nil, nil)
		}
		return specInternal(c)
	}
	return specSuccess(c, result)
}

// ─── Step 1: send OTP ─────────────────────────────────────────────────────────

type registerBody struct {
	Phone string `json:"phone"`
}

// RegisterQR godoc
// @Summary Send OTP to phone number
// @Description Validates the QR code and sends a 6-digit OTP to the given phone number.
// @Description The OTP is valid for 5 minutes.
// @Tags QR
// @Accept json
// @Produce json
// @Param qr_id path string true "QR code"
// @Param body body registerBody true "Phone number"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /qr/{qr_id} [post]
func (h *QRHandler) RegisterQR(c *fiber.Ctx) error {
	var body registerBody
	if err := c.BodyParser(&body); err != nil {
		return specError(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "Некорректное тело запроса", nil, nil)
	}
	phone := strings.TrimSpace(body.Phone)
	if phone == "" {
		return specError(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "Заполните обязательные поля",
			map[string]string{"phone": "Обязательное поле"}, nil)
	}

	if err := h.otpSvc.Send(c.Context(), phone); err != nil {
		return specInternal(c)
	}

	return specSuccess(c, fiber.Map{"message": "OTP отправлен на " + phone})
}

// ─── Step 2: verify OTP + register ───────────────────────────────────────────

type verifyBody struct {
	Phone string `json:"phone"`
	OTP   string `json:"otp"`
}

const ownerTokenTTL = 30 * 24 * time.Hour // 30 days

// VerifyQR godoc
// @Summary Verify OTP and register the QR code
// @Description Verifies the OTP, creates or links the user by phone,
// @Description then returns an access token valid for 30 days.
// @Description If the QR is already registered: owner phone → owner token; else if phone exists in DB → that account's token.
// @Tags QR
// @Accept json
// @Produce json
// @Param qr_id path string true "QR code"
// @Param body body verifyBody true "Phone + OTP"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /qr-verify/{qr_id} [post]
func (h *QRHandler) VerifyQR(c *fiber.Ctx) error {
	var body verifyBody
	if err := c.BodyParser(&body); err != nil {
		return specError(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "Некорректное тело запроса", nil, nil)
	}
	phone := strings.TrimSpace(body.Phone)
	otp := strings.TrimSpace(body.OTP)
	if phone == "" || otp == "" {
		return specError(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "Заполните обязательные поля",
			map[string]string{"phone": "Обязательное поле", "otp": "Обязательное поле"}, nil)
	}

	valid, err := h.otpSvc.Verify(c.Context(), phone, otp)
	if err != nil {
		return specInternal(c)
	}
	if !valid {
		return specError(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "Неверный или истёкший OTP",
			map[string]string{"otp": "Неверный код"}, nil)
	}

	userID, isNewUser, err := h.apiSvc.Register(c.Context(), services.QRRegisterInput{
		QRID:  c.Params("qr_id"),
		Phone: phone,
	})
	if err != nil {
		if services.IsSpecNotFound(err) {
			return specError(c, fiber.StatusNotFound, "NOT_FOUND", err.Error(), nil, nil)
		}
		if services.IsSpecValidation(err) {
			return specError(c, fiber.StatusBadRequest, "VALIDATION_ERROR", err.Error(), services.SpecErrorFields(err), nil)
		}
		return specInternal(c)
	}

	token, err := h.jwtMgr.GenerateToken(userID, false, false, ownerTokenTTL)
	if err != nil {
		return specInternal(c)
	}

	c.Cookie(&fiber.Cookie{
		Name:     middleware.OwnerTokenCookie,
		Value:    token,
		Expires:  time.Now().Add(ownerTokenTTL),
		HTTPOnly: true,
		SameSite: "Lax",
		Path:     "/",
	})

	return specSuccessCreated(c, fiber.Map{
		"access_token": token,
		"user_id":      userID,
		"is_new_user":  isNewUser,
	})
}

// ─── QR message ───────────────────────────────────────────────────────────────

type messageBody struct {
	Message string `json:"message"`
}

// PostQRMessage godoc
// @Summary Send a message to the QR owner
// @Description Saves the message and sends a Telegram notification to the owner if they have the bot linked.
// @Tags QR
// @Accept json
// @Produce json
// @Param qr_id path string true "QR code"
// @Param body body messageBody true "Message text (1–500 chars)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /qr-message/{qr_id} [post]
func (h *QRHandler) PostQRMessage(c *fiber.Ctx) error {
	var body messageBody
	if err := c.BodyParser(&body); err != nil {
		return specError(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "Некорректное тело запроса", nil, nil)
	}

	msg := strings.TrimSpace(body.Message)
	if msg == "" || len([]rune(msg)) > 500 {
		return specError(c, fiber.StatusBadRequest, "VALIDATION_ERROR",
			"Сообщение должно быть от 1 до 500 символов",
			map[string]string{"message": "Некорректное сообщение"}, nil)
	}

	_, err := h.msgSvc.SendMessageByQRRef(c.Context(), c.Params("qr_id"), msg)
	if err != nil {
		if err == services.ErrMessageQRUnavailable {
			return specError(c, fiber.StatusNotFound, "NOT_FOUND", "QR код не найден или не зарегистрирован", nil, nil)
		}
		return specInternal(c)
	}
	return specMessageOK(c, "Сообщение отправлено")
}

// ─── QR image ─────────────────────────────────────────────────────────────────

// GetQRImage godoc
// @Summary Get QR code as PNG image
// @Tags QR
// @Produce image/png
// @Param qr_id path string true "QR code"
// @Param size query int false "Image size in pixels" default(256)
// @Success 200 {file} binary
// @Router /qr/{qr_id}/image [get]
func (h *QRHandler) GetQRImage(c *fiber.Ctx) error {
	size, _ := strconv.Atoi(c.Query("size", "256"))
	if size <= 0 {
		size = 256
	}
	png, err := h.qrSvc.GetQRImage(c.Context(), c.Params("qr_id"), size)
	if err != nil {
		return specInternal(c)
	}
	c.Set("Content-Type", "image/png")
	return c.Send(png)
}
