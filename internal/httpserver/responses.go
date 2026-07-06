package httpserver

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

type errorResponse struct {
	Error string `json:"error"`
}

func errorJSON(c *fiber.Ctx, status int, message string) error {
	if message == "" {
		message = "request failed"
	}

	return c.Status(status).JSON(errorResponse{Error: message})
}

func sendPDF(c *fiber.Ctx, pdfBytes []byte, filename, disposition string) error {
	if disposition == "" {
		disposition = "attachment"
	}

	filename = safeFilename(filename, "report.pdf")

	c.Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	c.Set("Pragma", "no-cache")
	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", disposition+`; filename="`+filename+`"`)

	return c.Send(pdfBytes)
}

func sendXLSX(c *fiber.Ctx, xlsxBytes []byte, filename string) error {
	filename = safeFilename(filename, "report.xlsx")

	c.Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	c.Set("Pragma", "no-cache")
	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", `attachment; filename="`+filename+`"`)

	return c.Send(xlsxBytes)
}

func safeFilename(filename, fallback string) string {
	filename = strings.TrimSpace(filename)
	if filename == "" {
		filename = fallback
	}

	replacer := strings.NewReplacer(
		"/", "_",
		"\\", "_",
		`"`, "",
		"\n", "_",
		"\r", "_",
	)

	return replacer.Replace(filename)
}
