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
	app := New(testConfig())

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

func TestReportPayslipHTMLUsesDefaultType(t *testing.T) {
	app := New(testConfig())

	req := httptest.NewRequest(http.MethodGet, "/report/payslip/html", nil)
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

	body := readBody(t, resp)
	if !strings.Contains(body, "Tigersoft") {
		t.Fatal("html response missing default payslip data")
	}
	if !strings.Contains(body, `class="page payslip payslip-tigersoft"`) {
		t.Fatal("html response missing default payslip type class")
	}
}

func TestReportPayslipHTMLUsesThaiDemarTypeParam(t *testing.T) {
	app := New(testConfig())

	req := httptest.NewRequest(http.MethodGet, "/report/payslip/html?type=thai_demar", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("report payslip html request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}

	body := readBody(t, resp)
	if !strings.Contains(body, "Thai Demar") {
		t.Fatal("html response missing thai demar payslip data")
	}
	if !strings.Contains(body, `class="page payslip payslip-thai-demar"`) {
		t.Fatal("html response missing thai demar payslip type class")
	}
}

func TestReportPayslipHTMLRendersBatchCount(t *testing.T) {
	app := New(testConfig())

	req := httptest.NewRequest(http.MethodGet, "/report/payslip/html?type=default&count=3", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("report payslip html batch request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}

	html := readBody(t, resp)
	if got := strings.Count(html, `class="page payslip payslip-tigersoft"`); got != 3 {
		t.Fatalf("unexpected payslip page count: %d", got)
	}
	if !strings.Contains(html, "@font-face") {
		t.Fatal("html batch response missing embedded stylesheet")
	}
	if !strings.Contains(html, "Demo Employee 0003") {
		t.Fatal("html batch response missing sequenced employee data")
	}
}

func TestReportPayslipHTMLUnknownType(t *testing.T) {
	app := New(testConfig())

	req := httptest.NewRequest(http.MethodGet, "/report/payslip/html?type=missing", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("report payslip html request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}

	body := readBody(t, resp)
	if !strings.Contains(body, "available_types") {
		t.Fatal("error response missing available types")
	}
}

func testConfig() config.Config {
	return config.Config{
		AppName:     "report-service-test",
		Environment: "test",
		BodyLimit:   1 << 20,
	}
}

func readBody(t *testing.T, resp *http.Response) string {
	t.Helper()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response body: %v", err)
	}
	return string(body)
}
