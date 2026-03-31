package payslip

import (
	"bytes"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"report-service-m/report/fontmanager"
	"strconv"
	"strings"
	"time"

	"github.com/jung-kurt/gofpdf"
)

var imageHTTPClient = &http.Client{Timeout: 10 * time.Second}

func Render(data Payslip, orientationStr string) ([]byte, error) {
	orientation := normalizeOrientation(orientationStr)

	pdf := gofpdf.New(orientation, "mm", "A4", "")
	pdf.SetMargins(5, 5, 5)
	pdf.SetAutoPageBreak(false, 5)

	if err := fontmanager.LoadAll(pdf); err != nil {
		return nil, err
	}

	pdf.AddPage()

	pageW, _ := pdf.GetPageSize()
	left, _, right, _ := pdf.GetMargins()

	x := left
	y := 5.0
	w := pageW - left - right
	// h := 200.0

	pdf.SetDrawColor(60, 60, 60)
	pdf.SetLineWidth(0.35)
	// pdf.Rect(x, y, w, h, "D")

	headerH := 42.0
	infoH := 80.0
	compensationH := 78.0

	drawHeader(pdf, data.Report, x, y, w, headerH)
	drawEmployeeSection(pdf, data.Report, x, y+headerH, w, infoH)
	drawCompensationSection(pdf, data.Report, x, y+headerH+infoH, w, compensationH)

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func drawHeader(pdf *gofpdf.Fpdf, report Report, x, y, w, h float64) {
	pdf.Line(x, y+h, x+w, y+h)

	logoX := x + 4
	logoY := y + 5
	logoW := 64.0
	logoH := 24.0
	drawLogo(pdf, report.Company.Logo, report.Company.Name, logoX, logoY, logoW, logoH)

	fontmanager.Set(pdf, "th", "B", 10)
	pdf.SetXY(x+4, y+31)
	pdf.CellFormat(120, 5, ellipsisToWidth(pdf, defaultString(report.Company.Name, "Company Name"), 120), "", 0, "L", false, 0, "")

	titleX := x + w - 66
	titleY := y + 6
	titleW := 60.0
	titleH := 10.0

	pdf.Rect(titleX, titleY, titleW, titleH, "D")
	fontmanager.Set(pdf, "th", "BI", 11)
	pdf.SetXY(titleX, titleY+1)
	pdf.CellFormat(titleW, 7, defaultString(report.Document.Title, "PAY SLIP"), "", 0, "C", false, 0, "")

	fontmanager.Set(pdf, "th", "B", 8.5)
	pdf.SetXY(titleX, titleY+13)
	pdf.CellFormat(titleW, 5, defaultString(report.Document.ConfidentialTitle, "PRIVATE & CONFIDENTIAL"), "", 0, "C", false, 0, "")

	fontmanager.Set(pdf, "th", "BI", 6.5)
	pdf.SetXY(titleX, titleY+18.5)
	pdf.CellFormat(titleW, 4, defaultString(report.Document.ConfidentialSubTitle, "(TO BE OPENED BY ADDRESSEE ONLY)"), "", 0, "C", false, 0, "")
}

func drawEmployeeSection(pdf *gofpdf.Fpdf, report Report, x, y, w, h float64) {
	pdf.Rect(x, y, w, h, "D")

	headerH := 10.0
	pdf.Line(x, y+headerH, x+w, y+headerH)

	// =========================
	// GRID หลัก
	// =========================
	leftLabelW := 35.0
	leftValueW := 65.0
	rightW := w - leftLabelW - leftValueW

	col1 := x
	col2 := x + leftLabelW
	col3 := x + leftLabelW + leftValueW

	padding := 2.0

	// เส้นหลัก
	pdf.Line(col2, y, col2, y+h)
	pdf.Line(col3, y, col3, y+h)

	// =========================
	// HEADER (ใช้ 4 ช่องเท่านั้น!)
	// =========================
	fontmanager.Set(pdf, "th", "BI", 8.5)

	accumulationLabelW := rightW * 0.6
	rightSplit := col3 + accumulationLabelW

	// Period
	writeBoxText(pdf, col1+padding, y+2, leftLabelW-padding*2, "Period", "L")
	writeBoxText(pdf, col2+padding, y+2, leftValueW-padding*2, report.Payroll.Period, "L")

	// แบ่ง header โดยให้ช่อง Slip No. เริ่มตรงกับคอลัมน์ขวาด้านล่าง
	payDateLabelEnd := col3 + rightW*0.25
	slipValueStart := col3 + rightW*0.85

	// เส้นแบ่ง header
	pdf.Line(payDateLabelEnd, y, payDateLabelEnd, y+headerH)
	pdf.Line(rightSplit, y, rightSplit, y+headerH)
	pdf.Line(slipValueStart, y, slipValueStart, y+headerH)

	// Pay Date
	writeBoxText(pdf, col3+padding, y+2, payDateLabelEnd-col3-padding*2, "Pay Date", "L")
	writeBoxText(pdf, payDateLabelEnd+padding, y+2, rightSplit-payDateLabelEnd-padding*2, report.Payroll.PayDate, "C")

	// Slip No
	writeBoxText(pdf, rightSplit+padding, y+2, slipValueStart-rightSplit-padding*2, "Slip No.", "L")
	writeBoxText(pdf, slipValueStart+padding, y+2, col3+rightW-slipValueStart-padding*2, report.Payroll.SlipNo, "R")

	// =========================
	// BODY (ใช้ 2 ช่อง)
	// =========================

	rows := []LineItem{
		{Label: "Name", Value: report.Employee.Name},
		{Label: "Emp ID.", Value: report.Employee.EmpID},
		{Label: "Joined Date", Value: report.Employee.JoinedDate},
		{Label: "Division", Value: report.Employee.Division},
		{Label: "Department", Value: report.Employee.Department},
		{Label: "Section", Value: report.Employee.Section},
		{Label: "Position", Value: report.Employee.Position},
		{Label: "Tax ID", Value: report.Employee.TaxID},
	}

	startY := y + headerH + 3
	rowGap := 7.8

	// ซ้าย
	for idx, row := range rows {
		lineY := startY + float64(idx)*rowGap

		fontmanager.Set(pdf, "th", "BI", 8.5)
		writeBoxText(pdf, col1+padding, lineY, leftLabelW-padding*2, row.Label, "L")

		fontmanager.Set(pdf, "th", "", 8.5)
		writeBoxText(pdf, col2+padding, lineY, leftValueW-padding*2, row.Value, "L")
	}

	// ขวา (Accumulation)
	labelW := accumulationLabelW
	valueW := rightW - labelW
	pdf.Line(rightSplit, y+headerH, rightSplit, y+h)

	for idx, row := range report.Accumulations {
		lineY := startY + float64(idx)*rowGap
		if lineY > y+h-8 {
			break
		}

		fontmanager.Set(pdf, "th", "BI", 8.5)
		writeBoxText(pdf, col3+padding, lineY, labelW-padding*2, row.Label, "L")

		fontmanager.Set(pdf, "th", "", 8.5)
		writeBoxText(pdf, rightSplit+padding, lineY, valueW-padding*2, row.Value, "R")
	}
}

func drawCompensationSection(pdf *gofpdf.Fpdf, report Report, x, y, w, h float64) {
	pdf.Rect(x, y, w, h, "D")

	earnW := w * 0.36
	deductW := w * 0.29
	workW := w - earnW - deductW

	summaryH := 18.0
	bodyH := h - summaryH

	pdf.Line(x+earnW, y, x+earnW, y+h)
	pdf.Line(x+earnW+deductW, y, x+earnW+deductW, y+h)
	pdf.Line(x, y+bodyH, x+w, y+bodyH)

	maxRows := maxInt(len(report.Earnings), len(report.Deductions), len(report.WorkStats), 1)
	startY := y + 2.2
	rowH := calculateCompensationRowGap(bodyH, maxRows)

	fontmanager.Set(pdf, "th", "BI", 8.8)
	drawLineItems(pdf, x+2, startY, earnW-4, rowH, report.Earnings)
	drawLineItems(pdf, x+earnW+2, startY, deductW-4, rowH, report.Deductions)
	drawLineItems(pdf, x+earnW+deductW+2, startY, workW-4, rowH, report.WorkStats)

	summary := buildCompensationSummary(report)
	workSummary := buildWorkSummary(report)
	summaryY := y + bodyH + 2.0
	lineGap := 6.6

	drawSummaryCell(pdf, x+2, summaryY, earnW-4, "Total Income", summary.TotalIncome)
	drawSummaryCell(pdf, x+earnW+2, summaryY, deductW-4, "Total Deduct", summary.TotalDeduct)
	drawSummaryValue(pdf, x+earnW+deductW+2, summaryY, workW-4, workSummary.OvertimeTotal)

	drawSummaryCell(pdf, x+2, summaryY+lineGap, earnW-4, "Bank Account No.", summary.BankAccountNo)
	drawSummaryCell(pdf, x+earnW+2, summaryY+lineGap, deductW-4, "Net Pay", summary.NetPay)
	drawSummaryValue(pdf, x+earnW+deductW+2, summaryY+lineGap, workW-4, workSummary.AbsentValue)
}

func drawLogo(pdf *gofpdf.Fpdf, logoURL, companyName string, x, y, w, h float64) {
	pdf.Rect(x, y, w, h, "D")

	if logoURL != "" && drawLogoImage(pdf, logoURL, x+2, y+2, w-4, h-4) {
		return
	}

	fontmanager.Set(pdf, "th", "B", 11)
	pdf.SetTextColor(0, 89, 154)
	pdf.SetXY(x+2, y+7)
	pdf.CellFormat(w-4, 6, ellipsisToWidth(pdf, defaultString(companyName, "LOGO"), w-4), "", 0, "C", false, 0, "")

	fontmanager.Set(pdf, "th", "", 7)
	pdf.SetTextColor(107, 114, 128)
	pdf.SetXY(x+2, y+14)
	pdf.CellFormat(w-4, 4.5, "Logo URL", "", 0, "C", false, 0, "")
	pdf.SetTextColor(0, 0, 0)
}

func drawLogoImage(pdf *gofpdf.Fpdf, source string, x, y, w, h float64) bool {
	imgBytes, err := loadImageBytes(source)
	if err != nil {
		return false
	}

	_, format, err := image.DecodeConfig(bytes.NewReader(imgBytes))
	if err != nil {
		return false
	}

	options := gofpdf.ImageOptions{
		ImageType: strings.ToUpper(format),
		ReadDpi:   true,
	}

	info := pdf.RegisterImageOptionsReader(source, options, bytes.NewReader(imgBytes))
	if info == nil || !pdf.Ok() {
		return false
	}

	imgW := info.Width()
	imgH := info.Height()
	if imgW <= 0 || imgH <= 0 {
		return false
	}

	scale := math.Min(w/imgW, h/imgH)
	renderW := imgW * scale
	renderH := imgH * scale
	renderX := x + (w-renderW)/2
	renderY := y + (h-renderH)/2

	pdf.ImageOptions(source, renderX, renderY, renderW, renderH, false, options, 0, "")
	return pdf.Ok()
}

func loadImageBytes(source string) ([]byte, error) {
	source = strings.TrimSpace(source)
	if source == "" {
		return nil, os.ErrNotExist
	}

	if isRemoteURL(source) {
		return downloadImage(source)
	}

	return os.ReadFile(filepath.Clean(source))
}

func downloadImage(url string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "report-service-m/1.0")

	client := imageHTTPClient
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, io.ErrUnexpectedEOF
	}

	return io.ReadAll(io.LimitReader(resp.Body, 8<<20))
}

func drawLineItems(pdf *gofpdf.Fpdf, x, y, w, rowH float64, items []LineItem) {
	for idx, item := range items {
		lineY := y + float64(idx)*rowH
		writeLabelValue(pdf, x, lineY+1, w, item.Label, item.Value)
	}
}

func drawSummaryCell(pdf *gofpdf.Fpdf, x, y, w float64, label, value string) {
	labelW := w * 0.56
	valueW := w - labelW

	fontmanager.Set(pdf, "th", "BI", 8.6)
	pdf.SetXY(x, y)
	pdf.CellFormat(labelW, 4.5, ellipsisToWidth(pdf, label, labelW), "", 0, "L", false, 0, "")

	fontmanager.Set(pdf, "th", "", 8.5)
	pdf.SetXY(x+labelW, y)
	pdf.CellFormat(valueW, 4.5, ellipsisToWidth(pdf, value, valueW), "", 0, "R", false, 0, "")
}

func drawSummaryValue(pdf *gofpdf.Fpdf, x, y, w float64, value string) {
	fontmanager.Set(pdf, "th", "", 8.5)
	pdf.SetXY(x, y)
	pdf.CellFormat(w, 4.5, ellipsisToWidth(pdf, value, w), "", 0, "R", false, 0, "")
}

func writeLabelValue(pdf *gofpdf.Fpdf, x, y, w float64, label, value string) {
	labelW := w * 0.62
	valueW := w - labelW

	fontmanager.Set(pdf, "th", "BI", 8.5)
	pdf.SetXY(x, y)
	pdf.CellFormat(labelW, 4.5, ellipsisToWidth(pdf, label, labelW), "", 0, "L", false, 0, "")

	fontmanager.Set(pdf, "th", "", 8.5)
	pdf.SetXY(x+labelW, y)
	pdf.CellFormat(valueW, 4.5, ellipsisToWidth(pdf, value, valueW), "", 0, "R", false, 0, "")
}

func writeBoxText(pdf *gofpdf.Fpdf, x, y, w float64, text, align string) {
	pdf.SetXY(x, y)
	pdf.CellFormat(w, 4.5, ellipsisToWidth(pdf, text, w), "", 0, align, false, 0, "")
}

func normalizeOrientation(o string) string {
	switch strings.ToUpper(strings.TrimSpace(o)) {
	case "P", "PORTRAIT":
		return "P"
	case "L", "LANDSCAPE":
		return "L"
	default:
		return "P"
	}
}

func ellipsisToWidth(pdf *gofpdf.Fpdf, s string, maxW float64) string {
	if s == "" || pdf.GetStringWidth(s) <= maxW {
		return s
	}

	ellipsis := "..."
	available := maxW - pdf.GetStringWidth(ellipsis)
	if available <= 0 {
		return ellipsis
	}

	runes := []rune(s)
	for len(runes) > 0 && pdf.GetStringWidth(string(runes)) > available {
		runes = runes[:len(runes)-1]
	}

	return string(runes) + ellipsis
}

func defaultString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func isRemoteURL(value string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	return strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://")
}

type compensationSummary struct {
	TotalIncome   string
	TotalDeduct   string
	NetPay        string
	BankAccountNo string
}

type workSummary struct {
	OvertimeTotal string
	AbsentValue   string
}

func buildCompensationSummary(report Report) compensationSummary {
	totalIncome := strings.TrimSpace(report.Totals.TotalIncome)
	if totalIncome == "" {
		if total, ok := sumLineItemValues(report.Earnings); ok {
			totalIncome = formatAmount(total)
		}
	}

	totalDeduct := strings.TrimSpace(report.Totals.TotalDeduct)
	if totalDeduct == "" {
		if total, ok := sumLineItemValues(report.Deductions); ok {
			totalDeduct = formatAmount(total)
		}
	}

	netPay := strings.TrimSpace(report.Totals.NetPay)
	if netPay == "" {
		earn, earnOK := sumLineItemValues(report.Earnings)
		deduct, deductOK := sumLineItemValues(report.Deductions)
		if earnOK && deductOK {
			netPay = formatAmount(earn - deduct)
		}
	}

	return compensationSummary{
		TotalIncome:   totalIncome,
		TotalDeduct:   totalDeduct,
		NetPay:        netPay,
		BankAccountNo: strings.TrimSpace(report.Totals.BankAccountNo),
	}
}

func buildWorkSummary(report Report) workSummary {
	return workSummary{
		OvertimeTotal: sumWorkDurationValues(report.WorkStats, "ot"),
		AbsentValue:   findWorkStatValue(report.WorkStats, "absent"),
	}
}

func calculateCompensationRowGap(bodyH float64, maxRows int) float64 {
	if maxRows <= 1 {
		return 6.8
	}

	topPadding := 2.2
	bottomPadding := 6.0
	textH := 4.5
	usableH := bodyH - topPadding - bottomPadding - textH
	if usableH <= 0 {
		return 6.8
	}

	rowGap := usableH / float64(maxRows-1)
	if rowGap > 6.8 {
		return 6.8
	}
	if rowGap < 5.6 {
		return 5.6
	}

	return rowGap
}

func findWorkStatValue(items []LineItem, label string) string {
	target := strings.TrimSpace(strings.ToLower(label))
	for _, item := range items {
		if strings.TrimSpace(strings.ToLower(item.Label)) == target {
			return strings.TrimSpace(item.Value)
		}
	}

	return ""
}

func sumWorkDurationValues(items []LineItem, labelPrefix string) string {
	totalSeconds := 0
	hasDuration := false
	hasSeconds := false
	prefix := strings.TrimSpace(strings.ToLower(labelPrefix))

	for _, item := range items {
		label := strings.TrimSpace(strings.ToLower(item.Label))
		if !strings.HasPrefix(label, prefix) {
			continue
		}

		durationSeconds, includeSeconds, ok := parseDuration(item.Value)
		if !ok {
			continue
		}

		totalSeconds += durationSeconds
		hasDuration = true
		hasSeconds = hasSeconds || includeSeconds
	}

	if !hasDuration {
		return ""
	}

	return formatDuration(totalSeconds, hasSeconds)
}

func parseDuration(value string) (seconds int, hasSeconds bool, ok bool) {
	parts := strings.Split(strings.TrimSpace(value), ":")
	if len(parts) != 2 && len(parts) != 3 {
		return 0, false, false
	}

	numbers := make([]int, len(parts))
	for idx, part := range parts {
		number, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil {
			return 0, false, false
		}
		numbers[idx] = number
	}

	if len(numbers) == 2 {
		return numbers[0]*3600 + numbers[1]*60, false, true
	}

	return numbers[0]*3600 + numbers[1]*60 + numbers[2], true, true
}

func formatDuration(totalSeconds int, includeSeconds bool) string {
	hours := totalSeconds / 3600
	minutes := (totalSeconds % 3600) / 60
	seconds := totalSeconds % 60

	if includeSeconds {
		return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
	}

	return fmt.Sprintf("%02d:%02d", hours, minutes)
}

func sumLineItemValues(items []LineItem) (float64, bool) {
	total := 0.0
	found := false

	for _, item := range items {
		value, ok := parseAmount(item.Value)
		if !ok {
			continue
		}
		total += value
		found = true
	}

	return total, found
}

func parseAmount(value string) (float64, bool) {
	clean := strings.TrimSpace(value)
	clean = strings.ReplaceAll(clean, ",", "")
	if clean == "" {
		return 0, false
	}

	number, err := strconv.ParseFloat(clean, 64)
	if err != nil {
		return 0, false
	}

	return number, true
}

func formatAmount(value float64) string {
	negative := value < 0
	if negative {
		value = -value
	}

	parts := strings.Split(fmt.Sprintf("%.2f", value), ".")
	intPart := parts[0]

	for i := len(intPart) - 3; i > 0; i -= 3 {
		intPart = intPart[:i] + "," + intPart[i:]
	}

	if negative {
		intPart = "-" + intPart
	}

	return intPart + "." + parts[1]
}

func maxInt(values ...int) int {
	if len(values) == 0 {
		return 0
	}

	maxValue := values[0]
	for _, value := range values[1:] {
		if value > maxValue {
			maxValue = value
		}
	}

	return maxValue
}
