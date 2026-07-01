package payslip_html

import (
	"strings"
	"testing"
)

func TestRenderHTMLDefaultPayslip(t *testing.T) {
	data, err := DefaultPayslip("default")
	if err != nil {
		t.Fatalf("load default payslip: %v", err)
	}

	html, err := RenderHTML("", data)
	if err != nil {
		t.Fatalf("render default html: %v", err)
	}

	if !strings.Contains(html, "Tigersoft") {
		t.Fatal("default html missing company name")
	}
	if !strings.Contains(html, "ใบแจ้งเงินเดือน (Payroll Slip)") {
		t.Fatal("default html missing document title")
	}
	if !strings.Contains(html, `class="page payslip payslip-tigersoft"`) {
		t.Fatal("default html missing tigersoft type class")
	}
}

func TestRenderHTMLThaiDelmarPayslip(t *testing.T) {
	data, err := DefaultPayslip("thai_delmar")
	if err != nil {
		t.Fatalf("load thai delmar payslip: %v", err)
	}

	html, err := RenderHTML("thai_delmar", data)
	if err != nil {
		t.Fatalf("render thai delmar html: %v", err)
	}

	if !strings.Contains(html, "บริษัท ไทยเดลมาร์ จำกัด") {
		t.Fatal("thai delmar html missing company name")
	}
	if !strings.Contains(html, `class="slip"`) {
		t.Fatal("thai delmar html missing slip wrapper")
	}
	if !strings.Contains(html, `class="val-box"`) {
		t.Fatal("thai delmar html missing inline template value boxes")
	}
	if !strings.Contains(html, "data:image/png;base64,") {
		t.Fatal("thai delmar html missing embedded logo image")
	}
	if strings.Contains(html, "ขอบกระดาษ A4") {
		t.Fatal("thai delmar html should not render A4 edge note")
	}
}

func TestRenderHTMLCPPayslip(t *testing.T) {
	data, err := DefaultPayslip("cp")
	if err != nil {
		t.Fatalf("load cp payslip: %v", err)
	}

	html, err := RenderHTML("cp", data)
	if err != nil {
		t.Fatalf("render cp html: %v", err)
	}

	if !strings.Contains(html, "CP Group") {
		t.Fatal("cp html missing company name")
	}
	if !strings.Contains(html, `class="page payslip payslip-cp"`) {
		t.Fatal("cp html missing type class")
	}
}

func TestNormalizeTemplateTypeRejectsUnknownType(t *testing.T) {
	if _, err := NormalizeTemplateType("missing"); err == nil {
		t.Fatal("expected unknown payslip type error")
	}
}

func TestNormalizeTemplateTypeRejectsTigersoftType(t *testing.T) {
	if _, err := NormalizeTemplateType("tigersoft"); err == nil {
		t.Fatal("expected tigersoft to be unavailable as a template folder")
	}
}
