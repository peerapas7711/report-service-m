package payslip_html

import (
	"bytes"
	"embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
)

//go:embed templates/payslip/*/template.html templates/payslip/*/style.css templates/payslip/*/config.json
var templateFS embed.FS

const DefaultType = "default"

type TemplateConfig struct {
	Type       string  `json:"type"`
	Name       string  `json:"name"`
	BodyClass  string  `json:"body_class"`
	SampleData Payslip `json:"sample_data"`
}

type UnknownTemplateTypeError struct {
	Value string
}

func (e UnknownTemplateTypeError) Error() string {
	return fmt.Sprintf("unknown payslip type %q", e.Value)
}

type viewModel struct {
	Payslip
	Config          TemplateConfig
	TemplateType    string
	Stylesheet      template.CSS
	EmployeeFields  []LineItem
	PayrollFields   []LineItem
	HasEarnings     bool
	HasDeductions   bool
	HasWorkStats    bool
	HasAccumulation bool
}

func RenderHTML(templateType string, data Payslip) (string, error) {
	selectedType, err := NormalizeTemplateType(templateType)
	if err != nil {
		return "", err
	}

	tmpl, err := parsePayslipTemplate(selectedType)
	if err != nil {
		return "", fmt.Errorf("parse payslip html template %q: %w", selectedType, err)
	}

	model, err := buildViewModel(selectedType, data)
	if err != nil {
		return "", err
	}

	var out bytes.Buffer
	if err := tmpl.ExecuteTemplate(&out, "template.html", model); err != nil {
		return "", fmt.Errorf("execute payslip html template %q: %w", selectedType, err)
	}

	return out.String(), nil
}

func RenderHTMLBatch(templateType string, data []Payslip) (string, error) {
	if len(data) == 0 {
		return "", fmt.Errorf("payslip batch is empty")
	}

	selectedType, err := NormalizeTemplateType(templateType)
	if err != nil {
		return "", err
	}

	tmpl, err := parsePayslipTemplate(selectedType)
	if err != nil {
		return "", fmt.Errorf("parse payslip html template %q: %w", selectedType, err)
	}

	firstModel, err := buildViewModel(selectedType, data[0])
	if err != nil {
		return "", err
	}

	var out bytes.Buffer
	writeBatchHTMLStart(&out, firstModel)

	for _, item := range data {
		model, err := buildContentViewModel(selectedType, item)
		if err != nil {
			return "", err
		}
		if err := tmpl.ExecuteTemplate(&out, "content", model); err != nil {
			return "", fmt.Errorf("execute payslip html batch template %q: %w", selectedType, err)
		}
		out.WriteByte('\n')
	}

	out.WriteString("</body>\n</html>\n")
	return out.String(), nil
}

func DefaultPayslip(templateType string) (Payslip, error) {
	selectedType, err := NormalizeTemplateType(templateType)
	if err != nil {
		return Payslip{}, err
	}

	cfg, err := loadTemplateConfig(selectedType)
	if err != nil {
		return Payslip{}, err
	}

	data := cfg.SampleData
	if strings.TrimSpace(data.TemplateID) == "" {
		data.TemplateID = selectedType
	}
	if strings.TrimSpace(data.Report.Company.Name) == "" {
		data.Report.Company.Name = cfg.Name
	}
	return data, nil
}

func AvailableTemplates() []string {
	return []string{
		"default",
		"thai_demar",
		
	}
}

func NormalizeTemplateType(templateType string) (string, error) {
	name := strings.TrimSpace(strings.ToLower(templateType))
	name = strings.TrimSuffix(name, filepath.Ext(name))
	name = strings.ReplaceAll(name, "-", "_")

	switch name {
	case "", "default":
		return "default", nil
	case "thai_demar":
		return "thai_demar", nil
	
	default:
		return "", UnknownTemplateTypeError{Value: templateType}
	}
}

func parsePayslipTemplate(templateType string) (*template.Template, error) {
	return template.New("payslip-"+templateType).Funcs(template.FuncMap{
		"default": defaultText,
	}).ParseFS(
		templateFS,
		"templates/payslip/"+templateType+"/template.html",
	)
}

func loadTemplateConfig(templateType string) (TemplateConfig, error) {
	b, err := templateFS.ReadFile("templates/payslip/" + templateType + "/config.json")
	if err != nil {
		return TemplateConfig{}, err
	}

	var cfg TemplateConfig
	if err := json.Unmarshal(b, &cfg); err != nil {
		return TemplateConfig{}, err
	}
	if strings.TrimSpace(cfg.Type) == "" {
		cfg.Type = templateType
	}
	if strings.TrimSpace(cfg.BodyClass) == "" {
		cfg.BodyClass = "payslip-" + templateType
	}
	return cfg, nil
}

func writeBatchHTMLStart(out *bytes.Buffer, model viewModel) {
	out.WriteString("<!doctype html>\n")
	out.WriteString("<html lang=\"th\">\n")
	out.WriteString("<head>\n")
	out.WriteString("  <meta charset=\"utf-8\">\n")
	out.WriteString("  <title>")
	out.WriteString(template.HTMLEscapeString(defaultText(model.Report.Document.Title, "Payslip")))
	out.WriteString("</title>\n")
	out.WriteString("  <style>")
	out.WriteString(string(model.Stylesheet))
	out.WriteString("</style>\n")
	out.WriteString("</head>\n")
	out.WriteString("<body class=\"batch ")
	out.WriteString(template.HTMLEscapeString(model.Config.BodyClass))
	out.WriteString("\">\n")
}

func buildViewModel(templateType string, data Payslip) (viewModel, error) {
	return buildViewModelWithStylesheet(templateType, data, true)
}

func buildContentViewModel(templateType string, data Payslip) (viewModel, error) {
	return buildViewModelWithStylesheet(templateType, data, false)
}

func buildViewModelWithStylesheet(templateType string, data Payslip, includeStylesheet bool) (viewModel, error) {
	cfg, err := loadTemplateConfig(templateType)
	if err != nil {
		return viewModel{}, err
	}

	var css template.CSS
	if includeStylesheet {
		rawCSS, err := stylesheet(templateType)
		if err != nil {
			return viewModel{}, err
		}
		css = template.CSS(rawCSS)
	}

	return viewModel{
		Payslip:         data,
		Config:          cfg,
		TemplateType:    templateType,
		Stylesheet:      css,
		EmployeeFields:  employeeFields(data.Report.Employee),
		PayrollFields:   payrollFields(data.Report.Payroll),
		HasEarnings:     len(data.Report.Earnings) > 0,
		HasDeductions:   len(data.Report.Deductions) > 0,
		HasWorkStats:    len(data.Report.WorkStats) > 0,
		HasAccumulation: len(data.Report.Accumulations) > 0,
	}, nil
}

func employeeFields(employee Employee) []LineItem {
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

func payrollFields(payroll Payroll) []LineItem {
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

func stylesheet(templateType string) (string, error) {
	css, err := templateFS.ReadFile("templates/payslip/" + templateType + "/style.css")
	if err != nil {
		return "", err
	}

	out := string(css)
	out = strings.ReplaceAll(out, "url('/report-assets/fonts/Sarabun-Regular.ttf')", "url('"+fontDataURL("assets/fonts/Sarabun/Sarabun-Regular.ttf")+"')")
	out = strings.ReplaceAll(out, "url('/report-assets/fonts/Sarabun-Bold.ttf')", "url('"+fontDataURL("assets/fonts/Sarabun/Sarabun-Bold.ttf")+"')")
	return out, nil
}

func fontDataURL(path string) string {
	b, err := os.ReadFile(projectPath(path))
	if err != nil {
		return path
	}

	return "data:font/ttf;base64," + base64.StdEncoding.EncodeToString(b)
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
