package admin

import (
	"qr-parking/pkg/validator"
	"qr-parking/services"

	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

type AdminLoginRequest struct {
	Username string `json:"username" validate:"required" example:"admin"`
	Password string `json:"password" validate:"required" example:"password123"`
}

type AdminLoginResponse struct {
	Tokens       *TokensResponse `json:"tokens"`
	AdminSession bool            `json:"admin_session" example:"true"`
}

type TokensResponse struct {
	AccessToken  string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIs..."`
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIs..."`
}

// AdminLogin godoc
// @Summary Admin login
// @Description Authenticates admin with username/password and returns JWT with admin_session=true
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body AdminLoginRequest true "Admin credentials"
// @Success 200 {object} AdminLoginResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 429 {object} ErrorResponse
// @Router /admin/login [post]
func (h *AuthHandler) AdminLogin(c *fiber.Ctx) error {
	var req AdminLoginRequest
	if err := c.BodyParser(&req); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "invalid request body")
	}
	if err := validator.ValidateStruct(req); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	tokens, err := h.authService.AdminLogin(c.Context(), req.Username, req.Password)
	if err != nil {
		return errorResponse(c, fiber.StatusUnauthorized, err.Error())
	}

	return c.JSON(fiber.Map{"tokens": tokens, "admin_session": true})
}
