package payslip

const (
	templateModern  = "modern"
	templateKubota  = "kubota"
	defaultTemplate = templateModern
)

type RenderOptions struct {
	Orientation string
	TemplateID  string
}

type renderModel struct {
	TemplateID string
	Report     Report
	Values     map[string]string
	Sections   map[string][]LineItem
}

func buildRenderModel(data Payslip, templateID string) renderModel {
	report := data.Report

	return renderModel{
		TemplateID: normalizeTemplateID(templateID),
		Report:     report,
		Values:     buildValueMap(report),
		Sections:   buildSectionMap(report),
	}
}

func buildEmployeeFields(employee Employee) []LineItem {
	return []LineItem{
		{Label: "Name", Value: employee.Name},
		{Label: "Emp ID.", Value: employee.EmpID},
		{Label: "Joined Date", Value: employee.JoinedDate},
		{Label: "Division", Value: employee.Division},
		{Label: "Department", Value: employee.Department},
		{Label: "Section", Value: employee.Section},
		{Label: "Position", Value: employee.Position},
		{Label: "Tax ID", Value: employee.TaxID},
	}
}

func buildPayrollFields(payroll Payroll) []LineItem {
	return []LineItem{
		{Label: "Period", Value: payroll.Period},
		{Label: "Pay Date", Value: payroll.PayDate},
		{Label: "Slip No.", Value: payroll.SlipNo},
	}
}

func buildValueMap(report Report) map[string]string {
	summary := buildCompensationSummary(report)
	workSummary := buildWorkSummary(report)

	values := map[string]string{
		"company_name":                   report.Company.Name,
		"company_logo":                   report.Company.Logo,
		"document_title":                 report.Document.Title,
		"document_confidential_title":    report.Document.ConfidentialTitle,
		"document_confidential_subtitle": report.Document.ConfidentialSubTitle,
		"payroll_period":                 report.Payroll.Period,
		"payroll_pay_date":               report.Payroll.PayDate,
		"payroll_slip_no":                report.Payroll.SlipNo,
		"employee_name":                  report.Employee.Name,
		"employee_emp_id":                report.Employee.EmpID,
		"employee_joined_date":           report.Employee.JoinedDate,
		"employee_division":              report.Employee.Division,
		"employee_department":            report.Employee.Department,
		"employee_section":               report.Employee.Section,
		"employee_position":              report.Employee.Position,
		"employee_tax_id":                report.Employee.TaxID,
		"summary_total_income":           summary.TotalIncome,
		"summary_total_deduct":           summary.TotalDeduct,
		"summary_net_pay":                summary.NetPay,
		"summary_bank_account_no":        summary.BankAccountNo,
		"summary_overtime_total":         workSummary.OvertimeTotal,
		"summary_absent_value":           workSummary.AbsentValue,
	}

	for key, value := range report.Texts {
		values[key] = value
	}

	return values
}

func buildSectionMap(report Report) map[string][]LineItem {
	summary := buildCompensationSummary(report)
	workSummary := buildWorkSummary(report)

	sections := map[string][]LineItem{
		"employee_core":  cloneLineItems(buildEmployeeFields(report.Employee)),
		"payroll_header": cloneLineItems(buildPayrollFields(report.Payroll)),
		"accumulations":  cloneLineItems(report.Accumulations),
		"earnings":       cloneLineItems(report.Earnings),
		"deductions":     cloneLineItems(report.Deductions),
		"work_stats":     cloneLineItems(report.WorkStats),
		"summary_income": {
			{Label: "Total Income", Value: summary.TotalIncome},
			{Label: "Bank Account No.", Value: summary.BankAccountNo},
		},
		"summary_deductions": {
			{Label: "Total Deduct", Value: summary.TotalDeduct},
			{Label: "Net Pay", Value: summary.NetPay},
		},
		"summary_work": {
			{Label: "Overtime Total", Value: workSummary.OvertimeTotal},
			{Label: "Absent", Value: workSummary.AbsentValue},
		},
		"summary_work_values": {
			{Value: workSummary.OvertimeTotal},
			{Value: workSummary.AbsentValue},
		},
	}

	for key, items := range report.Sections {
		sections[key] = cloneLineItems(items)
	}

	return sections
}

func cloneLineItems(items []LineItem) []LineItem {
	if len(items) == 0 {
		return nil
	}

	cloned := make([]LineItem, len(items))
	copy(cloned, items)
	return cloned
}
