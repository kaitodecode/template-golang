package users

import (
	"template-golang/internal/features/users/handler"
	"template-golang/internal/features/users/service"

	"github.com/google/wire"
)

var Set = wire.NewSet(
	service.NewService,
	handler.NewHandler,
)
