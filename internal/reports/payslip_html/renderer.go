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
const tigersoftLogoPath = "assets/icon_logoapp.png"

type Payslip = payslip.Payslip
type LineItem = payslip.LineItem

type viewModel struct {
	Payslip
	TemplateName    string
	Stylesheet      template.CSS
	LogoPath        template.URL
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

	tmpl, err := parsePayslipTemplate(name)
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

func RenderHTMLBatch(templateName string, data []Payslip) (string, error) {
	if len(data) == 0 {
		return "", fmt.Errorf("payslip batch is empty")
	}

	name := normalizeTemplateName(templateName)
	tmpl, err := parsePayslipTemplate(name)
	if err != nil {
		return "", fmt.Errorf("parse payslip html template %q: %w", name, err)
	}

	var out bytes.Buffer
	writeBatchHTMLStart(&out, defaultText(data[0].Report.Document.Title, "Payslip"))

	for _, item := range data {
		model := buildContentViewModel(name, item)
		if err := tmpl.ExecuteTemplate(&out, name+".content", model); err != nil {
			return "", fmt.Errorf("execute payslip html batch template %q: %w", name, err)
		}
		out.WriteByte('\n')
	}

	out.WriteString("</body>\n</html>\n")
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

func parsePayslipTemplate(name string) (*template.Template, error) {
	return template.New(name).Funcs(template.FuncMap{
		"default": defaultText,
	}).ParseFS(
		templateFS,
		"templates/"+name+".html",
	)
}

func writeBatchHTMLStart(out *bytes.Buffer, title string) {
	out.WriteString("<!doctype html>\n")
	out.WriteString("<html lang=\"th\">\n")
	out.WriteString("<head>\n")
	out.WriteString("  <meta charset=\"utf-8\">\n")
	out.WriteString("  <title>")
	out.WriteString(template.HTMLEscapeString(title))
	out.WriteString("</title>\n")
	out.WriteString("  <style>")
	out.WriteString(stylesheet())
	out.WriteString("</style>\n")
	out.WriteString("</head>\n")
	out.WriteString("<body class=\"batch\">\n")
}

func buildViewModel(templateName string, data Payslip) viewModel {
	return buildViewModelWithStylesheet(templateName, data, true)
}

func buildContentViewModel(templateName string, data Payslip) viewModel {
	return buildViewModelWithStylesheet(templateName, data, false)
}

func buildViewModelWithStylesheet(templateName string, data Payslip, includeStylesheet bool) viewModel {
	logoPath := strings.TrimSpace(data.Report.Company.Logo)
	if templateName == "payslip_tigersoft" {
		if appLogo := imageFileDataURL(tigersoftLogoPath); appLogo != "" {
			logoPath = appLogo
		}
	}
	if logoPath == "" {
		logoPath = fallbackLogoDataURL()
	} else {
		logoPath = resolveLogoPath(logoPath)
	}
	data.Report.Company.Logo = logoPath

	var css template.CSS
	if includeStylesheet {
		css = template.CSS(stylesheet())
	}

	return viewModel{
		Payslip:         data,
		TemplateName:    templateName,
		Stylesheet:      css,
		LogoPath:        template.URL(logoPath),
		EmployeeFields:  employeeFields(data.Report.Employee),
		PayrollFields:   payrollFields(data.Report.Payroll),
		HasCompanyLogo:  logoPath != "",
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

func resolveLogoPath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" ||
		strings.HasPrefix(path, "data:") ||
		strings.HasPrefix(path, "http://") ||
		strings.HasPrefix(path, "https://") {
		return path
	}

	if dataURL := imageFileDataURL(path); dataURL != "" {
		return dataURL
	}
	return path
}

func imageFileDataURL(path string) string {
	b, err := os.ReadFile(projectPath(path))
	if err != nil {
		return ""
	}

	return "data:" + imageMIMEType(path) + ";base64," + base64.StdEncoding.EncodeToString(b)
}

func imageMIMEType(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".webp":
		return "image/webp"
	case ".gif":
		return "image/gif"
	default:
		return "image/png"
	}
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
