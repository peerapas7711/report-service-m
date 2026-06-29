package payslip_html

import (
	"strings"
	"testing"

	"report-service-m/internal/reports/payslip"
)

func TestRenderHTMLTigersoftUsesAppIconLogo(t *testing.T) {
	data, err := payslip.LoadFromFile("../../../mock/payslip_tigersoft.json")
	if err != nil {
		t.Fatalf("load tigersoft mock: %v", err)
	}

	html, err := RenderHTML("tigersoft", data)
	if err != nil {
		t.Fatalf("render tigersoft html: %v", err)
	}

	if !strings.Contains(html, `<img class="logo wide-logo" src="data:image/png;base64,`) {
		t.Fatal("tigersoft html missing embedded app icon logo")
	}
	if strings.Contains(html, data.Report.Company.Logo) {
		t.Fatal("tigersoft html still uses mock company logo URL")
	}
}
