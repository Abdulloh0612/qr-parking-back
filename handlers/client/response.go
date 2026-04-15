package client

import "github.com/gofiber/fiber/v2"

type ErrorResponse struct {
	Error string `json:"error" example:"error message"`
}

type MessageResponse struct {
	Message string `json:"message" example:"operation successful"`
}

type DataResponse struct {
	Data interface{} `json:"data"`
}

type PaginatedResponse struct {
	Data  interface{} `json:"data"`
	Total int         `json:"total" example:"100"`
	Page  int         `json:"page" example:"1"`
	Limit int         `json:"limit" example:"20"`
}

func errorResponse(c *fiber.Ctx, status int, msg string) error {
	return c.Status(status).JSON(ErrorResponse{Error: msg})
}

func successResponse(c *fiber.Ctx, data interface{}) error {
	return c.JSON(DataResponse{Data: data})
}

func paginatedResponse(c *fiber.Ctx, data interface{}, total, page, limit int) error {
	return c.JSON(PaginatedResponse{
		Data:  data,
		Total: total,
		Page:  page,
		Limit: limit,
	})
}
