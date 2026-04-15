package client

import (
	"errors"
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
}

func NewQRHandler(
	qrSvc *services.QRService,
	apiSvc *services.APISpecService,
	msgSvc *services.MessageService,
	jwtMgr *jwtpkg.Manager,
) *QRHandler {
	return &QRHandler{qrSvc: qrSvc, apiSvc: apiSvc, msgSvc: msgSvc, jwtMgr: jwtMgr}
}

// GetQR godoc
// @Summary Get QR code info
// @Description Returns the owner profile and vehicle info for a registered QR code, or {"registered":false} if not yet registered.
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

// registerBody accepts only the phone number.
// After registration the owner receives a JWT and can fill in the rest
// via PATCH /me and PATCH /me/vehicles/:id.
type registerBody struct {
	Phone string `json:"phone"`
}

const ownerTokenTTL = 24 * time.Hour

// RegisterQR godoc
// @Summary Register a QR code
// @Description Claims a QR code for the given phone number.
// @Description If a user with that phone already exists they are linked to this QR.
// @Description Sets an owner_token cookie (24 h) and returns the token in the response body.
// @Tags QR
// @Accept json
// @Produce json
// @Param qr_id path string true "QR code"
// @Param body body registerBody true "Phone number"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /qr/{qr_id} [post]
func (h *QRHandler) RegisterQR(c *fiber.Ctx) error {
	var body registerBody
	if err := c.BodyParser(&body); err != nil {
		return specError(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "Некорректное тело запроса", nil, nil)
	}
	if strings.TrimSpace(body.Phone) == "" {
		return specError(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "Заполните обязательные поля",
			map[string]string{"phone": "Обязательное поле"}, nil)
	}

	userID, isNewUser, err := h.apiSvc.Register(c.Context(), services.QRRegisterInput{
		QRID:  c.Params("qr_id"),
		Phone: body.Phone,
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
		"token":       token,
		"user_id":     userID,
		"is_new_user": isNewUser,
	})
}

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
// @Router /qr/{qr_id}/message [post]
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
	if errors.Is(err, services.ErrMessageQRUnavailable) {
		return specError(c, fiber.StatusNotFound, "NOT_FOUND", "QR код не найден или не зарегистрирован", nil, nil)
	}
	if err != nil {
		return specInternal(c)
	}
	return specMessageOK(c, "Сообщение отправлено")
}

// GetQRImage godoc
// @Summary Get QR code as PNG image
// @Description Returns a PNG image of the QR code encoding the scan URL.
// @Tags QR
// @Produce image/png
// @Param qr_id path string true "QR code"
// @Param size query int false "Image size in pixels" default(256)
// @Success 200 {file} binary
// @Failure 500 {object} map[string]interface{}
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
