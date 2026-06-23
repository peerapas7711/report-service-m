package payslipdata

import (
	"context"
	"errors"
	"testing"

	reportpayslip "report-service-m/internal/reports/payslip"
)

func TestMockSQLRepositoryFindsPayslipByAlias(t *testing.T) {
	repo := NewMockSQLRepository(RowsFromPayslips([]MockPayslipSeed{
		{
			MockName: "hopinn",
			Aliases:  []string{"1", "default"},
			Data:     samplePayslip(),
		},
	}))

	got, err := repo.FindPayslip(context.Background(), Query{MockName: "1"})
	if err != nil {
		t.Fatalf("find payslip: %v", err)
	}

	if got.Report.Company.Name != "Example Co., Ltd." {
		t.Fatalf("unexpected company: %q", got.Report.Company.Name)
	}
	if got.Report.Employee.EmpID != "EMP-001" {
		t.Fatalf("unexpected employee id: %q", got.Report.Employee.EmpID)
	}
	if got.Report.Earnings[0].Label != "Salary" {
		t.Fatalf("unexpected earning: %#v", got.Report.Earnings)
	}
}

func TestMockSQLRepositoryFindsPayslipBySQLLikeFilters(t *testing.T) {
	repo := NewMockSQLRepository(RowsFromPayslips([]MockPayslipSeed{
		{
			MockName: "hopinn",
			Data:     samplePayslip(),
		},
	}))

	got, err := repo.FindPayslip(context.Background(), Query{
		CompanyID:  "company_hopinn",
		EmployeeID: "EMP-001",
		Period:     "03/2026",
		SlipNo:     "7",
	})
	if err != nil {
		t.Fatalf("find payslip: %v", err)
	}

	if got.Report.Totals.NetPay != "9,500.00" {
		t.Fatalf("unexpected net pay: %q", got.Report.Totals.NetPay)
	}
}

func TestMockSQLRepositoryReturnsNotFound(t *testing.T) {
	repo := NewMockSQLRepository(RowsFromPayslips([]MockPayslipSeed{
		{
			MockName: "hopinn",
			Data:     samplePayslip(),
		},
	}))

	_, err := repo.FindPayslip(context.Background(), Query{MockName: "missing"})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func samplePayslip() reportpayslip.Payslip {
	return reportpayslip.Payslip{
		TemplateID: "modern",
		Report: reportpayslip.Report{
			Company: reportpayslip.Company{
				Name: "Example Co., Ltd.",
				Logo: "https://example.test/logo.png",
			},
			Document: reportpayslip.Document{
				Title:                "PAY SLIP",
				ConfidentialTitle:    "PRIVATE & CONFIDENTIAL",
				ConfidentialSubTitle: "(TO BE OPENED BY ADDRESSEE ONLY)",
			},
			Payroll: reportpayslip.Payroll{
				Period:  "03/2026",
				PayDate: "31/03/2026",
				SlipNo:  "7",
			},
			Employee: reportpayslip.Employee{
				Name:       "Test Employee",
				EmpID:      "EMP-001",
				JoinedDate: "01/01/2026",
				Division:   "Technology",
				Department: "Platform",
				Section:    "Backend",
				Position:   "Engineer",
				TaxID:      "1234567890123",
			},
			Earnings: []reportpayslip.LineItem{
				{Label: "Salary", Value: "10,000.00"},
			},
			Deductions: []reportpayslip.LineItem{
				{Label: "Tax", Value: "500.00"},
			},
			Totals: reportpayslip.Totals{
				TotalIncome:   "10,000.00",
				TotalDeduct:   "500.00",
				NetPay:        "9,500.00",
				BankAccountNo: "123-4-56789-0",
			},
		},
	}
}
