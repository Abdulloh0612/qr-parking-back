package client

import (
	"strings"

	"qr-parking/middleware"
	"qr-parking/services"
	"qr-parking/types"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type MeHandler struct {
	apiSvc *services.APISpecService
}

func NewMeHandler(apiSvc *services.APISpecService) *MeHandler {
	return &MeHandler{apiSvc: apiSvc}
}

// GetMe godoc
// @Summary Get own profile
// @Description Returns the authenticated owner's full profile including vehicles and socials.
// @Tags Me
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /me [get]
func (h *MeHandler) GetMe(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	data, err := h.apiSvc.GetOwnerData(c.Context(), userID)
	if err != nil {
		if services.IsSpecNotFound(err) {
			return specError(c, fiber.StatusNotFound, "NOT_FOUND", err.Error(), nil, nil)
		}
		return specInternal(c)
	}
	return specSuccess(c, data)
}

type patchMeBody struct {
	FirstName *string `json:"first_name"`
	LastName  *string `json:"last_name"`
	AvatarURL *string `json:"avatar_url"`
	// Socials: omit = skip, "" = delete, "value" = upsert
	WhatsApp  *string `json:"whatsapp"`
	Instagram *string `json:"instagram"`
	Telegram  *string `json:"telegram"`
	VK        *string `json:"vk"`
	Facebook  *string `json:"facebook"`
}

// PatchMe godoc
// @Summary Update own profile
// @Description Updates the authenticated owner's name, avatar and social profiles.
// @Description Send null/omit a field to leave it unchanged; send "" to delete a social link.
// @Tags Me
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body patchMeBody true "Fields to update"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /me [patch]
func (h *MeHandler) PatchMe(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)

	var body patchMeBody
	if err := c.BodyParser(&body); err != nil {
		return specError(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "Некорректное тело запроса", nil, nil)
	}

	// Trim string pointer helpers inline
	trimPtr := func(s *string) *string {
		if s == nil {
			return nil
		}
		v := strings.TrimSpace(*s)
		return &v
	}

	data, err := h.apiSvc.UpdateOwnerProfile(c.Context(), userID, services.OwnerProfileInput{
		FirstName: trimPtr(body.FirstName),
		LastName:  trimPtr(body.LastName),
		AvatarURL: trimPtr(body.AvatarURL),
		WhatsApp:  body.WhatsApp,
		Instagram: body.Instagram,
		Telegram:  body.Telegram,
		VK:        body.VK,
		Facebook:  body.Facebook,
	})
	if err != nil {
		if services.IsSpecNotFound(err) {
			return specError(c, fiber.StatusNotFound, "NOT_FOUND", err.Error(), nil, nil)
		}
		return specInternal(c)
	}
	return specSuccess(c, data)
}

// GetVehicles godoc
// @Summary List own vehicles
// @Description Returns all vehicles belonging to the authenticated owner.
// @Tags Me
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /me/vehicles [get]
func (h *MeHandler) GetVehicles(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	data, err := h.apiSvc.GetOwnerData(c.Context(), userID)
	if err != nil {
		if services.IsSpecNotFound(err) {
			return specError(c, fiber.StatusNotFound, "NOT_FOUND", err.Error(), nil, nil)
		}
		return specInternal(c)
	}
	return specSuccess(c, data.Vehicles)
}

type patchVehicleBody struct {
	PlateNumber     *string `json:"plate_number"`
	CarModel        *string `json:"car_model"`
	PhotoURL        *string `json:"photo_url"`
	TelegramEnabled *bool   `json:"telegram_enabled"`
}

// PatchVehicle godoc
// @Summary Update a vehicle
// @Description Updates a vehicle that belongs to the authenticated owner.
// @Description Omit a field to leave it unchanged.
// @Tags Me
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Vehicle UUID"
// @Param body body patchVehicleBody true "Fields to update"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /me/vehicles/{id} [patch]
func (h *MeHandler) PatchVehicle(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)

	vehicleID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return specError(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "Некорректный идентификатор автомобиля", nil, nil)
	}

	var body patchVehicleBody
	if err := c.BodyParser(&body); err != nil {
		return specError(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "Некорректное тело запроса", nil, nil)
	}

	veh, err := h.apiSvc.UpdateOwnerVehicle(c.Context(), userID, vehicleID, types.VehicleUpdate{
		PlateNumber:     body.PlateNumber,
		CarModel:        body.CarModel,
		PhotoURL:        body.PhotoURL,
		TelegramEnabled: body.TelegramEnabled,
	})
	if err != nil {
		if services.IsSpecNotFound(err) {
			return specError(c, fiber.StatusNotFound, "NOT_FOUND", err.Error(), nil, nil)
		}
		if services.IsSpecValidation(err) {
			return specError(c, fiber.StatusForbidden, "FORBIDDEN", err.Error(), nil, nil)
		}
		return specInternal(c)
	}
	return specSuccess(c, veh)
}

