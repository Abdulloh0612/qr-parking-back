package middleware

import (
	"strings"

	jwtpkg "qr-parking/pkg/jwt"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// OwnerTokenCookie is the cookie name used for client (QR owner) sessions.
const OwnerTokenCookie = "owner_token"

func JWTAuth(jwtManager *jwtpkg.Manager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		tokenStr := ""

		// Prefer Authorization header; fall back to owner_token cookie.
		if auth := c.Get("Authorization"); auth != "" {
			parts := strings.SplitN(auth, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid authorization header format"})
			}
			tokenStr = parts[1]
		} else if cookie := c.Cookies(OwnerTokenCookie); cookie != "" {
			tokenStr = cookie
		}

		if tokenStr == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "missing authorization"})
		}

		claims, err := jwtManager.ParseToken(tokenStr)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid or expired token"})
		}

		c.Locals("userID", claims.UserID)
		c.Locals("isAdmin", claims.IsAdmin)
		c.Locals("adminSession", claims.AdminSession)
		return c.Next()
	}
}

// AdminOnly требует вход в админку по логину/паролю (JWT с admin_session=true).
func AdminOnly() fiber.Handler {
	return func(c *fiber.Ctx) error {
		adminSession, ok := c.Locals("adminSession").(bool)
		if !ok || !adminSession {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "admin password login required"})
		}
		isAdmin, ok := c.Locals("isAdmin").(bool)
		if !ok || !isAdmin {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "admin access required"})
		}
		return c.Next()
	}
}

func GetUserID(c *fiber.Ctx) uuid.UUID {
	userID, _ := c.Locals("userID").(uuid.UUID)
	return userID
}
