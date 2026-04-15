package client

import "github.com/gofiber/fiber/v2"

func specSuccess(c *fiber.Ctx, data interface{}) error {
	return c.JSON(fiber.Map{"success": true, "data": data})
}

func specSuccessCreated(c *fiber.Ctx, data interface{}) error {
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"success": true, "data": data})
}

func specMessageOK(c *fiber.Ctx, msg string) error {
	return c.JSON(fiber.Map{"success": true, "message": msg})
}

func specError(c *fiber.Ctx, status int, code, message string, fields map[string]string, retryAfter *int) error {
	errObj := fiber.Map{"code": code, "message": message}
	if fields != nil {
		errObj["fields"] = fields
	}
	if retryAfter != nil {
		errObj["retry_after"] = *retryAfter
	}
	return c.Status(status).JSON(fiber.Map{"success": false, "error": errObj})
}

func specInternal(c *fiber.Ctx) error {
	return specError(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "Что-то пошло не так. Попробуйте позже", nil, nil)
}
