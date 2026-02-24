package report

import (
	"log"

	"github.com/jung-kurt/gofpdf"
)

func StypePDF() {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.AddPage()

	pdf.SetFont("Arial", "", 11)
	pdf.SetTextColor(15, 23, 42)

	title(pdf, "gofpdf Styles Demo (A4, mm)")
	pdf.Ln(2)

	// ========= 1) Lines =========
	sectionBar(pdf, "1) Lines: width / color / dashed")

	// thin gray line
	pdf.SetDrawColor(148, 163, 184)
	pdf.SetLineWidth(0.2)
	pdf.Line(20, pdf.GetY()+2, 190, pdf.GetY()+2)
	pdf.Text(20, pdf.GetY()+8, "Thin line: width=0.2")

	// thick blue line
	pdf.SetDrawColor(30, 64, 175)
	pdf.SetLineWidth(1.0)
	pdf.Line(20, pdf.GetY()+14, 190, pdf.GetY()+14)
	pdf.Text(20, pdf.GetY()+20, "Thick line: width=1.0, blue")

	// dashed line
	pdf.SetDrawColor(100, 116, 139)
	pdf.SetLineWidth(0.4)
	pdf.SetDashPattern([]float64{3, 2}, 0) // draw 3mm, gap 2mm
	pdf.Line(20, pdf.GetY()+26, 190, pdf.GetY()+26)
	pdf.SetDashPattern([]float64{}, 0)
	pdf.Text(20, pdf.GetY()+32, "Dashed line: pattern=[3,2]")

	pdf.Ln(28)

	// ========= 2) Rect styles =========
	sectionBar(pdf, "2) Rect: D / F / DF")

	// Rect D
	pdf.SetDrawColor(0, 0, 0)
	pdf.SetLineWidth(0.4)
	pdf.Rect(20, pdf.GetY()+2, 50, 22, "D")
	pdf.Text(22, pdf.GetY()+8, "Rect D")
	pdf.Text(22, pdf.GetY()+15, "Border only")

	// Rect F
	pdf.SetFillColor(219, 234, 254)
	pdf.Rect(80, pdf.GetY()+2, 50, 22, "F")
	pdf.Text(82, pdf.GetY()+8, "Rect F")
	pdf.Text(82, pdf.GetY()+15, "Fill only")

	// Rect DF
	pdf.SetDrawColor(148, 163, 184)
	pdf.SetFillColor(241, 245, 249)
	pdf.SetLineWidth(0.3)
	pdf.Rect(140, pdf.GetY()+2, 50, 22, "DF")
	pdf.Text(142, pdf.GetY()+8, "Rect DF")
	pdf.Text(142, pdf.GetY()+15, "Border+Fill")

	pdf.Ln(28)

	// ========= 3) RoundedRect =========
	sectionBar(pdf, "3) RoundedRect: corners + style")

	y0 := pdf.GetY() + 2
	pdf.SetDrawColor(30, 64, 175)
	pdf.SetFillColor(224, 231, 255)
	pdf.SetLineWidth(0.6)

	// all corners
	pdf.RoundedRect(20, y0, 55, 24, 4, "1234", "DF")
	pdf.Text(22, y0+8, "all corners")
	pdf.Text(22, y0+15, `corners="1234"`)

	// top corners only
	pdf.SetDrawColor(16, 185, 129)
	pdf.SetFillColor(220, 252, 231)
	pdf.RoundedRect(82, y0, 55, 24, 4, "12", "DF")
	pdf.Text(84, y0+8, "top only")
	pdf.Text(84, y0+15, `corners="12"`)

	// bottom corners only
	pdf.SetDrawColor(244, 63, 94)
	pdf.SetFillColor(255, 228, 230)
	pdf.RoundedRect(144, y0, 55, 24, 4, "34", "DF")
	pdf.Text(146, y0+8, "bottom only")
	pdf.Text(146, y0+15, `corners="34"`)

	pdf.Ln(32)

	// ========= 4) Typography =========
	sectionBar(pdf, "4) Typography: bold / color / center")

	pdf.SetFont("Arial", "B", 14)
	pdf.SetTextColor(30, 64, 175)
	pdf.CellFormat(0, 8, "Centered Title (CellFormat align=C)", "", 1, "C", false, 0, "")

	pdf.SetFont("Arial", "", 11)
	pdf.SetTextColor(15, 23, 42)
	pdf.MultiCell(0, 6,
		"MultiCell wraps long text automatically. "+
			"Use it for paragraphs, notes, and long labels. "+
			"This is essential for reports.",
		"", "L", false)

	pdf.Ln(2)

	// ========= 5) Header bar =========
	sectionBar(pdf, "5) Header Bar (like section title)")

	x := 15.0
	y := pdf.GetY() + 2
	w := 180.0
	h := 9.0

	pdf.SetDrawColor(226, 232, 240)
	pdf.SetFillColor(241, 245, 249)
	pdf.Rect(x, y, w, h, "DF")

	pdf.SetFont("Arial", "B", 11)
	pdf.SetTextColor(15, 23, 42)
	pdf.SetXY(x+3, y+2)
	pdf.CellFormat(w-6, 5, "Section: Companies", "", 1, "L", false, 0, "")
	pdf.Ln(6)

	// ========= 6) Card style =========
	sectionBar(pdf, "6) Card style (RoundedRect + padding)")

	y = pdf.GetY() + 2
	pdf.SetDrawColor(226, 232, 240)
	pdf.SetFillColor(248, 250, 252)
	pdf.RoundedRect(20, y, 170, 28, 3, "1234", "DF")

	pdf.SetFont("Arial", "B", 12)
	pdf.SetXY(24, y+5)
	pdf.Cell(0, 6, "Summary Card")

	pdf.SetFont("Arial", "", 10)
	pdf.SetXY(24, y+13)
	pdf.MultiCell(162, 5, "This looks like a modern UI card. Useful for summary blocks.", "", "L", false)
	pdf.Ln(18)

	// ========= 7) Table + zebra =========
	sectionBar(pdf, "7) Table header + zebra rows")

	tableX := 20.0
	tableY := pdf.GetY() + 2
	col1 := 20.0
	col2 := 130.0
	col3 := 10.0
	rowH := 7.0

	// header
	pdf.SetXY(tableX, tableY)
	pdf.SetFont("Arial", "B", 10)
	pdf.SetFillColor(226, 232, 240)
	pdf.SetDrawColor(148, 163, 184)
	pdf.CellFormat(col1, rowH, "Code", "1", 0, "C", true, 0, "")
	pdf.CellFormat(col2, rowH, "Name", "1", 0, "L", true, 0, "")
	pdf.CellFormat(col3, rowH, "", "1", 1, "C", true, 0, "")

	// rows
	pdf.SetFont("Arial", "", 10)
	for i := 1; i <= 6; i++ {
		fill := i%2 == 0
		if fill {
			pdf.SetFillColor(248, 250, 252)
		} else {
			pdf.SetFillColor(255, 255, 255)
		}
		pdf.SetX(tableX)
		pdf.CellFormat(col1, rowH, itoa(i), "1", 0, "C", fill, 0, "")
		pdf.CellFormat(col2, rowH, "Company Name "+itoa(i), "1", 0, "L", fill, 0, "")
		// checkbox cell (border)
		pdf.CellFormat(col3, rowH, "", "1", 1, "C", fill, 0, "")
		// checkbox draw (centered)
		cbx := tableX + col1 + col2 + 2.5
		cby := pdf.GetY() - rowH + 1.2
		drawCheckbox(pdf, cbx, cby, 4.5, i%3 != 0)
	}
	pdf.Ln(6)

	// ========= 8) Checkbox & Radio =========
	sectionBar(pdf, "8) Checkbox + Radio")

	y = pdf.GetY() + 2
	drawCheckbox(pdf, 20, y, 6, false)
	pdf.Text(28, y+4.5, "Checkbox unchecked")

	drawCheckbox(pdf, 85, y, 6, true)
	pdf.Text(93, y+4.5, "Checkbox checked")

	drawRadio(pdf, 150, y+3, 3, false)
	pdf.Text(156, y+4.5, "Radio off")

	drawRadio(pdf, 150, y+14, 3, true)
	pdf.Text(156, y+15.5, "Radio on")

	pdf.Ln(22)

	// ========= 9) Watermark =========
	sectionBar(pdf, "9) Watermark (rotate + light color)")

	// faint watermark across page
	pdf.SetFont("Arial", "B", 40)
	pdf.SetTextColor(226, 232, 240)

	pdf.TransformBegin()
	pdf.TransformRotate(30, 105, 148)
	pdf.Text(30, 160, "DEMO")
	pdf.TransformEnd()

	// reset for normal text
	pdf.SetTextColor(15, 23, 42)
	pdf.SetFont("Arial", "", 10)
	pdf.MultiCell(0, 5, "Watermark is done using TransformRotate + light TextColor.", "", "L", false)

	if err := pdf.OutputFileAndClose("styles_demo.pdf"); err != nil {
		log.Fatal(err)
	}
}

// ---------- helpers ----------

func title(pdf *gofpdf.Fpdf, s string) {
	pdf.SetFont("Arial", "B", 16)
	pdf.CellFormat(0, 10, s, "", 1, "C", false, 0, "")
	pdf.SetFont("Arial", "", 11)
}

func sectionBar(pdf *gofpdf.Fpdf, s string) {
	pdf.Ln(6)
	x := 15.0
	y := pdf.GetY()
	w := 180.0
	h := 8.0
	pdf.SetDrawColor(226, 232, 240)
	pdf.SetFillColor(241, 245, 249)
	pdf.Rect(x, y, w, h, "DF")
	pdf.SetFont("Arial", "B", 11)
	pdf.SetTextColor(15, 23, 42)
	pdf.SetXY(x+3, y+2)
	pdf.CellFormat(w-6, 5, s, "", 1, "L", false, 0, "")
	pdf.SetFont("Arial", "", 10)
	pdf.SetY(y + h)
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

func drawRadio(pdf *gofpdf.Fpdf, cx, cy, r float64, selected bool) {
	pdf.SetDrawColor(15, 23, 42)
	pdf.SetLineWidth(0.2)
	pdf.Circle(cx, cy, r, "D")
	if selected {
		pdf.SetFillColor(15, 23, 42)
		pdf.Circle(cx, cy, r-1.2, "F")
	}
}

func itoa(n int) string {
	// ไม่ใช้ strconv เพื่อให้ไฟล์เดโมสั้น
	if n == 0 {
		return "0"
	}
	sign := ""
	if n < 0 {
		sign = "-"
		n = -n
	}
	buf := make([]byte, 0, 12)
	for n > 0 {
		d := n % 10
		buf = append([]byte{byte('0' + d)}, buf...)
		n /= 10
	}
	return sign + string(buf)
}
