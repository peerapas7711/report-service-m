package httpserver

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"report-service-m/internal/reports/payslip_html"

	"github.com/gofiber/fiber/v2"
)

const (
	reportFormatHTML = "html"
	reportFormatPDF  = "pdf"
)

type payslipReportRequest struct {
	TemplateCode string            `json:"templateCode"`
	Format       string            `json:"format"`
	CompanyCode  string            `json:"companyCode"`
	Data         payslipReportData `json:"data"`
}

type payslipReportData struct {
	Company   string                  `json:"company"`
	CompanyEn string                  `json:"companyEn"`
	Address   string                  `json:"address"`
	Tel       string                  `json:"tel"`
	LogoURL   string                  `json:"logoUrl"`
	Employees []payslipReportEmployee `json:"employees"`
}

type payslipReportEmployee struct {
	EmpCode                 string                    `json:"empCode"`
	EmpName                 string                    `json:"empName"`
	Position                string                    `json:"position"`
	Department              string                    `json:"department"`
	EmpType                 string                    `json:"empType"`
	StartDate               string                    `json:"startDate"`
	Salary                  float64                   `json:"salary"`
	BankName                string                    `json:"bankName"`
	BankAccount             string                    `json:"bankAccount"`
	Year                    string                    `json:"year"`
	Month                   string                    `json:"month"`
	PeriodNo                string                    `json:"periodNo"`
	PayDate                 string                    `json:"payDate"`
	Earnings                []payslipReportAmountItem `json:"earnings"`
	Deductions              []payslipReportAmountItem `json:"deductions"`
	TotalEarnings           float64                   `json:"totalEarnings"`
	TotalDeductions         float64                   `json:"totalDeductions"`
	NetPay                  float64                   `json:"netPay"`
	YTDIncome               float64                   `json:"ytdIncome"`
	YTDTax                  float64                   `json:"ytdTax"`
	YTDSocialSecurity       float64                   `json:"ytdSocialSecurity"`
	YTDProvidentFund        float64                   `json:"ytdProvidentFund"`
	YTDProvidentFundCompany float64                   `json:"ytdProvidentFundCompany"`
	TaxAllowances           payslipReportTaxAllowance `json:"taxAllowances"`
}

type payslipReportTaxAllowance struct {
	Expenses   []payslipReportAmountItem `json:"expenses"`
	Allowances []payslipReportAmountItem `json:"allowances"`
}

type payslipReportAmountItem struct {
	Name   string  `json:"name"`
	Amount float64 `json:"amount"`
}

func loadPayslipReportRequest(c *fiber.Ctx) (payslipReportRequest, error) {
	body := c.Body()
	if len(strings.TrimSpace(string(body))) == 0 {
		return payslipReportRequest{}, fiber.NewError(fiber.StatusBadRequest, "request body is required")
	}

	var req payslipReportRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return payslipReportRequest{}, fiber.NewError(fiber.StatusBadRequest, "invalid payslip report request json: "+err.Error())
	}
	if strings.TrimSpace(req.TemplateCode) == "" {
		return payslipReportRequest{}, fiber.NewError(fiber.StatusBadRequest, "templateCode is required")
	}
	if strings.TrimSpace(req.CompanyCode) == "" {
		return payslipReportRequest{}, fiber.NewError(fiber.StatusBadRequest, "companyCode is required")
	}
	if len(req.Data.Employees) == 0 {
		return payslipReportRequest{}, fiber.NewError(fiber.StatusBadRequest, "data.employees is required")
	}

	return req, nil
}

func payslipReportFormat(req payslipReportRequest, override string) (string, error) {
	value := strings.TrimSpace(override)
	if value == "" {
		value = strings.TrimSpace(req.Format)
	}
	value = strings.ToLower(value)

	switch value {
	case reportFormatHTML, reportFormatPDF:
		return value, nil
	case "":
		return "", fiber.NewError(fiber.StatusBadRequest, "format is required")
	default:
		return "", fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("unsupported report format %q", value))
	}
}

func renderPayslipReportHTML(req payslipReportRequest) (string, int, string, error) {
	templateType, cfg, payslips, err := buildPayslipReport(req)
	if err != nil {
		return "", 0, "", err
	}

	if len(payslips) == 1 {
		cfg.SampleData = payslips[0]
		html, err := payslip_html.RenderHTMLWithConfig(cfg)
		return html, 1, templateType, err
	}

	html, err := payslip_html.RenderHTMLBatchWithConfig(cfg, payslips)
	return html, len(payslips), templateType, err
}

func buildPayslipReport(req payslipReportRequest) (string, payslip_html.TemplateConfig, []payslip_html.Payslip, error) {
	templateType, err := payslip_html.NormalizeTemplateType(req.TemplateCode)
	if err != nil {
		return "", payslip_html.TemplateConfig{}, nil, err
	}

	cfg, err := payslip_html.DefaultTemplateConfig(templateType)
	if err != nil {
		return "", payslip_html.TemplateConfig{}, nil, err
	}

	base, err := payslip_html.DefaultPayslip(templateType)
	if err != nil {
		return "", payslip_html.TemplateConfig{}, nil, err
	}

	companyName := firstNonEmpty(req.Data.Company, req.Data.CompanyEn, cfg.Name)
	if companyName != "" {
		cfg.Name = companyName
	}

	payslips := make([]payslip_html.Payslip, 0, len(req.Data.Employees))
	for _, employee := range req.Data.Employees {
		payslips = append(payslips, buildEmployeePayslip(req, templateType, companyName, base, employee))
	}
	cfg.SampleData = payslips[0]

	return templateType, cfg, payslips, nil
}

func buildEmployeePayslip(req payslipReportRequest, templateType, companyName string, base payslip_html.Payslip, employee payslipReportEmployee) payslip_html.Payslip {
	item := base
	item.TemplateID = templateType
	item.Report.Company.Name = companyName
	item.Report.Company.NameEn = strings.TrimSpace(req.Data.CompanyEn)
	item.Report.Company.Address = strings.TrimSpace(req.Data.Address)
	item.Report.Company.Tel = strings.TrimSpace(req.Data.Tel)
	item.Report.Company.Logo = strings.TrimSpace(req.Data.LogoURL)
	item.Report.Document.ConfidentialTitle = companyName
	item.Report.Document.ConfidentialSubTitle = companyContactText(req.Data)
	item.Report.Payroll.Period = periodText(employee.Month, employee.Year)
	item.Report.Payroll.PayDate = strings.TrimSpace(employee.PayDate)
	item.Report.Payroll.SlipNo = payslipNo(req.CompanyCode, employee)
	item.Report.Employee.Name = strings.TrimSpace(employee.EmpName)
	item.Report.Employee.EmpID = strings.TrimSpace(employee.EmpCode)
	item.Report.Employee.JoinedDate = strings.TrimSpace(employee.StartDate)
	item.Report.Employee.Division = strings.TrimSpace(employee.Department)
	item.Report.Employee.Department = strings.TrimSpace(employee.Department)
	item.Report.Employee.Section = strings.TrimSpace(employee.EmpType)
	item.Report.Employee.Position = strings.TrimSpace(employee.Position)
	item.Report.Accumulations = employeeAccumulations(employee)
	item.Report.Earnings = amountLineItems(employee.Earnings)
	item.Report.Deductions = amountLineItems(employee.Deductions)
	item.Report.WorkStats = nil
	item.Report.Totals = payslip_html.Totals{
		TotalIncome:   moneyText(employee.TotalEarnings),
		TotalDeduct:   moneyText(employee.TotalDeductions),
		NetPay:        moneyText(employee.NetPay),
		BankAccountNo: strings.TrimSpace(employee.BankAccount),
	}
	item.Report.Texts = map[string]string{
		"payment_type": strings.TrimSpace(employee.EmpType),
		"bank_name":    strings.TrimSpace(employee.BankName),
		"period_no":    strings.TrimSpace(employee.PeriodNo),
		"company_code": strings.TrimSpace(req.CompanyCode),
	}
	item.Report.Sections = map[string][]payslip_html.LineItem{
		"tax_deductions_left":  amountLineItems(employee.TaxAllowances.Expenses),
		"tax_deductions_right": amountLineItems(employee.TaxAllowances.Allowances),
	}

	return item
}

func employeeAccumulations(employee payslipReportEmployee) []payslip_html.LineItem {
	return []payslip_html.LineItem{
		{Label: "รายได้สะสม", Value: moneyText(employee.YTDIncome)},
		{Label: "ภาษีสะสม", Value: moneyText(employee.YTDTax)},
		{Label: "กองทุนประกันสังคม", Value: moneyText(employee.YTDSocialSecurity)},
		{Label: "กองทุนสำรองเลี้ยงชีพสะสม", Value: moneyText(employee.YTDProvidentFund)},
		{Label: "กองทุนสำรองเลี้ยงชีพสมทบ", Value: moneyText(employee.YTDProvidentFundCompany)},
	}
}

func amountLineItems(items []payslipReportAmountItem) []payslip_html.LineItem {
	out := make([]payslip_html.LineItem, 0, len(items))
	for _, item := range items {
		label := strings.TrimSpace(item.Name)
		if label == "" {
			label = "รายการ"
		}
		out = append(out, payslip_html.LineItem{
			Label: label,
			Value: moneyText(item.Amount),
		})
	}
	return out
}

func companyContactText(data payslipReportData) string {
	parts := make([]string, 0, 2)
	if address := strings.TrimSpace(data.Address); address != "" {
		parts = append(parts, address)
	}
	if tel := strings.TrimSpace(data.Tel); tel != "" {
		parts = append(parts, "โทร. "+tel)
	}
	return strings.Join(parts, "\n")
}

func periodText(month, year string) string {
	month = strings.TrimSpace(month)
	year = strings.TrimSpace(year)
	if month == "" {
		return year
	}
	if year == "" {
		return month
	}
	return month + " " + year
}

func payslipNo(companyCode string, employee payslipReportEmployee) string {
	parts := []string{
		strings.TrimSpace(companyCode),
		strings.TrimSpace(employee.EmpCode),
		strings.TrimSpace(employee.PeriodNo),
	}

	filtered := parts[:0]
	for _, part := range parts {
		if part != "" {
			filtered = append(filtered, part)
		}
	}
	return strings.Join(filtered, "-")
}

func moneyText(value float64) string {
	sign := ""
	if value < 0 {
		sign = "-"
		value = -value
	}

	value = math.Round(value*100) / 100
	parts := strings.Split(fmt.Sprintf("%.2f", value), ".")
	whole := parts[0]
	decimal := parts[1]

	var grouped []byte
	for i, r := range whole {
		if i > 0 && (len(whole)-i)%3 == 0 {
			grouped = append(grouped, ',')
		}
		grouped = append(grouped, byte(r))
	}

	return sign + string(grouped) + "." + decimal
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
