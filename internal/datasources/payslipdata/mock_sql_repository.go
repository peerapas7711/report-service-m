package payslipdata

import (
	"context"
	"sort"
	"strings"

	reportpayslip "report-service-m/internal/reports/payslip"
)

type MockPayslipSeed struct {
	MockName string
	Aliases  []string
	Data     reportpayslip.Payslip
}

type MockSQLRows struct {
	Aliases        []MockAliasRow
	Companies      []CompanyRow
	Employees      []EmployeeRow
	PayrollPeriods []PayrollPeriodRow
	Payslips       []PayslipRow
	LineItems      []LineItemRow
	Texts          []TextRow
}

type MockAliasRow struct {
	MockName string
	SlipID   string
}

type CompanyRow struct {
	CompanyID string
	Name      string
	LogoURL   string
}

type EmployeeRow struct {
	EmployeeID string
	CompanyID  string
	EmpNo      string
	Name       string
	JoinedDate string
	Division   string
	Department string
	Section    string
	Position   string
	TaxID      string
}

type PayrollPeriodRow struct {
	PeriodID  string
	CompanyID string
	Label     string
	PayDate   string
}

type PayslipRow struct {
	SlipID               string
	CompanyID            string
	EmployeeID           string
	PeriodID             string
	TemplateID           string
	SlipNo               string
	DocumentTitle        string
	ConfidentialTitle    string
	ConfidentialSubtitle string
	TotalIncome          string
	TotalDeduct          string
	NetPay               string
	BankAccountNo        string
}

type LineItemRow struct {
	SlipID    string
	Section   string
	Label     string
	Value     string
	SortOrder int
}

type TextRow struct {
	SlipID string
	Key    string
	Value  string
}

type MockSQLRepository struct {
	rows MockSQLRows
}

func NewMockSQLRepository(rows MockSQLRows) *MockSQLRepository {
	return &MockSQLRepository{rows: rows}
}

func RowsFromPayslips(seeds []MockPayslipSeed) MockSQLRows {
	rows := MockSQLRows{}

	for _, seed := range seeds {
		mockName := normalizeKey(seed.MockName)
		if mockName == "" {
			continue
		}

		data := seed.Data
		slipID := "slip_" + mockName
		companyID := "company_" + mockName
		employeeID := defaultString(data.Report.Employee.EmpID, "employee_"+mockName)
		periodID := "period_" + mockName

		rows.Aliases = append(rows.Aliases, MockAliasRow{MockName: mockName, SlipID: slipID})
		for _, alias := range seed.Aliases {
			alias = normalizeKey(alias)
			if alias != "" {
				rows.Aliases = append(rows.Aliases, MockAliasRow{MockName: alias, SlipID: slipID})
			}
		}

		rows.Companies = append(rows.Companies, CompanyRow{
			CompanyID: companyID,
			Name:      data.Report.Company.Name,
			LogoURL:   data.Report.Company.Logo,
		})
		rows.Employees = append(rows.Employees, EmployeeRow{
			EmployeeID: employeeID,
			CompanyID:  companyID,
			EmpNo:      data.Report.Employee.EmpID,
			Name:       data.Report.Employee.Name,
			JoinedDate: data.Report.Employee.JoinedDate,
			Division:   data.Report.Employee.Division,
			Department: data.Report.Employee.Department,
			Section:    data.Report.Employee.Section,
			Position:   data.Report.Employee.Position,
			TaxID:      data.Report.Employee.TaxID,
		})
		rows.PayrollPeriods = append(rows.PayrollPeriods, PayrollPeriodRow{
			PeriodID:  periodID,
			CompanyID: companyID,
			Label:     data.Report.Payroll.Period,
			PayDate:   data.Report.Payroll.PayDate,
		})
		rows.Payslips = append(rows.Payslips, PayslipRow{
			SlipID:               slipID,
			CompanyID:            companyID,
			EmployeeID:           employeeID,
			PeriodID:             periodID,
			TemplateID:           data.TemplateID,
			SlipNo:               data.Report.Payroll.SlipNo,
			DocumentTitle:        data.Report.Document.Title,
			ConfidentialTitle:    data.Report.Document.ConfidentialTitle,
			ConfidentialSubtitle: data.Report.Document.ConfidentialSubTitle,
			TotalIncome:          data.Report.Totals.TotalIncome,
			TotalDeduct:          data.Report.Totals.TotalDeduct,
			NetPay:               data.Report.Totals.NetPay,
			BankAccountNo:        data.Report.Totals.BankAccountNo,
		})

		rows.LineItems = appendLineItems(rows.LineItems, slipID, "accumulations", data.Report.Accumulations)
		rows.LineItems = appendLineItems(rows.LineItems, slipID, "earnings", data.Report.Earnings)
		rows.LineItems = appendLineItems(rows.LineItems, slipID, "deductions", data.Report.Deductions)
		rows.LineItems = appendLineItems(rows.LineItems, slipID, "work_stats", data.Report.WorkStats)

		sectionNames := sortedMapKeys(data.Report.Sections)
		for _, section := range sectionNames {
			rows.LineItems = appendLineItems(rows.LineItems, slipID, section, data.Report.Sections[section])
		}

		textKeys := sortedMapKeys(data.Report.Texts)
		for _, key := range textKeys {
			rows.Texts = append(rows.Texts, TextRow{
				SlipID: slipID,
				Key:    key,
				Value:  data.Report.Texts[key],
			})
		}
	}

	return rows
}

func (r *MockSQLRepository) FindPayslip(ctx context.Context, query Query) (reportpayslip.Payslip, error) {
	select {
	case <-ctx.Done():
		return reportpayslip.Payslip{}, ctx.Err()
	default:
	}

	row, ok := r.findPayslipRow(query)
	if !ok {
		return reportpayslip.Payslip{}, ErrNotFound
	}

	return r.assemblePayslip(row), nil
}

func (r *MockSQLRepository) AvailableMocks() []string {
	names := make([]string, 0, len(r.rows.Aliases))
	seen := map[string]bool{}
	for _, alias := range r.rows.Aliases {
		name := normalizeKey(alias.MockName)
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (r *MockSQLRepository) findPayslipRow(query Query) (PayslipRow, bool) {
	query.MockName = normalizeKey(query.MockName)
	query.SlipID = normalizeKey(query.SlipID)
	query.CompanyID = normalizeKey(query.CompanyID)
	query.EmployeeID = strings.TrimSpace(query.EmployeeID)
	query.PeriodID = normalizeKey(query.PeriodID)
	query.Period = strings.TrimSpace(query.Period)
	query.SlipNo = strings.TrimSpace(query.SlipNo)

	if query.SlipID != "" {
		return r.findPayslipRowBySlipID(query.SlipID)
	}

	if query.MockName != "" {
		if slipID, ok := r.resolveAlias(query.MockName); ok {
			return r.findPayslipRowBySlipID(slipID)
		}
		return PayslipRow{}, false
	}

	for _, row := range r.rows.Payslips {
		if query.CompanyID != "" && normalizeKey(row.CompanyID) != query.CompanyID {
			continue
		}
		if query.EmployeeID != "" && row.EmployeeID != query.EmployeeID {
			continue
		}
		if query.PeriodID != "" && normalizeKey(row.PeriodID) != query.PeriodID {
			continue
		}
		if query.Period != "" {
			period, ok := r.findPeriod(row.PeriodID)
			if !ok || period.Label != query.Period {
				continue
			}
		}
		if query.SlipNo != "" && row.SlipNo != query.SlipNo {
			continue
		}
		return row, true
	}

	return PayslipRow{}, false
}

func (r *MockSQLRepository) findPayslipRowBySlipID(slipID string) (PayslipRow, bool) {
	slipID = normalizeKey(slipID)
	for _, row := range r.rows.Payslips {
		if normalizeKey(row.SlipID) == slipID {
			return row, true
		}
	}
	return PayslipRow{}, false
}

func (r *MockSQLRepository) resolveAlias(alias string) (string, bool) {
	alias = normalizeKey(alias)
	for _, row := range r.rows.Aliases {
		if normalizeKey(row.MockName) == alias {
			return row.SlipID, true
		}
	}
	return "", false
}

func (r *MockSQLRepository) assemblePayslip(row PayslipRow) reportpayslip.Payslip {
	company, _ := r.findCompany(row.CompanyID)
	employee, _ := r.findEmployee(row.EmployeeID)
	period, _ := r.findPeriod(row.PeriodID)

	report := reportpayslip.Report{
		Company: reportpayslip.Company{
			Name: company.Name,
			Logo: company.LogoURL,
		},
		Document: reportpayslip.Document{
			Title:                row.DocumentTitle,
			ConfidentialTitle:    row.ConfidentialTitle,
			ConfidentialSubTitle: row.ConfidentialSubtitle,
		},
		Payroll: reportpayslip.Payroll{
			Period:  period.Label,
			PayDate: period.PayDate,
			SlipNo:  row.SlipNo,
		},
		Employee: reportpayslip.Employee{
			Name:       employee.Name,
			EmpID:      employee.EmpNo,
			JoinedDate: employee.JoinedDate,
			Division:   employee.Division,
			Department: employee.Department,
			Section:    employee.Section,
			Position:   employee.Position,
			TaxID:      employee.TaxID,
		},
		Totals: reportpayslip.Totals{
			TotalIncome:   row.TotalIncome,
			TotalDeduct:   row.TotalDeduct,
			NetPay:        row.NetPay,
			BankAccountNo: row.BankAccountNo,
		},
	}

	items := r.lineItemsForSlip(row.SlipID)
	for _, item := range items {
		lineItem := reportpayslip.LineItem{Label: item.Label, Value: item.Value}
		switch item.Section {
		case "accumulations":
			report.Accumulations = append(report.Accumulations, lineItem)
		case "earnings":
			report.Earnings = append(report.Earnings, lineItem)
		case "deductions":
			report.Deductions = append(report.Deductions, lineItem)
		case "work_stats":
			report.WorkStats = append(report.WorkStats, lineItem)
		default:
			if report.Sections == nil {
				report.Sections = map[string][]reportpayslip.LineItem{}
			}
			report.Sections[item.Section] = append(report.Sections[item.Section], lineItem)
		}
	}

	for _, text := range r.rows.Texts {
		if normalizeKey(text.SlipID) != normalizeKey(row.SlipID) {
			continue
		}
		if report.Texts == nil {
			report.Texts = map[string]string{}
		}
		report.Texts[text.Key] = text.Value
	}

	return reportpayslip.Payslip{
		TemplateID: row.TemplateID,
		Report:     report,
	}
}

func (r *MockSQLRepository) findCompany(companyID string) (CompanyRow, bool) {
	companyID = normalizeKey(companyID)
	for _, row := range r.rows.Companies {
		if normalizeKey(row.CompanyID) == companyID {
			return row, true
		}
	}
	return CompanyRow{}, false
}

func (r *MockSQLRepository) findEmployee(employeeID string) (EmployeeRow, bool) {
	for _, row := range r.rows.Employees {
		if row.EmployeeID == employeeID {
			return row, true
		}
	}
	return EmployeeRow{}, false
}

func (r *MockSQLRepository) findPeriod(periodID string) (PayrollPeriodRow, bool) {
	periodID = normalizeKey(periodID)
	for _, row := range r.rows.PayrollPeriods {
		if normalizeKey(row.PeriodID) == periodID {
			return row, true
		}
	}
	return PayrollPeriodRow{}, false
}

func (r *MockSQLRepository) lineItemsForSlip(slipID string) []LineItemRow {
	slipID = normalizeKey(slipID)
	items := make([]LineItemRow, 0)
	for _, item := range r.rows.LineItems {
		if normalizeKey(item.SlipID) == slipID {
			items = append(items, item)
		}
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Section == items[j].Section {
			return items[i].SortOrder < items[j].SortOrder
		}
		return items[i].Section < items[j].Section
	})
	return items
}

func appendLineItems(rows []LineItemRow, slipID, section string, items []reportpayslip.LineItem) []LineItemRow {
	for i, item := range items {
		rows = append(rows, LineItemRow{
			SlipID:    slipID,
			Section:   section,
			Label:     item.Label,
			Value:     item.Value,
			SortOrder: i + 1,
		})
	}
	return rows
}

func sortedMapKeys[T any](values map[string]T) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func normalizeKey(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	value = strings.ReplaceAll(value, " ", "_")
	value = strings.ReplaceAll(value, "-", "_")
	return value
}

func defaultString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
