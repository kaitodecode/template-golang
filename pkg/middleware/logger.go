package middleware

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	fiberLogger "github.com/gofiber/fiber/v2/middleware/logger"
)

func LoggerMiddleware() fiber.Handler {
	return fiberLogger.New(fiberLogger.Config{
		Format:     "${time} | ${requestID} | ${status} | ${latency} | ${ip} | ${method} | ${path} | ${error}\n",
		TimeFormat: "15:04:05",
		TimeZone:   "Asia/Jakarta",
		CustomTags: map[string]fiberLogger.LogFunc{
			"latency": func(output fiberLogger.Buffer, c *fiber.Ctx, data *fiberLogger.Data, extraParam string) (int, error) {
				latency := data.Stop.Sub(data.Start)
			var formatted string
			switch {
			case latency < time.Microsecond:
				formatted = fmt.Sprintf("%dns", latency.Nanoseconds())
			case latency < time.Millisecond:
				formatted = fmt.Sprintf("%dµs", latency.Microseconds())
			case latency < time.Second:
				formatted = fmt.Sprintf("%dms", latency.Milliseconds())
			default:
				formatted = fmt.Sprintf("%.2fs", latency.Seconds())
			}

			return output.WriteString(formatted)
		},
		"requestID": func(output fiberLogger.Buffer, c *fiber.Ctx, data *fiberLogger.Data, extraParam string) (int, error) {
			requestID := c.Locals("requestID")
			return output.WriteString(requestID.(string))
		},
	},
})
}