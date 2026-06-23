package payslipdata

import (
	"context"
	"errors"

	reportpayslip "report-service-m/internal/reports/payslip"
)

var ErrNotFound = errors.New("payslip not found")

type Query struct {
	MockName   string
	SlipID     string
	CompanyID  string
	EmployeeID string
	PeriodID   string
	Period     string
	SlipNo     string
}

type Repository interface {
	FindPayslip(ctx context.Context, query Query) (reportpayslip.Payslip, error)
}

type MockCatalog interface {
	AvailableMocks() []string
}
