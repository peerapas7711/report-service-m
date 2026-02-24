package permission

import (
	"fmt"
	"math"

	"github.com/jung-kurt/gofpdf"
)

func drawHeader(pdf *gofpdf.Fpdf, data Report) {
	pdf.SetFont("TH", "B", 16)
	pdf.CellFormat(0, 10, data.Title, "", 0, "C", false, 0, "")
	pdf.Ln(12)

	// เส้นคั่น
	x1, y := pdf.GetX(), pdf.GetY()
	pageW, _ := pdf.GetPageSize()
	lm, _, rm, _ := pdf.GetMargins()
	pdf.SetLineWidth(0.3)
	pdf.Line(lm, y, pageW-rm, y)
	pdf.Ln(6)

	// ข้อมูลหัวกระดาษ
	pdf.SetFont("TH", "", 12)
	pdf.CellFormat(0, 7, "กลุ่มผู้ใช้: "+data.Group, "", 1, "L", false, 0, "")
	pdf.CellFormat(0, 7, "ชื่อผู้ใช้: "+data.User.Username, "", 1, "L", false, 0, "")
	pdf.Ln(4)

	_ = x1
}

func ensureSpace(pdf *gofpdf.Fpdf, needH float64) {
	_, pageH := pdf.GetPageSize()
	_, _, _, bm := pdf.GetMargins()
	bottom := pageH - bm
	if pdf.GetY()+needH > bottom {
		pdf.AddPage()
	}
}

func drawBox(pdf *gofpdf.Fpdf, x, y, w, h float64) {
	pdf.SetLineWidth(0.2)
	pdf.Rect(x, y, w, h, "D")
}

func drawCheckbox(pdf *gofpdf.Fpdf, x, y, size float64, checked bool) {
	pdf.Rect(x, y, size, size, "D")
	if checked {
		pdf.SetLineWidth(0.6)
		pdf.Line(x+size*0.2, y+size*0.55, x+size*0.45, y+size*0.78)
		pdf.Line(x+size*0.45, y+size*0.78, x+size*0.82, y+size*0.22)
		pdf.SetLineWidth(0.2)
	}
}

// -------------------- Checklist Box --------------------

func drawChecklistBox(pdf *gofpdf.Fpdf, title string, items []CheckItem) {
	pageW, _ := pdf.GetPageSize()
	lm, _, rm, _ := pdf.GetMargins()

	boxW := pageW - lm - rm
	padding := 3.0
	lineH := 6.5
	cb := 5.0
	gap := 2.5

	// เราจะวาดเป็น "หลายกล่องย่อย" เมื่อขึ้นหน้าใหม่ เพื่อไม่ต้องคำนวณความสูงทั้งหมดล่วงหน้า
	startNewBox := func() (x, y float64) {
		ensureSpace(pdf, 18) // พื้นที่ขั้นต่ำสำหรับหัวกล่อง
		x = lm
		y = pdf.GetY()

		// วาดกรอบกล่องแบบ “ชั่วคราว” (เราจะไม่รู้ h สุดท้าย) → เทคนิค: วาดเฉพาะหัว + เส้นคั่น และไม่ต้องมีกรอบปิดก็ได้
		// แต่เพื่อให้เหมือนฟอร์ม: เราจะวาดกรอบตอนรู้ yEnd (ทำทีหลังไม่ได้ง่ายใน gofpdf)
		// ดังนั้นใช้วิธี: วาดกรอบเป็น "ต่อชิ้น" คือหัวกล่องเป็นกรอบ และแต่ละแถวไม่จำเป็นต้องมีกรอบย่อย
		drawBox(pdf, x, y, boxW, 10) // หัวกล่องสูง 10
		pdf.SetXY(x+padding, y+2)
		pdf.SetFont("TH", "B", 11)
		pdf.CellFormat(boxW-2*padding, 6, title, "", 0, "L", false, 0, "")
		pdf.SetY(y + 12) // หลังหัวกล่อง
		return x, y
	}

	// เริ่มกล่องแรก
	startNewBox()

	pdf.SetFont("TH", "", 11)

	for _, it := range items {
		// คำนวณความสูงของข้อความด้านซ้าย (MultiCell)
		x := lm
		y := pdf.GetY()

		rightX := lm + boxW - padding - cb
		textW := boxW - (padding * 2) - cb - gap

		lines := pdf.SplitLines([]byte(it.Label), textW)
		rowH := math.Max(lineH, float64(len(lines))*lineH)

		// ถ้าจะล้นหน้า → ขึ้นหน้าใหม่ + เปิดกล่องหัวใหม่
		ensureSpace(pdf, rowH+6)
		if pdf.GetY() != y {
			// หลัง AddPage แล้ว ให้เริ่มกล่องใหม่
			startNewBox()
			x = lm
			y = pdf.GetY()
			rightX = lm + boxW - padding - cb
		}

		// วาดข้อความ (ซ้าย)
		pdf.SetXY(x+padding, y)
		pdf.MultiCell(textW, lineH, it.Label, "", "L", false)

		// วาด checkbox (ขวา) ให้อยู่ระดับบนของแถว
		drawCheckbox(pdf, rightX, y+0.5, cb, it.Checked)

		// เลื่อน Y ลงท้ายแถว
		pdf.SetXY(lm, y+rowH+1)
	}

	pdf.Ln(4)
}

// -------------------- Table Check Box --------------------

func drawTableCheckBox(pdf *gofpdf.Fpdf, title string, nameColLabel string, rows []RowCheck) {
	pageW, _ := pdf.GetPageSize()
	lm, _, rm, _ := pdf.GetMargins()

	boxW := pageW - lm - rm
	padding := 3.0
	rowH := 7.0
	cb := 5.0
	gap := 2.5

	codeW := 18.0
	nameW := boxW - (padding * 2) - codeW - gap - cb

	startBox := func() {
		ensureSpace(pdf, 22)
		x := lm
		y := pdf.GetY()

		// หัวกล่อง
		drawBox(pdf, x, y, boxW, 10)
		pdf.SetXY(x+padding, y+2)
		pdf.SetFont("TH", "B", 11)
		pdf.CellFormat(boxW-2*padding, 6, title, "", 0, "L", false, 0, "")
		pdf.SetY(y + 12)

		// หัวตาราง
		pdf.SetFont("TH", "B", 11)
		pdf.SetX(lm + padding)
		pdf.CellFormat(codeW, rowH, "รหัส", "1", 0, "C", false, 0, "")
		pdf.CellFormat(nameW, rowH, nameColLabel, "1", 0, "L", false, 0, "")
		// ช่อง checkbox header (เว้นเปล่า)
		pdf.CellFormat(cb, rowH, "", "1", 0, "C", false, 0, "")
		pdf.Ln(-1)

		pdf.SetFont("TH", "", 11)
	}

	startBox()

	for _, r := range rows {
		// วัดความสูง name เผื่อ wrap
		lines := pdf.SplitLines([]byte(r.Name), nameW)
		h := math.Max(rowH, float64(len(lines))*rowH)

		// ถ้าจะล้นหน้า ให้ขึ้นหน้าใหม่ และวาดหัวตารางซ้ำ
		ensureSpace(pdf, h+10)
		// ถ้าเพิ่ง AddPage ต้องวาดหัวกล่อง+หัวตารางใหม่
		// (เช็คแบบง่าย: ถ้า Y ใกล้ขอบบนมาก แปลว่าเพิ่งขึ้นหน้า)
		if pdf.GetY() < 30 {
			startBox()
		}

		y := pdf.GetY()
		x := lm + padding

		// code cell
		pdf.SetXY(x, y)
		pdf.CellFormat(codeW, h, r.Code, "1", 0, "C", false, 0, "")

		// name cell (MultiCell)
		pdf.SetXY(x+codeW, y)
		pdf.MultiCell(nameW, rowH, r.Name, "1", "L", false)

		// checkbox cell border
		cbX := x + codeW + nameW
		pdf.SetXY(cbX, y)
		pdf.CellFormat(cb, h, "", "1", 0, "C", false, 0, "")

		// checkbox centered vertically in cell
		drawCheckbox(pdf, cbX+(cb-5)/2, y+(h-5)/2, 5, r.Checked)

		// move to next row
		pdf.SetXY(lm, y+h)
	}

	pdf.Ln(4)
}

// small helper (no strconv to keep draw.go minimal)
func itoa(n int) string { return fmt.Sprintf("%d", n) }
