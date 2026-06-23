package httpserver

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"report-service-m/internal/datasources/payslipdata"
	"report-service-m/internal/reports/payslip"
	"report-service-m/internal/reports/payslip_html"
	"report-service-m/internal/reports/systemaccesspermission"

	"github.com/gofiber/fiber/v2"
)

type payslipRenderRequest struct {
	Data        payslip.Payslip `json:"data"`
	Orientation string          `json:"orientation,omitempty"`
	TemplateID  string          `json:"template_id,omitempty"`
	FileName    string          `json:"file_name,omitempty"`
}

type systemAccessRenderRequest struct {
	Data        systemaccesspermission.SystemAccess `json:"data"`
	Orientation string                              `json:"orientation,omitempty"`
	Lang        string                              `json:"lang,omitempty"`
	FileName    string                              `json:"file_name,omitempty"`
}

type queueTicketRenderRequest struct {
	Title      string `json:"title,omitempty"`
	Subtitle   string `json:"subtitle,omitempty"`
	QueueLabel string `json:"queue_label,omitempty"`
	Label      string `json:"label,omitempty"`
	Prefix     string `json:"prefix,omitempty"`
	Lang       string `json:"lang,omitempty"`
	Start      int    `json:"start,omitempty"`
	Total      int    `json:"total,omitempty"`
	Digits     int    `json:"digits,omitempty"`
	FileName   string `json:"file_name,omitempty"`
}

type reportHandlers struct {
	payslipRepo    payslipdata.Repository
	payslipRepoErr error
}

func newReportHandlers() reportHandlers {
	repo, err := newPreviewPayslipRepository()
	return reportHandlers{
		payslipRepo:    repo,
		payslipRepoErr: err,
	}
}

func renderPayslip(c *fiber.Ctx) error {
	var req payslipRenderRequest
	if err := c.BodyParser(&req); err != nil {
		return errorJSON(c, fiber.StatusBadRequest, "invalid payslip request: "+err.Error())
	}

	pdfBytes, err := payslip.RenderWithOptions(req.Data, payslip.RenderOptions{
		Orientation: req.Orientation,
		TemplateID:  req.TemplateID,
	})
	if err != nil {
		return errorJSON(c, fiber.StatusInternalServerError, "render payslip failed: "+err.Error())
	}

	return sendPDF(c, pdfBytes, defaultString(req.FileName, "payslip.pdf"), "attachment")
}

func renderSystemAccessPermission(c *fiber.Ctx) error {
	var req systemAccessRenderRequest
	if err := c.BodyParser(&req); err != nil {
		return errorJSON(c, fiber.StatusBadRequest, "invalid system access request: "+err.Error())
	}

	pdfBytes, err := systemaccesspermission.SystemAccessPermission(
		req.Data,
		defaultString(req.Orientation, "P"),
		defaultString(req.Lang, "th"),
	)
	if err != nil {
		return errorJSON(c, fiber.StatusInternalServerError, "render system access permission failed: "+err.Error())
	}

	return sendPDF(c, pdfBytes, defaultString(req.FileName, "system_access_permission.pdf"), "attachment")
}

func (h reportHandlers) previewPayslip(c *fiber.Ctx) error {
	data, err := h.loadPreviewPayslipData(c)
	if err != nil {
		return err
	}

	pdfBytes, err := payslip.Render(data, c.Query("orientation", "P"))
	if err != nil {
		return errorJSON(c, fiber.StatusInternalServerError, "render payslip preview failed: "+err.Error())
	}

	disposition := "inline"
	if isTruthy(c.Query("download")) {
		disposition = "attachment"
	}

	return sendPDF(c, pdfBytes, "payslip_preview.pdf", disposition)
}

func (h reportHandlers) previewPayslipHTML(c *fiber.Ctx) error {
	data, err := h.loadPreviewPayslipData(c)
	if err != nil {
		return err
	}

	html, err := payslip_html.RenderHTML(c.Query("template", data.TemplateID), data)
	if err != nil {
		return errorJSON(c, fiber.StatusInternalServerError, "render payslip html failed: "+err.Error())
	}

	c.Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	c.Set("Content-Type", "text/html; charset=utf-8")
	return c.SendString(html)
}

func (h reportHandlers) previewPayslipHTMLPDF(c *fiber.Ctx) error {
	data, err := h.loadPreviewPayslipData(c)
	if err != nil {
		return err
	}

	html, err := payslip_html.RenderHTML(c.Query("template", data.TemplateID), data)
	if err != nil {
		return errorJSON(c, fiber.StatusInternalServerError, "render payslip html failed: "+err.Error())
	}

	pdfBytes, err := payslip_html.GeneratePDF(html)
	if err != nil {
		return errorJSON(c, fiber.StatusInternalServerError, "generate payslip html pdf failed: "+err.Error())
	}

	disposition := "inline"
	if isTruthy(c.Query("download")) {
		disposition = "attachment"
	}

	return sendPDF(c, pdfBytes, "payslip_html_preview.pdf", disposition)
}

func (h reportHandlers) loadPreviewPayslipData(c *fiber.Ctx) (payslip.Payslip, error) {
	if h.payslipRepoErr != nil {
		return payslip.Payslip{}, errorJSON(c, fiber.StatusInternalServerError, "initialize payslip repository failed: "+h.payslipRepoErr.Error())
	}

	query := payslipQuery(c)
	data, err := h.payslipRepo.FindPayslip(c.UserContext(), query)
	if err != nil {
		if errors.Is(err, payslipdata.ErrNotFound) {
			return payslip.Payslip{}, c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":           "payslip mock not found",
				"available_mocks": availablePayslipMocks(h.payslipRepo),
			})
		}
		return payslip.Payslip{}, errorJSON(c, fiber.StatusInternalServerError, "load payslip data failed: "+err.Error())
	}

	if companyName := strings.TrimSpace(c.Query("company_name")); companyName != "" {
		data.Report.Company.Name = companyName
	}

	logoURL := strings.TrimSpace(c.Query("logo"))
	if logoURL == "" {
		logoURL = strings.TrimSpace(c.Query("logo_url"))
	}
	if logoURL != "" {
		data.Report.Company.Logo = logoURL
	}

	if templateID := strings.TrimSpace(c.Query("template")); templateID != "" {
		data.TemplateID = templateID
	}

	return data, nil
}

func payslipQuery(c *fiber.Ctx) payslipdata.Query {
	hasSQLLikeFilters := strings.TrimSpace(c.Query("slip_id")) != "" ||
		strings.TrimSpace(c.Query("company_id")) != "" ||
		strings.TrimSpace(c.Query("employee_id")) != "" ||
		strings.TrimSpace(c.Query("period_id")) != "" ||
		strings.TrimSpace(c.Query("period")) != "" ||
		strings.TrimSpace(c.Query("slip_no")) != ""

	mockName := strings.TrimSpace(c.Query("mock"))
	if mockName == "" && !hasSQLLikeFilters {
		mockName = "hopinn"
	}

	return payslipdata.Query{
		MockName:   mockName,
		SlipID:     c.Query("slip_id"),
		CompanyID:  c.Query("company_id"),
		EmployeeID: c.Query("employee_id"),
		PeriodID:   c.Query("period_id"),
		Period:     c.Query("period"),
		SlipNo:     c.Query("slip_no"),
	}
}

func availablePayslipMocks(repo payslipdata.Repository) []string {
	if catalog, ok := repo.(payslipdata.MockCatalog); ok {
		return catalog.AvailableMocks()
	}
	return []string{}
}

func newPreviewPayslipRepository() (payslipdata.Repository, error) {
	seeds := []struct {
		name    string
		path    string
		aliases []string
	}{
		{name: "hopinn", path: "mock/payslip_hopinn.json", aliases: []string{"1", "default", "modern"}},
		{name: "tigersoft", path: "mock/payslip_tigersoft.json", aliases: []string{"2", "tiger_soft"}},
		{name: "bluewave", path: "mock/payslip_bluewave.json", aliases: []string{"3"}},
		{name: "kubota", path: "mock/payslip_kubota.json", aliases: []string{"4"}},
	}

	mockSeeds := make([]payslipdata.MockPayslipSeed, 0, len(seeds))
	for _, seed := range seeds {
		data, err := payslip.LoadFromFile(projectFile(seed.path))
		if err != nil {
			return nil, err
		}
		mockSeeds = append(mockSeeds, payslipdata.MockPayslipSeed{
			MockName: seed.name,
			Aliases:  seed.aliases,
			Data:     data,
		})
	}

	return payslipdata.NewMockSQLRepository(payslipdata.RowsFromPayslips(mockSeeds)), nil
}

func previewSystemAccessPermission(c *fiber.Ctx) error {
	data, err := systemaccesspermission.LoadFromFile(projectFile("mock/permissionreport.json"))
	if err != nil {
		return errorJSON(c, fiber.StatusInternalServerError, "load system access mock failed: "+err.Error())
	}

	pdfBytes, err := systemaccesspermission.SystemAccessPermission(
		data,
		c.Query("orientation", "P"),
		c.Query("lang", "th"),
	)
	if err != nil {
		return errorJSON(c, fiber.StatusInternalServerError, "render system access preview failed: "+err.Error())
	}

	return sendPDF(c, pdfBytes, "system_access_permission_preview.pdf", "inline")
}

func queueTicketFilename(start, total int) string {
	if start <= 0 {
		start = 1
	}
	if total <= 0 {
		total = 700
	}

	return fmt.Sprintf("queue_tickets_%04d_%04d.pdf", start, start+total-1)
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

func defaultString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
