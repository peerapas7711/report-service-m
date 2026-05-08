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
	return RenderWithOptions(data, RenderOptions{
		Orientation: orientationStr,
	})
}

func RenderWithOptions(data Payslip, opts RenderOptions) ([]byte, error) {
	templateID := strings.TrimSpace(opts.TemplateID)
	if templateID == "" {
		templateID = data.TemplateID
	}

	layoutTemplate, err := loadLayoutTemplate(templateID)
	if err != nil {
		return nil, err
	}

	orientationValue := strings.TrimSpace(opts.Orientation)
	if orientationValue == "" {
		orientationValue = layoutTemplate.Page.DefaultOrientation
	}
	orientation := normalizeOrientation(orientationValue)
	model := buildRenderModel(data, layoutTemplate.ID)

	margin := layoutTemplate.Page.Margin
	if margin <= 0 {
		margin = 5
	}

	pdf := gofpdf.New(orientation, "mm", "A4", "")
	pdf.SetMargins(margin, margin, margin)
	pdf.SetAutoPageBreak(false, margin)

	if err := fontmanager.LoadAll(pdf); err != nil {
		return nil, err
	}

	pdf.AddPage()
	renderLayoutTemplate(pdf, model, layoutTemplate)

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func drawLogo(pdf *gofpdf.Fpdf, logoURL, companyName string, x, y, w, h float64, showBorder bool) {
	if showBorder {
		pdf.Rect(x, y, w, h, "D")
	}

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
