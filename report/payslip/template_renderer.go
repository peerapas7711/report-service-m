package payslip

import (
	"strings"

	"report-service-m/report/fontmanager"

	"github.com/jung-kurt/gofpdf"
)

type canvasRect struct {
	X float64
	Y float64
	W float64
	H float64
}

func renderLayoutTemplate(pdf *gofpdf.Fpdf, model renderModel, tmpl LayoutTemplate) {
	pageW, pageH := pdf.GetPageSize()
	left, top, right, bottom := pdf.GetMargins()

	root := canvasRect{
		X: left,
		Y: top,
		W: pageW - left - right,
		H: pageH - top - bottom,
	}

	renderLayoutElements(pdf, model, root, tmpl.Elements)
}

func renderLayoutElements(pdf *gofpdf.Fpdf, model renderModel, parent canvasRect, elements []LayoutElement) {
	for _, element := range elements {
		renderLayoutElement(pdf, model, parent, element)
	}
}

func renderLayoutElement(pdf *gofpdf.Fpdf, model renderModel, parent canvasRect, element LayoutElement) {
	switch strings.ToLower(strings.TrimSpace(element.Type)) {
	case "group":
		renderLayoutElements(pdf, model, resolveRect(parent, element), element.Children)
	case "rect":
		drawRectElement(pdf, parent, element)
	case "line":
		drawLineElement(pdf, parent, element)
	case "text":
		drawTextElement(pdf, model, parent, element)
	case "logo":
		drawLogoElement(pdf, model, parent, element)
	case "section_list":
		drawSectionListElement(pdf, model, parent, element)
	}
}

func drawRectElement(pdf *gofpdf.Fpdf, parent canvasRect, element LayoutElement) {
	rect := resolveRect(parent, element)

	setDrawColor(pdf, element.DrawColor, 32, 32, 32)
	if element.FillColor != nil {
		pdf.SetFillColor(element.FillColor.R, element.FillColor.G, element.FillColor.B)
	}
	pdf.SetLineWidth(defaultFloat(element.LineWidth, 0.2))

	style := "D"
	if element.FillColor != nil {
		style = "FD"
	}

	if element.Radius > 0 {
		corners := element.Corners
		if corners == "" {
			corners = "1234"
		}
		pdf.RoundedRect(rect.X, rect.Y, rect.W, rect.H, element.Radius, corners, style)
		return
	}

	pdf.Rect(rect.X, rect.Y, rect.W, rect.H, style)
}

func drawLineElement(pdf *gofpdf.Fpdf, parent canvasRect, element LayoutElement) {
	x1, y1 := resolvePoint(parent, element.X, element.Y)
	x2, y2 := resolvePoint(parent, element.X2, element.Y2)

	setDrawColor(pdf, element.DrawColor, 32, 32, 32)
	pdf.SetLineWidth(defaultFloat(element.LineWidth, 0.2))
	pdf.Line(x1, y1, x2, y2)
}

func drawTextElement(pdf *gofpdf.Fpdf, model renderModel, parent canvasRect, element LayoutElement) {
	rect := resolveRect(parent, element)
	text := resolveText(model, element)
	if text == "" {
		return
	}

	padding := defaultFloat(element.Padding, 0)
	x := rect.X + padding
	y := rect.Y + padding
	w := rect.W - padding*2
	h := rect.H - padding*2
	if w <= 0 || h <= 0 {
		return
	}

	applyFontSpec(pdf, element.Font, "th", "", 8.5)
	setTextColor(pdf, element.TextColor)

	align := defaultString(strings.ToUpper(strings.TrimSpace(element.Align)), "L")
	if element.Multiline {
		pdf.SetXY(x, y)
		pdf.MultiCell(w, defaultFloat(element.LineHeight, 4.2), text, "", align, false)
		pdf.SetTextColor(0, 0, 0)
		return
	}

	pdf.SetXY(x, y)
	pdf.CellFormat(w, h, ellipsisToWidth(pdf, text, w), "", 0, align, false, 0, "")
	pdf.SetTextColor(0, 0, 0)
}

func drawLogoElement(pdf *gofpdf.Fpdf, model renderModel, parent canvasRect, element LayoutElement) {
	rect := resolveRect(parent, element)
	source := strings.TrimSpace(element.Source)
	if source == "" {
		source = "company_logo"
	}

	showBorder := true
	if element.ShowBorder != nil {
		showBorder = *element.ShowBorder
	}

	drawLogo(pdf, model.lookupValue(source), model.lookupValue("company_name"), rect.X, rect.Y, rect.W, rect.H, showBorder)
}

func drawSectionListElement(pdf *gofpdf.Fpdf, model renderModel, parent canvasRect, element LayoutElement) {
	items := model.lookupSection(element.Source)
	if len(items) == 0 {
		return
	}

	if element.MaxRows > 0 && len(items) > element.MaxRows {
		items = items[:element.MaxRows]
	}

	rect := resolveRect(parent, element)
	padding := defaultFloat(element.Padding, 0)
	inner := canvasRect{
		X: rect.X + padding,
		Y: rect.Y + padding,
		W: rect.W - padding*2,
		H: rect.H - padding*2,
	}
	if inner.W <= 0 || inner.H <= 0 {
		return
	}

	lineHeight := defaultFloat(element.LineHeight, 4.2)
	labelRatio := element.LabelRatio
	if labelRatio < 0 {
		labelRatio = 0
	}

	valueAlign := defaultString(strings.ToUpper(strings.TrimSpace(element.ValueAlign)), "R")

	if element.Stacked {
		rowGap := element.RowGap
		if rowGap <= 0 {
			rowGap = calculateElementRowGap(inner.H, len(items)*2, lineHeight, defaultFloat(element.MinRowGap, 4.8), defaultFloat(element.MaxRowGap, 6.8))
		}

		for idx, item := range items {
			labelY := inner.Y + float64(idx*2)*rowGap
			if labelY+lineHeight > inner.Y+inner.H {
				break
			}

			if label := strings.TrimSpace(item.Label); label != "" {
				applyFontSpec(pdf, element.LabelFont, "th", "", 8.2)
				pdf.SetXY(inner.X, labelY)
				pdf.CellFormat(inner.W, lineHeight, ellipsisToWidth(pdf, label, inner.W), "", 0, "L", false, 0, "")
			}

			valueY := inner.Y + float64(idx*2+1)*rowGap
			if valueY+lineHeight > inner.Y+inner.H {
				break
			}

			if value := strings.TrimSpace(item.Value); value != "" {
				applyFontSpec(pdf, element.ValueFont, "th", "", 8.2)
				pdf.SetXY(inner.X, valueY)
				pdf.CellFormat(inner.W, lineHeight, ellipsisToWidth(pdf, value, inner.W), "", 0, valueAlign, false, 0, "")
			}
		}
		return
	}

	rowGap := element.RowGap
	if rowGap <= 0 {
		rowGap = calculateElementRowGap(inner.H, len(items), lineHeight, defaultFloat(element.MinRowGap, 4.8), defaultFloat(element.MaxRowGap, 6.8))
	}

	for idx, item := range items {
		lineY := inner.Y + float64(idx)*rowGap
		if lineY+lineHeight > inner.Y+inner.H {
			break
		}

		if labelRatio > 0 && strings.TrimSpace(item.Label) != "" {
			labelW := inner.W * labelRatio
			if labelW < 0 {
				labelW = 0
			}
			if labelW > inner.W {
				labelW = inner.W
			}
			valueW := inner.W - labelW

			if labelW > 0 {
				applyFontSpec(pdf, element.LabelFont, "th", "", 8.2)
				pdf.SetXY(inner.X, lineY)
				pdf.CellFormat(labelW, lineHeight, ellipsisToWidth(pdf, item.Label, labelW), "", 0, "L", false, 0, "")
			}

			if valueW > 0 {
				applyFontSpec(pdf, element.ValueFont, "th", "", 8.2)
				pdf.SetXY(inner.X+labelW, lineY)
				pdf.CellFormat(valueW, lineHeight, ellipsisToWidth(pdf, item.Value, valueW), "", 0, valueAlign, false, 0, "")
			}
			continue
		}

		applyFontSpec(pdf, element.ValueFont, "th", "", 8.2)
		pdf.SetXY(inner.X, lineY)
		pdf.CellFormat(inner.W, lineHeight, ellipsisToWidth(pdf, item.Value, inner.W), "", 0, valueAlign, false, 0, "")
	}
}

func applyFontSpec(pdf *gofpdf.Fpdf, spec *FontSpec, defaultLang, defaultStyle string, defaultSize float64) {
	lang := defaultLang
	style := defaultStyle
	size := defaultSize

	if spec != nil {
		if strings.TrimSpace(spec.Lang) != "" {
			lang = strings.TrimSpace(spec.Lang)
		}
		if strings.TrimSpace(spec.Style) != "" {
			style = strings.TrimSpace(spec.Style)
		}
		if spec.Size > 0 {
			size = spec.Size
		}
	}

	fontmanager.Set(pdf, lang, style, size)
}

func setDrawColor(pdf *gofpdf.Fpdf, color *ColorSpec, fallbackR, fallbackG, fallbackB int) {
	if color == nil {
		pdf.SetDrawColor(fallbackR, fallbackG, fallbackB)
		return
	}

	pdf.SetDrawColor(color.R, color.G, color.B)
}

func setTextColor(pdf *gofpdf.Fpdf, color *ColorSpec) {
	if color == nil {
		pdf.SetTextColor(0, 0, 0)
		return
	}

	pdf.SetTextColor(color.R, color.G, color.B)
}

func resolveRect(parent canvasRect, element LayoutElement) canvasRect {
	return canvasRect{
		X: parent.X + parent.W*element.X,
		Y: parent.Y + parent.H*element.Y,
		W: parent.W * element.W,
		H: parent.H * element.H,
	}
}

func resolvePoint(parent canvasRect, xRatio, yRatio float64) (float64, float64) {
	return parent.X + parent.W*xRatio, parent.Y + parent.H*yRatio
}

func resolveText(model renderModel, element LayoutElement) string {
	if strings.TrimSpace(element.Source) != "" {
		if value := model.lookupValue(element.Source); value != "" {
			return value
		}
	}

	return element.Text
}

func calculateElementRowGap(height float64, rows int, lineHeight, minGap, maxGap float64) float64 {
	if rows <= 1 {
		return maxGap
	}

	rowGap := (height - lineHeight) / float64(rows)
	if rowGap > maxGap {
		return maxGap
	}
	if rowGap < minGap {
		return minGap
	}

	return rowGap
}

func defaultFloat(value, fallback float64) float64 {
	if value <= 0 {
		return fallback
	}
	return value
}

func (m renderModel) lookupValue(key string) string {
	if value, ok := m.Values[strings.TrimSpace(key)]; ok {
		return value
	}
	return ""
}

func (m renderModel) lookupSection(key string) []LineItem {
	if items, ok := m.Sections[strings.TrimSpace(key)]; ok {
		return items
	}
	return nil
}
