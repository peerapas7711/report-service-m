package httpserver

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"report-service-m/internal/config"
)

func TestHealthRoute(t *testing.T) {
	app := New(config.Config{
		AppName:     "report-service-test",
		Environment: "test",
		BodyLimit:   1 << 20,
	})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("health request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}
}

func TestPayslipPageRoute(t *testing.T) {
	app := New(config.Config{
		AppName:     "report-service-test",
		Environment: "test",
		BodyLimit:   1 << 20,
	})

	req := httptest.NewRequest(http.MethodGet, "/payslip", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("payslip page request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}
	if got := resp.Header.Get("Content-Type"); !strings.HasPrefix(got, "text/html") {
		t.Fatalf("unexpected content type: %s", got)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read payslip page: %v", err)
	}
	if !strings.Contains(string(body), `action="/preview/payslip"`) {
		t.Fatal("payslip page missing preview form")
	}
}

func TestPreviewPayslipDownloadDisposition(t *testing.T) {
	app := New(config.Config{
		AppName:     "report-service-test",
		Environment: "test",
		BodyLimit:   1 << 20,
	})

	req := httptest.NewRequest(http.MethodGet, "/preview/payslip?mock=hopinn&download=1", nil)

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("preview payslip request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}
	if got := resp.Header.Get("Content-Type"); got != "application/pdf" {
		t.Fatalf("unexpected content type: %s", got)
	}
	if got := resp.Header.Get("Content-Disposition"); !strings.HasPrefix(got, "attachment;") {
		t.Fatalf("unexpected disposition: %s", got)
	}
}
