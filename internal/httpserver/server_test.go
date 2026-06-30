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

func TestPostReportPayslipUsesRequestFormatHTML(t *testing.T) {
	app := New(testConfig())

	req := httptest.NewRequest(
		http.MethodPost,
		"/report/payslip",
		strings.NewReader(strings.Replace(actualPayslipReportJSON, `"format":"pdf"`, `"format":"html"`, 1)),
	)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("post report payslip request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}
	if got := resp.Header.Get("Content-Type"); !strings.HasPrefix(got, "text/html") {
		t.Fatalf("unexpected content type: %s", got)
	}

	body := readBody(t, resp)
	if !strings.Contains(body, "บริษัท ตังค์ จำกัด") {
		t.Fatal("html response missing posted company data")
	}
	if !strings.Contains(body, "Mr.Prasert  TEST") {
		t.Fatal("html response missing first posted employee data")
	}
	if !strings.Contains(body, "Ms.Payload  TEST") {
		t.Fatal("html response missing second posted employee data")
	}
	if strings.Count(body, `class="page payslip payslip-thai-demar"`) != 2 {
		t.Fatal("html response should render one payslip page per employee")
	}
	if !strings.Contains(body, "25,324.00") {
		t.Fatal("html response missing formatted earning amount")
	}
}

func TestPostReportPayslipHTMLPathOverridesRequestFormat(t *testing.T) {
	app := New(testConfig())

	req := httptest.NewRequest(http.MethodPost, "/report/payslip/html", strings.NewReader(actualPayslipReportJSON))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("post report payslip html request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}
	if got := resp.Header.Get("Content-Type"); !strings.HasPrefix(got, "text/html") {
		t.Fatalf("unexpected content type: %s", got)
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

func TestPostReportPayslipHTMLUnknownTemplateCode(t *testing.T) {
	app := New(testConfig())

	body := strings.Replace(actualPayslipReportJSON, `"templateCode":"PAYROLL/PAYSLIP_THAI_DELMAR"`, `"templateCode":"PAYROLL/PAYSLIP_MISSING"`, 1)
	req := httptest.NewRequest(http.MethodPost, "/report/payslip/html", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("post report payslip html request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}

	responseBody := readBody(t, resp)
	if !strings.Contains(responseBody, "available_types") {
		t.Fatal("error response missing available types")
	}
}

func TestPostReportPayslipRejectsUnsupportedFormat(t *testing.T) {
	app := New(testConfig())

	body := strings.Replace(actualPayslipReportJSON, `"format":"pdf"`, `"format":"xlsx"`, 1)
	req := httptest.NewRequest(http.MethodPost, "/report/payslip", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("post report payslip request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}

	responseBody := readBody(t, resp)
	if !strings.Contains(responseBody, "unsupported report format") {
		t.Fatal("error response missing unsupported format message")
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

const actualPayslipReportJSON = `{
  "templateCode":"PAYROLL/PAYSLIP_THAI_DELMAR",
  "format":"pdf",
  "companyCode":"1",
  "data":{
    "company":"บริษัท ตังค์ จำกัด",
    "companyEn":"Tung Co. , Ltd.",
    "address":"700/359 ม.6",
    "tel":"038-743923",
    "logoUrl":"",
    "employees":[
      {
        "empCode":"0160",
        "empName":"Mr.Prasert  TEST",
        "position":"02-16 : UF-ALL / AF-ALL",
        "department":"MANU - Production Time : Production",
        "empType":"",
        "startDate":"28/08/2543",
        "salary":0,
        "bankName":"Bank of Ayudhaya Public Company Limited",
        "bankAccount":"4474029128",
        "year":"2569",
        "month":"June",
        "periodNo":"01",
        "payDate":"30/Jun/2026",
        "earnings":[
          {"name":"OT3","amount":25324},
          {"name":"Salary","amount":25324}
        ],
        "deductions":[],
        "totalEarnings":50648,
        "totalDeductions":0,
        "netPay":22041.6,
        "ytdIncome":66247.31,
        "ytdTax":0,
        "ytdSocialSecurity":1625,
        "ytdProvidentFund":5064.8,
        "ytdProvidentFundCompany":3038.88,
        "taxAllowances":{
          "expenses":[
            {"name":"ค่าใช้จ่าย","amount":60000},
            {"name":"ลดหย่อนผู้มีรายได้","amount":100000}
          ],
          "allowances":[
            {"name":"ประกันสังคม","amount":6125},
            {"name":"กองทุนสำรองเลี้ยงชีพ","amount":10259.2}
          ]
        }
      },
      {
        "empCode":"0187",
        "empName":"Ms.Payload  TEST",
        "position":"03-01 : AF-FA / AF-HV",
        "department":"MANU - Production Time : Production",
        "empType":"",
        "startDate":"12/03/2544",
        "salary":0,
        "bankName":"Bank of Ayudhaya Public Company Limited",
        "bankAccount":"4474029144",
        "year":"2569",
        "month":"June",
        "periodNo":"01",
        "payDate":"30/Jun/2026",
        "earnings":[
          {"name":"OT3","amount":23998},
          {"name":"Salary","amount":23998},
          {"name":"Position","amount":-3000}
        ],
        "deductions":[],
        "totalEarnings":44996,
        "totalDeductions":0,
        "netPay":17848.2,
        "ytdIncome":61375.08,
        "ytdTax":0,
        "ytdSocialSecurity":1625,
        "ytdProvidentFund":4799.6,
        "ytdProvidentFundCompany":2879.76,
        "taxAllowances":{
          "expenses":[
            {"name":"ค่าใช้จ่าย","amount":60000},
            {"name":"ลดหย่อนผู้มีรายได้","amount":98082.34}
          ],
          "allowances":[
            {"name":"ประกันสังคม","amount":6125},
            {"name":"กองทุนสำรองเลี้ยงชีพ","amount":9198.4}
          ]
        }
      }
    ]
  }
}`
