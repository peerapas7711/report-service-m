package payslip

import (
	"bytes"
	"encoding/base64"
	"io"
	"net/http"
	"testing"
)

func TestRenderPayslipMocks(t *testing.T) {
	mockFiles := []string{
		"../../mock/payslip_hopinn.json",
		"../../mock/payslip_tigersoft.json",
		"../../mock/payslip_bluewave.json",
	}
	orientations := []string{"P", "L"}

	for _, path := range mockFiles {
		data := MustPayslipFromFile(path)

		for _, orientation := range orientations {
			pdfBytes, err := Render(data, orientation)
			if err != nil {
				t.Fatalf("render %s orientation %s: %v", path, orientation, err)
			}

			if len(pdfBytes) == 0 {
				t.Fatalf("render %s orientation %s returned empty pdf", path, orientation)
			}
		}
	}
}

func TestRenderPayslipLogoFromURL(t *testing.T) {
	data := MustPayslipFromFile("../../mock/payslip_hopinn.json")

	pngBytes, err := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVQIHWP4////fwAJ+wP9KobjigAAAABJRU5ErkJggg==")
	if err != nil {
		t.Fatalf("decode png: %v", err)
	}

	originalClient := imageHTTPClient
	imageHTTPClient = &http.Client{
		Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"image/png"}},
				Body:       io.NopCloser(bytes.NewReader(pngBytes)),
			}, nil
		}),
	}
	defer func() {
		imageHTTPClient = originalClient
	}()

	data.Report.Company.Logo = "https://example.com/logo.png"

	pdfBytes, err := Render(data, "P")
	if err != nil {
		t.Fatalf("render url logo: %v", err)
	}

	if len(pdfBytes) == 0 {
		t.Fatal("render url logo returned empty pdf")
	}
}

func TestNormalizeOrientationDefaultsToPortrait(t *testing.T) {
	if got := normalizeOrientation(""); got != "P" {
		t.Fatalf("unexpected default orientation: %s", got)
	}

	if got := normalizeOrientation("portrait"); got != "P" {
		t.Fatalf("unexpected portrait orientation: %s", got)
	}

	if got := normalizeOrientation("landscape"); got != "L" {
		t.Fatalf("unexpected landscape orientation: %s", got)
	}
}

func TestBuildCompensationSummaryFallsBackToLineItems(t *testing.T) {
	report := Report{
		Earnings: []LineItem{
			{Label: "Salary", Value: "30,000.00"},
			{Label: "Allowance", Value: "1,250.00"},
		},
		Deductions: []LineItem{
			{Label: "Tax", Value: "500.00"},
			{Label: "Social Security", Value: "750.00"},
		},
		Totals: Totals{
			BankAccountNo: "123-4-56789-0",
		},
	}

	summary := buildCompensationSummary(report)

	if summary.TotalIncome != "31,250.00" {
		t.Fatalf("unexpected total income: %s", summary.TotalIncome)
	}
	if summary.TotalDeduct != "1,250.00" {
		t.Fatalf("unexpected total deduct: %s", summary.TotalDeduct)
	}
	if summary.NetPay != "30,000.00" {
		t.Fatalf("unexpected net pay: %s", summary.NetPay)
	}
	if summary.BankAccountNo != "123-4-56789-0" {
		t.Fatalf("unexpected bank account no: %s", summary.BankAccountNo)
	}
}

func TestBuildWorkSummary(t *testing.T) {
	report := Report{
		WorkStats: []LineItem{
			{Label: "Works", Value: "30:00:00"},
			{Label: "Absent", Value: "00:00:00"},
			{Label: "OT 1", Value: "06:00"},
			{Label: "OT 1.5", Value: "04:30"},
			{Label: "OT 2", Value: "00:00"},
		},
	}

	summary := buildWorkSummary(report)

	if summary.OvertimeTotal != "10:30" {
		t.Fatalf("unexpected overtime total: %s", summary.OvertimeTotal)
	}
	if summary.AbsentValue != "00:00:00" {
		t.Fatalf("unexpected absent value: %s", summary.AbsentValue)
	}
}

func TestCalculateCompensationRowGapClampsSpacing(t *testing.T) {
	if got := calculateCompensationRowGap(60, 3); got != 6.8 {
		t.Fatalf("unexpected capped row gap: %v", got)
	}

	if got := calculateCompensationRowGap(60, 20); got != 5.6 {
		t.Fatalf("unexpected minimum row gap: %v", got)
	}
}

func TestParseDuration(t *testing.T) {
	tests := []struct {
		name           string
		value          string
		wantSeconds    int
		wantHasSeconds bool
		wantOK         bool
	}{
		{
			name:           "hours and minutes",
			value:          "06:30",
			wantSeconds:    23400,
			wantHasSeconds: false,
			wantOK:         true,
		},
		{
			name:           "hours minutes seconds",
			value:          "01:02:03",
			wantSeconds:    3723,
			wantHasSeconds: true,
			wantOK:         true,
		},
		{
			name:   "non duration",
			value:  "1 Day",
			wantOK: false,
		},
	}

	for _, tt := range tests {
		seconds, hasSeconds, ok := parseDuration(tt.value)
		if ok != tt.wantOK {
			t.Fatalf("%s: unexpected ok=%v", tt.name, ok)
		}
		if !tt.wantOK {
			continue
		}
		if seconds != tt.wantSeconds {
			t.Fatalf("%s: unexpected seconds=%d", tt.name, seconds)
		}
		if hasSeconds != tt.wantHasSeconds {
			t.Fatalf("%s: unexpected hasSeconds=%v", tt.name, hasSeconds)
		}
	}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (fn roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}
