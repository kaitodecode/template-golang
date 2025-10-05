package internal

import (
	user_handler "template-golang/internal/features/users/handler"
	"template-golang/pkg/fileUploader"
	"template-golang/pkg/middleware"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/swagger" // swagger handler
)

func NewUtschoolApp(
	userHandler *user_handler.Handler,
) *fiber.App {

	app := fiber.New(fiber.Config{
		ServerHeader: "Fiber",
		JSONEncoder:  json.Marshal,
		JSONDecoder:  json.Unmarshal,
		ErrorHandler: middleware.ErrorHandlerFunc,
		BodyLimit:    30 * 1024 * 1024, // 20 MB limit
	})
	app.Use(recover.New())

	app.Use(middleware.RequestIDMiddleware())
	app.Use(middleware.LoggerMiddleware())
	app.Use(middleware.CorsMiddleware())

	app.Get("/swagger/*", swagger.New(swagger.Config{
		DocExpansion: "none",
	}))
	app.Get("/swagger/doc.json", func(c *fiber.Ctx) error {
		return c.SendFile("./swagger/swagger.json")
	})

	if err := fileUploader.InitS3Client(); err != nil {
		panic(err)
	}

	api := app.Group("/api/v1")
	userHandler.RegisterRoutes(api)

	api.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
		})
	})

	return app
}
