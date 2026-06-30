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
	app.Post("/report/payslip", handlers.renderPayslipReport)
	app.Get("/report/payslip/html", handlers.previewPayslipHTML)
	app.Post("/report/payslip/html", handlers.renderPayslipHTML)
	app.Get("/report/payslip/pdf", handlers.previewPayslipHTMLPDF)
	app.Post("/report/payslip/pdf", handlers.renderPayslipHTMLPDF)

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
				"POST /report/payslip",
				"GET /report/payslip/html",
				"POST /report/payslip/html",
				"GET /report/payslip/pdf",
				"POST /report/payslip/pdf",
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
