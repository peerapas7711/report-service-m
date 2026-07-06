package httpserver

import (
	"archive/zip"
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
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
	sheetXML := xlsxPart(t, xlsx, "xl/worksheets/sheet1.xml")
	if !strings.Contains(sheetXML, "วันที่") {
		t.Fatal("excel response missing single mock header")
	}
	if !strings.Contains(sheetXML, "01/06/2569") {
		t.Fatal("excel response missing first single mock date")
	}
	if !strings.Contains(sheetXML, "<t>01</t>") {
		t.Fatal("excel response should preserve shift code leading zero")
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
	workbookXML := xlsxPart(t, xlsx, "xl/workbook.xml")
	if !strings.Contains(workbookXML, `name="All"`) {
		t.Fatal("excel response missing wrapped sheet name")
	}

	sheetXML := xlsxPart(t, xlsx, "xl/worksheets/sheet1.xml")
	if !strings.Contains(sheetXML, "รหัสพนักงาน") {
		t.Fatal("excel response missing multi mock employee code header")
	}
	if !strings.Contains(sheetXML, "เด็กชายเพลไรท์ อิอิ") {
		t.Fatal("excel response missing multi mock employee name")
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

func xlsxPart(t *testing.T, data []byte, name string) string {
	t.Helper()

	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("open xlsx zip: %v", err)
	}

	for _, file := range zr.File {
		if file.Name != name {
			continue
		}

		rc, err := file.Open()
		if err != nil {
			t.Fatalf("open xlsx part %s: %v", name, err)
		}
		defer rc.Close()

		body, err := io.ReadAll(rc)
		if err != nil {
			t.Fatalf("read xlsx part %s: %v", name, err)
		}
		return string(body)
	}

	t.Fatalf("xlsx part not found: %s", name)
	return ""
}
