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

func TestReportPayslipHTMLUsesMockSQLRepository(t *testing.T) {
	app := New(config.Config{
		AppName:     "report-service-test",
		Environment: "test",
		BodyLimit:   1 << 20,
	})

	req := httptest.NewRequest(http.MethodGet, "/report/payslip/html?mock=hopinn", nil)

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("report payslip html request: %v", err)
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
		t.Fatalf("read report payslip html: %v", err)
	}
	if !strings.Contains(string(body), "Hop Inn Hotel Public Company Limited") {
		t.Fatal("html response missing SQL-like mock payslip data")
	}
}

func TestReportPayslipHTMLRendersBatchCount(t *testing.T) {
	app := New(config.Config{
		AppName:     "report-service-test",
		Environment: "test",
		BodyLimit:   1 << 20,
	})

	req := httptest.NewRequest(http.MethodGet, "/report/payslip/html?mock=tigersoft&count=3", nil)

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("report payslip html batch request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read report payslip html batch: %v", err)
	}

	html := string(body)
	if got := strings.Count(html, `class="page payslip payslip-tigersoft"`); got != 3 {
		t.Fatalf("unexpected payslip page count: %d", got)
	}
	if !strings.Contains(html, "Peerapat S. 0003") {
		t.Fatal("html batch response missing sequenced employee data")
	}
}

func TestReportPayslipHTMLUnknownMock(t *testing.T) {
	app := New(config.Config{
		AppName:     "report-service-test",
		Environment: "test",
		BodyLimit:   1 << 20,
	})

	req := httptest.NewRequest(http.MethodGet, "/report/payslip/html?mock=missing", nil)

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("report payslip html request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}
}
