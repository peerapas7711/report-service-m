package payslip_html

import (
	"fmt"
	"html/template"
	"math"
	"strconv"
	"strings"
)

func buildFlatPayrollTemplateData(cfg TemplateConfig, payslips []Payslip) map[string]any {
	first := applyTemplateDefaults(cfg, payslips[0])
	company := first.Report.Company
	address, tel := flatPayrollContactText(company, first.Report.Document.ConfidentialSubTitle)

	employees := make([]map[string]any, 0, len(payslips))
	for _, payslip := range payslips {
		employees = append(employees, flatPayrollEmployeeData(applyTemplateDefaults(cfg, payslip)))
	}

	return map[string]any{
		"company":   firstNonEmptyText(company.Name, cfg.Name, company.NameEn),
		"companyEn": firstNonEmptyText(company.NameEn, company.Name, cfg.Name),
		"address":   address,
		"tel":       tel,
		"logoUrl":   assetDataURL(company.Logo),
		"employees": employees,
		"t":         flatPayrollTranslations(first.Report.Document.Title),
	}
}

func flatPayrollEmployeeData(payslip Payslip) map[string]any {
	report := payslip.Report
	texts := report.Texts
	accumulations := report.Accumulations
	taxExpenses := flatPayrollLineItems(report.Sections["tax_deductions_left"])
	taxAllowances := flatPayrollLineItems(report.Sections["tax_deductions_right"])

	var taxAllowanceData any
	if len(taxExpenses) > 0 || len(taxAllowances) > 0 {
		taxAllowanceData = map[string]any{
			"expenses":   taxExpenses,
			"allowances": taxAllowances,
		}
	}

	return map[string]any{
		"empCode":                 report.Employee.EmpID,
		"empName":                 report.Employee.Name,
		"position":                report.Employee.Position,
		"department":              firstNonEmptyText(report.Employee.Department, report.Employee.Division),
		"empType":                 texts["payment_type"],
		"startDate":               report.Employee.JoinedDate,
		"bankName":                texts["bank_name"],
		"bankAccount":             report.Totals.BankAccountNo,
		"month":                   report.Payroll.Period,
		"periodNo":                firstNonEmptyText(texts["period_no"], report.Payroll.SlipNo),
		"payDate":                 report.Payroll.PayDate,
		"earnings":                flatPayrollLineItems(report.Earnings),
		"deductions":              flatPayrollLineItems(report.Deductions),
		"totalEarnings":           report.Totals.TotalIncome,
		"totalDeductions":         report.Totals.TotalDeduct,
		"netPay":                  report.Totals.NetPay,
		"ytdIncome":               lineItemValue(accumulations, 0),
		"ytdTax":                  lineItemValue(accumulations, 1),
		"ytdSocialSecurity":       lineItemValue(accumulations, 2),
		"ytdProvidentFund":        lineItemValue(accumulations, 3),
		"ytdProvidentFundCompany": lineItemValue(accumulations, 4),
		"taxAllowances":           taxAllowanceData,
	}
}

func flatPayrollLineItems(items []LineItem) []map[string]any {
	out := make([]map[string]any, 0, len(items))
	for _, item := range items {
		out = append(out, map[string]any{
			"name":   item.Label,
			"amount": item.Value,
		})
	}
	return out
}

func flatPayrollContactText(company Company, fallback string) (string, string) {
	address := strings.TrimSpace(company.Address)
	tel := strings.TrimSpace(company.Tel)
	if address != "" && tel != "" {
		return address, tel
	}

	var addressLines []string
	for _, line := range strings.Split(fallback, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if tel == "" && (strings.Contains(line, "โทร") || strings.Contains(strings.ToLower(line), "tel")) {
			tel = line
			continue
		}
		addressLines = append(addressLines, line)
	}
	if address == "" {
		address = strings.Join(addressLines, "\n")
	}

	return address, tel
}

func flatPayrollTranslations(title string) map[string]string {
	return map[string]string{
		"title":                         defaultText(title, "ใบแจ้งรายได้ (Pay Slip)"),
		"emp_code":                      "รหัส",
		"period":                        "สำหรับงวดจ่าย",
		"month_prefix":                  "",
		"period_no":                     "งวดที่",
		"emp_name":                      "ชื่อพนักงาน",
		"pay_date":                      "วันที่จ่ายเงิน",
		"position":                      "ตำแหน่ง",
		"emp_type":                      "ประเภทพนักงาน",
		"department":                    "ฝ่าย",
		"bank_transfer":                 "เงินนำส่ง",
		"default_bank":                  "-",
		"start_date":                    "วันที่เข้างาน",
		"bank_account":                  "เลขที่บัญชี",
		"earnings_header":               "รายการได้ / Income",
		"deductions_header":             "รายการหัก / Deduct",
		"total_earnings":                "รวมรายได้ทั้งหมด / Total Income",
		"total_deductions":              "รวมรายการหักทั้งหมด",
		"net_pay":                       "เงินได้สุทธิ",
		"ytd_header":                    "ยอดสะสม",
		"ytd_income":                    "รายได้สะสม",
		"ytd_tax":                       "ภาษีสะสม",
		"ytd_social_security":           "กองทุนประกันสังคม",
		"ytd_provident_fund":            "กองทุนสำรองเลี้ยงชีพสะสม",
		"ytd_provident_fund_company":    "กองทุนสำรองเลี้ยงชีพสมทบ",
		"total_provident_fund_employee": "เงินกองทุนสำรองเลี้ยงชีพสะสมทั้งสิ้น",
		"total_provident_fund_company":  "เงินกองทุนสำรองเลี้ยงชีพสมทบทั้งสิ้น",
		"tax_allowances_header":         "รายการหักใช้จ่าย / ลดหย่อนภาษี",
	}
}

func lineItemValue(items []LineItem, index int) string {
	if index < 0 || index >= len(items) {
		return ""
	}
	return items[index].Value
}

func moneyTemplateText(value any) string {
	switch v := value.(type) {
	case string:
		return strings.TrimSpace(v)
	case template.HTML:
		return string(v)
	case fmt.Stringer:
		return v.String()
	case float64:
		return formatMoneyTemplateText(v)
	case float32:
		return formatMoneyTemplateText(float64(v))
	case int:
		return formatMoneyTemplateText(float64(v))
	case int64:
		return formatMoneyTemplateText(float64(v))
	case int32:
		return formatMoneyTemplateText(float64(v))
	case uint:
		return formatMoneyTemplateText(float64(v))
	case uint64:
		return formatMoneyTemplateText(float64(v))
	case uint32:
		return formatMoneyTemplateText(float64(v))
	case nil:
		return ""
	default:
		return strings.TrimSpace(fmt.Sprint(v))
	}
}

func formatMoneyTemplateText(value float64) string {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return ""
	}

	sign := ""
	if value < 0 {
		sign = "-"
		value = -value
	}

	value = math.Round(value*100) / 100
	parts := strings.Split(strconv.FormatFloat(value, 'f', 2, 64), ".")
	whole := parts[0]
	decimal := parts[1]

	grouped := make([]byte, 0, len(whole)+len(whole)/3)
	for i := 0; i < len(whole); i++ {
		if i > 0 && (len(whole)-i)%3 == 0 {
			grouped = append(grouped, ',')
		}
		grouped = append(grouped, whole[i])
	}

	return sign + string(grouped) + "." + decimal
}

func firstNonEmptyText(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
