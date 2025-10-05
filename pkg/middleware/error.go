package middleware

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"

	"template-golang/pkg/apperror"
	"template-golang/pkg/helper"
	"template-golang/pkg/logger"
	"template-golang/pkg/response"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

var Env = os.Getenv("SERVER_ENV") // bisa "dev" atau "prod"

func ErrorHandlerFunc(c *fiber.Ctx, err error) error {
	// ----------- 1. Default values -----------
	if errors.Is(err, fiber.ErrRequestEntityTooLarge) {
		return response.Json(c.Status(400), nil, "File size exceeds maximum limit")
	}
	statusCode := fiber.StatusInternalServerError
	message := "Internal Server Error"

	// ----------- 2. Error khusus Fiber (404, dll.) -----------
	if e, ok := err.(*fiber.Error); ok {
		statusCode = e.Code
		message = e.Message

	}

	// ----------- 3. Error khusus app -----------
	if appErr, ok := err.(*apperror.AppError); ok {
		logger.L().Infof("AppError: %v", appErr)
		statusCode = appErr.StatusCode
		message = appErr.Message

		// Jika validation error (422)
		if statusCode == fiber.StatusUnprocessableEntity {
			return response.Json(c.Status(statusCode), appErr.Detail, message)
		}
		return response.Json(c.Status(statusCode), err, message)
	}

	if err == sql.ErrNoRows {
		statusCode = fiber.StatusNotFound
		message = "Resource not found"
		return response.Json(c.Status(statusCode), nil, message)
	}

	// Handle GORM errors
	if errors.Is(err, gorm.ErrRecordNotFound) {
		statusCode = fiber.StatusNotFound
		message = "Resource not found"
		return response.Json(c.Status(statusCode), nil, message)
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505": // unique_violation
			statusCode = fiber.StatusConflict
			// pgErr.Detail is like: "Key (column)=(value) already exists."
			detail := pgErr.Detail
			var column, value string
			if strings.HasPrefix(detail, "Key (") {
				after := strings.TrimPrefix(detail, "Key (")
				parts := strings.SplitN(after, ")=(", 2)
				if len(parts) == 2 {
					column = parts[0]
					remain := parts[1]
					if end := strings.Index(remain, ") already exists."); end != -1 {
						value = remain[:end]
					}
					// remove unique key prefix if any
					column = strings.TrimPrefix(column, "uk_")
				}
			}
			if column != "" && value != "" {
				message = fmt.Sprintf("%s '%s' already exists", helper.FormatWord(column), value)
			} else {
				message = detail
			}
		default:
			statusCode = fiber.StatusBadRequest
			message = pgErr.Message
		}
		return response.Json(c.Status(statusCode), nil, message)
	}

	// ----------- 4. Error khusus lain (misal file terlalu besar) -----------
	if err.Error() == "Request Entity Too Large" {
		statusCode = fiber.StatusRequestEntityTooLarge
		message = "File size exceeds maximum limit"
		return response.Json(c.Status(fiber.StatusRequestEntityTooLarge), nil, message)
	}

	// ----------- 5. Ambil requestID aman -----------
	requestID, _ := c.Locals("requestID").(string)
	if requestID == "" {
		requestID = "unknown"
	}

	// ----------- 6. Buat log entry -----------
	now := time.Now()
	stack := string(debug.Stack())

	logEntry := map[string]interface{}{
		"timestamp":   now.Format(time.RFC3339),
		"level":       "ERROR",
		"message":     err.Error(),
		"request_id":  requestID,
		"method":      c.Method(),
		"path":        c.OriginalURL(),
		"ip":          c.IP(),
		"user_agent":  c.Get("User-Agent"),
		"status_code": statusCode,
		"stack_trace": stack,
	}

	// ----------- 7. Tulis log ke file (per bulan) -----------
	yearDir := fmt.Sprintf("%d", now.Year())
	monthFile := fmt.Sprintf("%02d.jsonl", now.Month()) // JSONL: satu baris = satu log
	logDir := filepath.Join("internal", "logs", yearDir)
	_ = os.MkdirAll(logDir, 0755)
	logPath := filepath.Join(logDir, monthFile)

	f, _ := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer f.Close()
	logJSON, _ := json.Marshal(logEntry)
	f.WriteString(string(logJSON) + "\n")

	// ----------- 8. Print ke terminal -----------
	log.Errorf("[ERROR] %s | %s %s | %s", now.Format(time.RFC3339), c.Method(), c.Path(), err.Error())

	// ----------- 9. Response ke client -----------
	if Env == "dev" {
		// Dev mode -> kirim stack trace
		return response.Json(c.Status(statusCode), map[string]any{
			"request_id": requestID,
			"stack":      stack,
		}, message)
	}
	// Prod mode -> jangan bocorkan stack trace
	return response.Json(c.Status(statusCode), map[string]any{
		"request_id": requestID,
	}, message)
}

func splitOnce(s, sep string) []string {
	idx := 0
	for i := 0; i < len(s)-len(sep)+1; i++ {
		if s[i:i+len(sep)] == sep {
			idx = i
			break
		}
	}
	if idx == 0 {
		return []string{s}
	}
	return []string{s[:idx], s[idx+len(sep):]}
}
