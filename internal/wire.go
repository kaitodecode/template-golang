//go:build wireinject
// +build wireinject

package internal

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/wire"

	"template-golang/internal/db"
	"template-golang/internal/features/base"
	"template-golang/internal/features/users"

	"template-golang/pkg/redisx"
)

func InitApp() (*fiber.App, error) {
	wire.Build(
		db.ConnectDB,
		redisx.New,
		base.Set,
		users.Set,
		NewUtschoolApp,
	)
	return nil, nil
}
