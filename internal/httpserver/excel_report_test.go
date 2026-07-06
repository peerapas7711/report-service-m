package httpserver

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/xuri/excelize/v2"
)

func TestPostReportExcelRendersSingleAfterProcessMock(t *testing.T) {
	app := New(testConfig())
	body := readMockJSON(t, "AfterProcess_20260630_1700 1.json")

	req := httptest.NewRequest(http.MethodPost, "/report/excel?filename=afterprocess_single", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("post report excel request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}
	if got := resp.Header.Get("Content-Type"); got != "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" {
		t.Fatalf("unexpected content type: %s", got)
	}
	if got := resp.Header.Get("Content-Disposition"); !strings.Contains(got, `filename="afterprocess_single.xlsx"`) {
		t.Fatalf("unexpected content disposition: %s", got)
	}

	xlsx := readBodyBytes(t, resp)
	file := openWorkbook(t, xlsx)
	defer func() {
		_ = file.Close()
	}()

	if got := cellValue(t, file, "AfterProcess", "A1"); got != "วันที่" {
		t.Fatal("excel response missing single mock header")
	}
	if got := cellValue(t, file, "AfterProcess", "A2"); got != "01/06/2569" {
		t.Fatalf("excel response missing first single mock date: %q", got)
	}
	if got := cellValue(t, file, "AfterProcess", "B2"); got != "01" {
		t.Fatalf("excel response should preserve shift code leading zero: %q", got)
	}
}

func TestPostReportExcelRendersWrappedMultiAfterProcessMock(t *testing.T) {
	app := New(testConfig())
	mock := readMockJSON(t, "AfterProcess_All_20260630_1707.json")
	body := []byte(`{"filename":"afterprocess_all","sheetName":"All","data":` + string(mock) + `}`)

	req := httptest.NewRequest(http.MethodPost, "/v1/reports/excel", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("post v1 report excel request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}

	xlsx := readBodyBytes(t, resp)
	file := openWorkbook(t, xlsx)
	defer func() {
		_ = file.Close()
	}()

	if got := file.GetSheetList()[0]; got != "All" {
		t.Fatal("excel response missing wrapped sheet name")
	}
	if got := cellValue(t, file, "All", "A1"); got != "รหัสพนักงาน" {
		t.Fatalf("excel response missing multi mock employee code header: %q", got)
	}
	if got := cellValue(t, file, "All", "B2"); got != "เด็กชายเพลไรท์ อิอิ" {
		t.Fatalf("excel response missing multi mock employee name: %q", got)
	}
}

func readMockJSON(t *testing.T, name string) []byte {
	t.Helper()

	body, err := os.ReadFile("../../mock/" + name)
	if err != nil {
		t.Fatalf("read mock json %s: %v", name, err)
	}
	return body
}

func readBodyBytes(t *testing.T, resp *http.Response) []byte {
	t.Helper()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response body: %v", err)
	}
	return body
}

func openWorkbook(t *testing.T, data []byte) *excelize.File {
	t.Helper()

	file, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("open xlsx workbook: %v", err)
	}
	return file
}

func cellValue(t *testing.T, file *excelize.File, sheet, cell string) string {
	t.Helper()

	value, err := file.GetCellValue(sheet, cell)
	if err != nil {
		t.Fatalf("read cell %s!%s: %v", sheet, cell, err)
	}
	return value
}
