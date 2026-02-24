package report

import (
	"bytes"
	"fmt"
	"math"

	"github.com/jung-kurt/gofpdf"
)

func RenderPDF(t Template, data map[string]any) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.SetAutoPageBreak(true, 15)
	pdf.AddPage()

	pdf.AddUTF8Font("TH", "", "assets/fonts/Poppins/Poppins-Regular.ttf")
	if !pdf.Ok() {
		return nil, pdf.Error()
	}
	pdf.AddUTF8Font("TH", "B", "assets/fonts/Poppins/Poppins-Bold.ttf")
	if !pdf.Ok() {
		return nil, pdf.Error()
	}

	// Title
	pdf.SetFont("TH", "B", 16)
	pdf.Cell(0, 10, t.Title)
	pdf.Ln(12)

	pdf.SetFont("TH", "", 12)

	leftW := 45.0 // ความกว้างคอลัมน์ label (mm)
	gap := 4.0
	pageW, _ := pdf.GetPageSize()
	lm, _, rm, _ := pdf.GetMargins()
	rightW := pageW - lm - rm - leftW - gap

	lineH := 7.0

	drawKV := func(label, value string) {
		x := pdf.GetX()
		y := pdf.GetY()

		lines := pdf.SplitLines([]byte(value), rightW)
		rowH := math.Max(lineH, float64(len(lines))*lineH)

		_, pageH := pdf.GetPageSize()
		_, _, _, bm := pdf.GetMargins()
		if y+rowH > pageH-bm {
			pdf.AddPage()
			x = pdf.GetX()
			y = pdf.GetY()
		}

		pdf.SetXY(x, y)
		pdf.SetFont("TH", "B", 12)
		pdf.MultiCell(leftW, lineH, label, "", "L", false)

		pdf.SetXY(x+leftW+gap, y)
		pdf.SetFont("TH", "", 12)
		pdf.MultiCell(rightW, lineH, value, "", "L", false)

		pdf.SetXY(x, y+rowH)
		pdf.Ln(1) // เว้นระยะระหว่างแถวเล็กน้อย
	}

	for _, s := range t.Sections {
		switch s.Type {
		case "kv":
			val := Getbypath(data, s.Key)
			drawKV(s.Label, fmt.Sprintf("%v", val))
		default:
			drawKV("Unsupported", fmt.Sprintf("section type: %s", s.Type))
		}
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
