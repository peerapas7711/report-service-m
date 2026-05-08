package httpserver

import (
	"report-service-m/internal/config"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func New(cfg config.Config) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:      cfg.AppName,
		BodyLimit:    cfg.BodyLimit,
		ErrorHandler: handleError,
	})

	app.Use(recover.New())
	registerRoutes(app, cfg)

	return app
}

func handleError(c *fiber.Ctx, err error) error {
	status := fiber.StatusInternalServerError
	if fiberErr, ok := err.(*fiber.Error); ok {
		status = fiberErr.Code
	}

	return errorJSON(c, status, err.Error())
}
