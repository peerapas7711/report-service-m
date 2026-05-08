package httpserver

import (
	"time"

	"report-service-m/internal/config"

	"github.com/gofiber/fiber/v2"
)

func registerRoutes(app *fiber.App, cfg config.Config) {
	app.Get("/", serviceInfo(cfg))
	app.Get("/health", health(cfg))
	app.Get("/payslip", payslipPage)
	app.Get("/ready", health(cfg))

	v1 := app.Group("/v1")
	v1.Get("/health", health(cfg))

	reports := v1.Group("/reports")
	reports.Post("/payslip/render", renderPayslip)
	reports.Post("/system-access-permission/render", renderSystemAccessPermission)

	preview := app.Group("/preview")
	preview.Get("/payslip", previewPayslip)
	preview.Get("/system-access", previewSystemAccessPermission)
}

func serviceInfo(cfg config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"service":     cfg.AppName,
			"environment": cfg.Environment,
			"routes": []string{
				"GET /health",
				"GET /payslip",
				"POST /v1/reports/payslip/render",
				"POST /v1/reports/system-access-permission/render",
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
