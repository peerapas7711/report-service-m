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

func TestRenderHTMLThaiDemarPayslip(t *testing.T) {
	data, err := DefaultPayslip("thai_demar")
	if err != nil {
		t.Fatalf("load thai demar payslip: %v", err)
	}

	html, err := RenderHTML("thai_demar", data)
	if err != nil {
		t.Fatalf("render thai demar html: %v", err)
	}

	if !strings.Contains(html, "Thai Demar") {
		t.Fatal("thai demar html missing company name")
	}
	if !strings.Contains(html, `class="page payslip payslip-thai-demar"`) {
		t.Fatal("thai demar html missing type class")
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
