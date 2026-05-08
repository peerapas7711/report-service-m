package payslip

import (
	"encoding/json"
	"log"
	"os"
)

type Payslip struct {
	TemplateID string `json:"template_id,omitempty"`
	Report     Report `json:"report"`
}

type Report struct {
	Company       Company               `json:"company"`
	Document      Document              `json:"document"`
	Payroll       Payroll               `json:"payroll"`
	Employee      Employee              `json:"employee"`
	Accumulations []LineItem            `json:"accumulations"`
	Earnings      []LineItem            `json:"earnings"`
	Deductions    []LineItem            `json:"deductions"`
	WorkStats     []LineItem            `json:"work_stats"`
	Totals        Totals                `json:"totals"`
	Texts         map[string]string     `json:"texts,omitempty"`
	Sections      map[string][]LineItem `json:"sections,omitempty"`
}

type Company struct {
	Name string `json:"name"`
	Logo string `json:"logo"`
}

type Document struct {
	Title                string `json:"title"`
	ConfidentialTitle    string `json:"confidential_title"`
	ConfidentialSubTitle string `json:"confidential_subtitle"`
}

type Payroll struct {
	Period  string `json:"period"`
	PayDate string `json:"pay_date"`
	SlipNo  string `json:"slip_no"`
}

type Employee struct {
	Name       string `json:"name"`
	EmpID      string `json:"emp_id"`
	JoinedDate string `json:"joined_date"`
	Division   string `json:"division"`
	Department string `json:"department"`
	Section    string `json:"section"`
	Position   string `json:"position"`
	TaxID      string `json:"tax_id"`
}

type LineItem struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

type Totals struct {
	TotalIncome   string `json:"total_income"`
	TotalDeduct   string `json:"total_deduct"`
	NetPay        string `json:"net_pay"`
	BankAccountNo string `json:"bank_account_no"`
}

func MustPayslipFromFile(path string) Payslip {
	data, err := LoadFromFile(path)
	if err != nil {
		log.Fatal(err)
	}

	return data
}

func LoadFromFile(path string) (Payslip, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Payslip{}, err
	}

	var data Payslip
	if err := json.Unmarshal(b, &data); err != nil {
		return Payslip{}, err
	}

	return data, nil
}
