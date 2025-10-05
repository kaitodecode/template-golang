package middleware

import (
	"context"

	"template-golang/internal/db/model"
	"template-golang/pkg/helper"
	"template-golang/pkg/response"

	"github.com/gofiber/fiber/v2"
)

func AuthMiddleware(roles *[]string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Ambil token dari header
		tokenString, err := helper.GetTokenFromHeader(c)
		if err != nil {
			return response.Json(c.Status(fiber.StatusUnauthorized), err.Error(), "Unauthorized")
		}

		// Verifikasi token
		claims, err := helper.VerifyJwtToken(tokenString)
		if err != nil {
			return response.Json(c.Status(fiber.StatusUnauthorized), err.Error(), "Unauthorized")
		}

		// Extract data dari claims (asumsikan helper.VerifyJwtToken return jwt.MapClaims / custom struct)
		userID := claims["user_id"].(string) // atau int, sesuai isi token
		user := claims["user"].(model.User)
		role := claims["role"].(string)

		// Simpan ke context
		c.Locals("user_id", userID)
		c.Locals("user", user)
		c.Locals("token", tokenString)
		c.Locals("role", role)

		// Store the same values in context.Context so they can be retrieved via ctx.Value() later
		ctx := context.WithValue(c.Context(), "user_id", userID)
		ctx = context.WithValue(ctx, "user", user)
		ctx = context.WithValue(ctx, "token", tokenString)
		ctx = context.WithValue(ctx, "role", role)
		c.SetUserContext(ctx)

		// Role-based check
		if len(*roles) > 0 {
			if !containsRole(*roles, role) {
				return response.Json(c.Status(fiber.StatusForbidden), nil, "Forbidden")
			}
		}

		return c.Next()
	}
}

func containsRole(roles []string, role string) bool {
	for _, v := range roles {
		if v == role {
			return true
		}
	}
	return false
}
