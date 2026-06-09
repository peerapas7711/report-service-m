package payslip_html

import (
	"bytes"
	"embed"
	"encoding/base64"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"report-service-m/internal/reports/payslip"
)

//go:embed templates/*.html templates/assets/*
var templateFS embed.FS

const DefaultTemplate = "payslip_default"

type Payslip = payslip.Payslip
type LineItem = payslip.LineItem

type viewModel struct {
	Payslip
	TemplateName    string
	Stylesheet      template.CSS
	LogoPath        string
	EmployeeFields  []LineItem
	PayrollFields   []LineItem
	HasCompanyLogo  bool
	HasEarnings     bool
	HasDeductions   bool
	HasWorkStats    bool
	HasAccumulation bool
}

func RenderHTML(templateName string, data Payslip) (string, error) {
	name := normalizeTemplateName(templateName)

	tmpl, err := template.New(name).Funcs(template.FuncMap{
		"default": defaultText,
	}).ParseFS(
		templateFS,
		"templates/"+name+".html",
	)
	if err != nil {
		return "", fmt.Errorf("parse payslip html template %q: %w", name, err)
	}

	model := buildViewModel(name, data)

	var out bytes.Buffer
	if err := tmpl.ExecuteTemplate(&out, name+".html", model); err != nil {
		return "", fmt.Errorf("execute payslip html template %q: %w", name, err)
	}

	return out.String(), nil
}

func AvailableTemplates() []string {
	return []string{
		"payslip_default",
		"payslip_tigersoft",
	}
}

func normalizeTemplateName(templateName string) string {
	name := strings.TrimSpace(strings.ToLower(templateName))
	name = strings.TrimSuffix(name, filepath.Ext(name))
	name = strings.ReplaceAll(name, "-", "_")

	switch name {
	case "", "default", "modern", "payslip_default":
		return "payslip_default"
	case "tigersoft", "tiger_soft", "payslip_tigersoft":
		return "payslip_tigersoft"
	default:
		return name
	}
}

func buildViewModel(templateName string, data Payslip) viewModel {
	if strings.TrimSpace(data.Report.Company.Logo) == "" {
		data.Report.Company.Logo = fallbackLogoDataURL()
	}

	return viewModel{
		Payslip:         data,
		TemplateName:    templateName,
		Stylesheet:      template.CSS(stylesheet()),
		LogoPath:        data.Report.Company.Logo,
		EmployeeFields:  employeeFields(data.Report.Employee),
		PayrollFields:   payrollFields(data.Report.Payroll),
		HasCompanyLogo:  strings.TrimSpace(data.Report.Company.Logo) != "",
		HasEarnings:     len(data.Report.Earnings) > 0,
		HasDeductions:   len(data.Report.Deductions) > 0,
		HasWorkStats:    len(data.Report.WorkStats) > 0,
		HasAccumulation: len(data.Report.Accumulations) > 0,
	}
}

func employeeFields(employee payslip.Employee) []LineItem {
	return []LineItem{
		{Label: "Employee Name", Value: employee.Name},
		{Label: "Employee ID", Value: employee.EmpID},
		{Label: "Joined Date", Value: employee.JoinedDate},
		{Label: "Division", Value: employee.Division},
		{Label: "Department", Value: employee.Department},
		{Label: "Section", Value: employee.Section},
		{Label: "Position", Value: employee.Position},
		{Label: "Tax ID", Value: employee.TaxID},
	}
}

func payrollFields(payroll payslip.Payroll) []LineItem {
	return []LineItem{
		{Label: "Period", Value: payroll.Period},
		{Label: "Pay Date", Value: payroll.PayDate},
		{Label: "Slip No.", Value: payroll.SlipNo},
	}
}

func defaultText(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func stylesheet() string {
	css, err := templateFS.ReadFile("templates/assets/report.css")
	if err != nil {
		return ""
	}

	out := string(css)
	out = strings.ReplaceAll(out, "url('/report-assets/fonts/Sarabun-Regular.ttf')", "url('"+fontDataURL("assets/fonts/Sarabun/Sarabun-Regular.ttf")+"')")
	out = strings.ReplaceAll(out, "url('/report-assets/fonts/Sarabun-Bold.ttf')", "url('"+fontDataURL("assets/fonts/Sarabun/Sarabun-Bold.ttf")+"')")
	return out
}

func fontDataURL(path string) string {
	b, err := os.ReadFile(projectPath(path))
	if err != nil {
		return path
	}

	return "data:font/ttf;base64," + base64.StdEncoding.EncodeToString(b)
}

func fallbackLogoDataURL() string {
	b, err := templateFS.ReadFile("templates/assets/logo.png")
	if err != nil {
		return ""
	}

	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(b)
}

func projectPath(path string) string {
	if _, err := os.Stat(path); err == nil {
		return path
	}

	wd, err := os.Getwd()
	if err != nil {
		return path
	}

	for {
		candidate := filepath.Join(wd, path)
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}

		parent := filepath.Dir(wd)
		if parent == wd {
			return path
		}
		wd = parent
	}
}
