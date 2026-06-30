package httpserver

import (
	"time"

	"report-service-m/internal/config"

	"github.com/gofiber/fiber/v2"
)

func registerRoutes(app *fiber.App, cfg config.Config, handlers reportHandlers) {
	app.Get("/", serviceInfo(cfg))
	app.Get("/health", health(cfg))
	app.Get("/ready", health(cfg))
	app.Get("/report/payslip/html", handlers.previewPayslipHTML)
	app.Get("/report/payslip/pdf", handlers.previewPayslipHTMLPDF)

	v1 := app.Group("/v1")
	v1.Get("/health", health(cfg))
}

func serviceInfo(cfg config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"service":     cfg.AppName,
			"environment": cfg.Environment,
			"routes": []string{
				"GET /health",
				"GET /report/payslip/html",
				"GET /report/payslip/pdf",
			},
		})
	}
}

func health(cfg config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"ok":          true,
			"service":     cfg.AppName,
			"environment": cfg.Environment,
			"time":        time.Now().UTC().Format(time.RFC3339),
		})
	}
}
