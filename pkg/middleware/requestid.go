package middleware

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/nrednav/cuid2"
	"time"
)

func RequestIDMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		reqID := c.Get(fiber.HeaderXRequestID) // coba ambil dari client dulu

		if reqID == "" {
			reqID = fmt.Sprintf("req-%02d-%04d-%s", time.Now().Month(), time.Now().Year(), cuid2.Generate())
		}

		// simpan ke response header biar client bisa lihat
		c.Set(fiber.HeaderXRequestID, reqID)

		// simpan ke context (Locals) supaya handler bisa ambil
		c.Locals("requestID", reqID)

		return c.Next()
	}
}
