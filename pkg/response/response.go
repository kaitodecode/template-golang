package response

import (
	"github.com/gofiber/fiber/v2"
)

type BaseResponse struct {
	Status  bool        `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type Response[T any] struct {
	Status  bool        `json:"status"`
	Message string      `json:"message"`
	Data    T           `json:"data"`
}


type ErrorResponse struct {
	Status  bool        `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type SuccessResponse struct {
	Status  bool        `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// Success returns a successful response with data
func Success(c *fiber.Ctx, data interface{}) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  true,
		"message": "Success",
		"data":    data,
	})
}

// Error returns an error response with message
func Error(c *fiber.Ctx, message string, err interface{}) error {
	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
		"status":  false,
		"message": message,
		"data":    err,
	})
}

func Json(c *fiber.Ctx, data interface{}, message string) error {
	return c.JSON(fiber.Map{
		"status":  c.Response().StatusCode() >= 200 && c.Response().StatusCode() <= 299,
		"message": message,
		"data":    data,
	})
}

// ServerError returns a 500 internal server error response
func ServerError(c *fiber.Ctx, err error) error {
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"status":  false,
		"message": err.Error(),
		"data":    nil,
	})
}

// NotFound returns a 404 not found response
func NotFound(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
		"status":  false,
		"message": "Resource not found",
		"data":    nil,
	})
}
