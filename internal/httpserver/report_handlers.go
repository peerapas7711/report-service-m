package httpserver

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"report-service-m/internal/reports/payslip_html"

	"github.com/gofiber/fiber/v2"
)

type reportHandlers struct{}

const maxPayslipHTMLReportCount = 1000

func newReportHandlers() reportHandlers {
	return reportHandlers{}
}

func (h reportHandlers) previewPayslipHTML(c *fiber.Ctx) error {
	templateType, data, err := h.loadPreviewPayslipHTMLData(c)
	if err != nil {
		return payslipHTMLErrorJSON(c, err)
	}

	html, _, err := renderPreviewPayslipHTML(c, templateType, data)
	if err != nil {
		return errorJSON(c, fiber.StatusInternalServerError, "render payslip html failed: "+err.Error())
	}

	c.Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	c.Set("Content-Type", "text/html; charset=utf-8")
	return c.SendString(html)
}

func (h reportHandlers) previewPayslipHTMLPDF(c *fiber.Ctx) error {
	templateType, data, err := h.loadPreviewPayslipHTMLData(c)
	if err != nil {
		return payslipHTMLErrorJSON(c, err)
	}

	html, count, err := renderPreviewPayslipHTML(c, templateType, data)
	if err != nil {
		return errorJSON(c, fiber.StatusInternalServerError, "render payslip html failed: "+err.Error())
	}

	pdfBytes, err := payslip_html.GeneratePDFWithOptions(html, payslip_html.PDFOptions{
		Timeout: payslipHTMLPDFTimeout(count),
	})
	if err != nil {
		return errorJSON(c, fiber.StatusInternalServerError, "generate payslip html pdf failed: "+err.Error())
	}

	disposition := "inline"
	if isTruthy(c.Query("download")) {
		disposition = "attachment"
	}

	filename := fmt.Sprintf("payslip_%s.pdf", templateType)
	if count > 1 {
		filename = fmt.Sprintf("payslip_%s_%04d.pdf", templateType, count)
	}

	return sendPDF(c, pdfBytes, filename, disposition)
}

func (h reportHandlers) loadPreviewPayslipHTMLData(c *fiber.Ctx) (string, payslip_html.Payslip, error) {
	templateType, err := payslip_html.NormalizeTemplateType(payslipTypeQuery(c))
	if err != nil {
		return "", payslip_html.Payslip{}, err
	}

	data, err := payslip_html.DefaultPayslip(templateType)
	if err != nil {
		return "", payslip_html.Payslip{}, errorJSON(c, fiber.StatusInternalServerError, "load payslip html config failed: "+err.Error())
	}

	return templateType, data, nil
}

func payslipHTMLErrorJSON(c *fiber.Ctx, err error) error {
	var unknownType payslip_html.UnknownTemplateTypeError
	if errors.As(err, &unknownType) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":           unknownType.Error(),
			"available_types": payslip_html.AvailableTemplates(),
		})
	}
	return err
}

func payslipTypeQuery(c *fiber.Ctx) string {
	if value := strings.TrimSpace(c.Query("type")); value != "" {
		return value
	}
	return payslip_html.DefaultType
}

func renderPreviewPayslipHTML(c *fiber.Ctx, templateType string, data payslip_html.Payslip) (string, int, error) {
	count := payslipHTMLReportCount(c)
	if count <= 1 {
		html, err := payslip_html.RenderHTML(templateType, data)
		return html, 1, err
	}

	html, err := payslip_html.RenderHTMLBatch(
		templateType,
		makePayslipTestBatch(data, count),
	)
	return html, count, err
}

func payslipHTMLReportCount(c *fiber.Ctx) int {
	count := intQuery(c, "count", 0)
	if count <= 0 {
		count = intQuery(c, "total", 1)
	}
	if count <= 0 {
		return 1
	}
	if count > maxPayslipHTMLReportCount {
		return maxPayslipHTMLReportCount
	}
	return count
}

func makePayslipTestBatch(base payslip_html.Payslip, count int) []payslip_html.Payslip {
	batch := make([]payslip_html.Payslip, 0, count)
	for i := 1; i <= count; i++ {
		item := base
		item.Report.Employee.Name = sequenceText(base.Report.Employee.Name, "Test Employee", i)
		item.Report.Employee.EmpID = sequenceCode(base.Report.Employee.EmpID, "EMP", i)
		item.Report.Payroll.SlipNo = sequenceCode(base.Report.Payroll.SlipNo, "SLIP", i)
		batch = append(batch, item)
	}
	return batch
}

func sequenceText(value, fallback string, seq int) string {
	value = strings.TrimSpace(value)
	if value == "" {
		value = fallback
	}
	return fmt.Sprintf("%s %04d", value, seq)
}

func sequenceCode(value, fallback string, seq int) string {
	value = strings.TrimSpace(value)
	if value == "" {
		value = fallback
	}
	return fmt.Sprintf("%s-%04d", value, seq)
}

func payslipHTMLPDFTimeout(count int) time.Duration {
	if count <= 1 {
		return 30 * time.Second
	}

	timeout := 30*time.Second + time.Duration(count)*500*time.Millisecond
	if timeout > 10*time.Minute {
		return 10 * time.Minute
	}
	return timeout
}

func intQuery(c *fiber.Ctx, key string, fallback int) int {
	value := strings.TrimSpace(c.Query(key))
	if value == "" {
		return fallback
	}

	number, err := strconv.Atoi(value)
	if err != nil || number < 0 {
		return fallback
	}

	return number
}

func isTruthy(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "y", "on":
		return true
	default:
		return false
	}
}
