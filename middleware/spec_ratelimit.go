package middleware

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

// APISpecMessageRateLimit enforces 1 POST /qr/:id/message per IP per minute (api-spec).
func APISpecMessageRateLimit(redisClient *redis.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		key := fmt.Sprintf("rate:apispec_qr_msg:%s", c.IP())
		count, err := redisClient.Incr(c.Context(), key).Result()
		if err != nil {
			return c.Next()
		}
		if count == 1 {
			redisClient.Expire(c.Context(), key, time.Minute)
		}
		if count > 1 {
			retry := 60
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"success": false,
				"error": fiber.Map{
					"code":        "RATE_LIMIT",
					"message":     "Слишком много запросов. Попробуйте через 1 минуту",
					"retry_after": retry,
				},
			})
		}
		return c.Next()
	}
}
