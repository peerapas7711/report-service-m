package systemaccesspermission

import (
	"bytes"
	"encoding/json"
	"log"
	"os"
	"report-service-m/report/fontmanager"

	"github.com/jung-kurt/gofpdf"
)

func SystemAccessPermission(data SystemAccess, orientationStr, lang string) ([]byte, error) {
	orientation := normalizeOrientation(orientationStr)
	pdf := gofpdf.New(orientation, "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)

	if err := fontmanager.LoadAll(pdf); err != nil {
		return nil, err
	}

	pdf.AddPage()

	title := pickTitle(data.Report.Title, lang)
	fontmanager.Set(pdf, lang, "", 14)
	pdf.CellFormat(0, 10, title, "", 1, "C", false, 0, "")
	pdf.Ln(2)

	company := data.Report.Company.Code + " - " + pickName(data.Report.Company.Name, lang)
	fontmanager.Set(pdf, lang, "BU", 10)
	pdf.CellFormat(0, 6, company, "", 1, "", false, 0, "")
	pdf.Ln(1)

	group := pickName(data.Report.Group.Name, lang)
	fontmanager.Set(pdf, lang, "BU", 8)
	pdf.SetX(18)
	pdf.CellFormat(0, 6, group, "", 1, "", false, 0, "")

	// ----- table layout -----
	tableX := 18.0
	tableY := pdf.GetY() + 2

	// colCode := 30.0
	// colName := 120.0
	// colView := 30.0

	pageW, _ := pdf.GetPageSize()
	left, _, right, _ := pdf.GetMargins()
	contentW := pageW - left - right
	colCode := contentW * 0.20
	colView := contentW * 0.15
	colName := contentW - colCode - colView

	rowH := 8.0
	headerH := 8.0

	rows := len(data.Report.Menus)
	tableW := colCode + colName + colView
	tableH := headerH + float64(rows)*rowH

	// ----- draw outer border + vertical lines -----
	pdf.SetDrawColor(0, 0, 0)
	pdf.SetLineWidth(0.2)

	// กรอบนอก
	pdf.Rect(tableX, tableY, tableW, tableH, "D")

	// เส้นตั้งแบ่งคอลัมน์
	pdf.Line(tableX+colCode, tableY, tableX+colCode, tableY+tableH)
	pdf.Line(tableX+colCode+colName, tableY, tableX+colCode+colName, tableY+tableH)

	// เส้นใต้หัวตาราง
	pdf.Line(tableX, tableY+headerH, tableX+tableW, tableY+headerH)

	hCode := map[string]string{"th": "รหัสเมนู", "en": "Menu Code", "my": "မီနူးကုဒ်"}[lang]
	if hCode == "" {
		hCode = "Menu Code"
	}
	hName := map[string]string{"th": "ชื่อเมนู", "en": "Menu Name", "my": "မီနူးအမည်"}[lang]
	if hName == "" {
		hName = "Menu Name"
	}

	fontmanager.Set(pdf, lang, "", 9)

	// Code header (center)
	pdf.SetXY(tableX+2, tableY)
	pdf.CellFormat(colCode, headerH, hCode, "", 0, "L", false, 0, "")

	// Name header (center)
	pdf.SetXY(tableX+colCode+2, tableY)
	pdf.CellFormat(colName, headerH, hName, "", 0, "L", false, 0, "")

	// View header (center)
	pdf.SetXY(tableX+colCode+colName, tableY)
	pdf.CellFormat(colView, headerH, "View", "", 0, "C", false, 0, "")

	// ----- rows -----
	fontmanager.Set(pdf, lang, "", 9)

	for i, m := range data.Report.Menus {
		y := tableY + headerH + float64(i)*rowH

		// code
		pdf.SetXY(tableX+2, y) // +2 ให้มี padding
		pdf.CellFormat(colCode-4, rowH, m.Code, "", 0, "L", false, 0, "")

		// name
		name := pickName(m.Name, lang)
		name = ellipsisToWidth(pdf, name, colName-4) // -4 เผื่อ padding
		pdf.SetXY(tableX+colCode+2, y)
		pdf.CellFormat(colName-4, rowH, name, "", 0, "L", false, 0, "")

		// checkbox (center in view col)
		cbx := tableX + colCode + colName + (colView / 2) - 2.5
		cby := y + (rowH / 2) - 2.5
		drawCheckbox(pdf, cbx, cby, 5.0, m.View)
	}

	// เลื่อน cursor ลงหลังตาราง
	pdf.SetY(tableY + tableH + 4)
	// ? ออกไฟล์เลย -------
	// filename := "system_access_permission_" + lang + ".pdf"
	// if err := pdf.OutputFileAndClose(filename); err != nil {
	// 	log.Fatal(err)
	// }

	// ? API -------
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func normalizeOrientation(o string) string {
	switch o {
	case "P", "p", "portrait":
		return "P"
	case "L", "l", "landscape":
		return "L"
	default:
		return "P" // default
	}
}

func MustSystemAccessFromFile(path string) SystemAccess {
	b, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	var d SystemAccess
	if err := json.Unmarshal(b, &d); err != nil {
		log.Fatal(err)
	}
	return d
}
func pickTitle(t Title, lang string) string {
	switch lang {
	case "th":
		if t.Th != "" {
			return t.Th
		}
	case "my":
		if t.My != "" {
			return t.My
		}
	default:
		if t.En != "" {
			return t.En
		}
	}
	// fallback กันว่าง
	if t.En != "" {
		return t.En
	}
	if t.Th != "" {
		return t.Th
	}
	return t.My
}

func pickName(n Name, lang string) string {
	switch lang {
	case "th":
		if n.Th != "" {
			return n.Th
		}
	case "my":
		if n.My != "" {
			return n.My
		}
	default:
		if n.En != "" {
			return n.En
		}
	}
	if n.En != "" {
		return n.En
	}
	if n.Th != "" {
		return n.Th
	}
	return n.My
}

func drawCheckbox(pdf *gofpdf.Fpdf, x, y, size float64, checked bool) {
	// box
	pdf.SetDrawColor(15, 23, 42)
	pdf.SetLineWidth(0.25)
	pdf.Rect(x, y, size, size, "D")

	if !checked {
		return
	}

	// smoother checkmark
	pdf.SetLineWidth(0.8)

	// สำคัญ: ทำให้ปลายเส้น/มุม “มน”
	pdf.SetLineCapStyle("round")  // butt|round|square
	pdf.SetLineJoinStyle("round") // miter|round|bevel

	// ปรับสัดส่วนให้ดูคล้าย ✓ ใน UI ทั่วไป
	x1, y1 := x+size*0.20, y+size*0.55
	x2, y2 := x+size*0.42, y+size*0.76
	x3, y3 := x+size*0.82, y+size*0.24

	pdf.Line(x1, y1, x2, y2)
	pdf.Line(x2, y2, x3, y3)

	// reset (กันไปกระทบส่วนอื่น)
	pdf.SetLineWidth(0.25)
	pdf.SetLineCapStyle("butt")
	pdf.SetLineJoinStyle("miter")
}

func ellipsisToWidth(pdf *gofpdf.Fpdf, s string, maxW float64) string {
	if pdf.GetStringWidth(s) <= maxW {
		return s
	}
	ellipsis := "..."
	avail := maxW - pdf.GetStringWidth(ellipsis)
	if avail <= 0 {
		return ellipsis
	}

	// ตัดทีละ rune (รองรับไทย/พม่า)
	rs := []rune(s)
	for len(rs) > 0 && pdf.GetStringWidth(string(rs)) > avail {
		rs = rs[:len(rs)-1]
	}
	return string(rs) + ellipsis
}

func linesForWidth(pdf *gofpdf.Fpdf, text string, width float64) int {
	// SplitLines ต้องใช้ []byte
	lines := pdf.SplitLines([]byte(text), width)
	if len(lines) == 0 {
		return 1
	}
	return len(lines)
}
