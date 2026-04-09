package queueticket

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"

	"report-service-m/report/fontmanager"

	"github.com/jung-kurt/gofpdf"
)

const (
	defaultLang       = "th"
	defaultTitle      = "บัตรคิว"
	defaultQueueLabel = "หมายเลขคิว"
	defaultSubtitle   = "Queue Ticket"
	defaultTotal      = 700
	defaultStart      = 1
	defaultDigits     = 3
	ticketsPerPage    = 8
	ticketColumns     = 2
	ticketRows        = 4
)

type Options struct {
	Title      string
	Subtitle   string
	QueueLabel string
	Prefix     string
	Lang       string
	Start      int
	Total      int
	Digits     int
	Now        time.Time
}

func Render(opts Options) ([]byte, error) {
	opts = normalizeOptions(opts)

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(8, 8, 8)
	pdf.SetAutoPageBreak(false, 0)
	pdf.AliasNbPages("")

	if err := fontmanager.LoadAll(pdf); err != nil {
		return nil, err
	}

	configureFooter(pdf, opts)
	drawTickets(pdf, opts)

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func normalizeOptions(opts Options) Options {
	if strings.TrimSpace(opts.Title) == "" {
		opts.Title = defaultTitle
	}
	if strings.TrimSpace(opts.Subtitle) == "" {
		opts.Subtitle = defaultSubtitle
	}
	if strings.TrimSpace(opts.QueueLabel) == "" {
		opts.QueueLabel = defaultQueueLabel
	}
	if strings.TrimSpace(opts.Lang) == "" {
		opts.Lang = defaultLang
	}
	if opts.Start <= 0 {
		opts.Start = defaultStart
	}
	if opts.Total <= 0 {
		opts.Total = defaultTotal
	}
	if opts.Now.IsZero() {
		opts.Now = time.Now()
	}

	maxNumber := opts.Start + opts.Total - 1
	autoDigits := len(strconv.Itoa(maxNumber))
	if autoDigits < defaultDigits {
		autoDigits = defaultDigits
	}
	if opts.Digits < autoDigits {
		opts.Digits = autoDigits
	}

	opts.Title = strings.TrimSpace(opts.Title)
	opts.Subtitle = strings.TrimSpace(opts.Subtitle)
	opts.QueueLabel = strings.TrimSpace(opts.QueueLabel)
	opts.Prefix = strings.TrimSpace(opts.Prefix)
	opts.Lang = strings.TrimSpace(opts.Lang)

	return opts
}

func configureFooter(pdf *gofpdf.Fpdf, opts Options) {
	// generatedAt := opts.Now.Format("02/01/2006 15:04")

	// pdf.SetFooterFunc(func() {
	// 	pdf.SetY(-7)
	// 	fontmanager.Set(pdf, opts.Lang, "", 7)
	// 	pdf.SetTextColor(90, 96, 110)

	// 	pageW, _ := pdf.GetPageSize()
	// 	left, _, right, _ := pdf.GetMargins()
	// 	contentW := pageW - left - right

	// 	footer := fmt.Sprintf("พิมพ์ %s    หน้า %d/{nb}", generatedAt, pdf.PageNo())
	// 	pdf.SetX(left)
	// 	pdf.CellFormat(contentW, 4, footer, "", 0, "R", false, 0, "")

	// 	pdf.SetTextColor(0, 0, 0)
	// })
}

func drawTickets(pdf *gofpdf.Fpdf, opts Options) {
	pageW, pageH := pdf.GetPageSize()
	left, top, right, bottom := pdf.GetMargins()
	contentW := pageW - left - right
	contentH := pageH - top - bottom - 6

	gapX := 5.0
	gapY := 5.0
	ticketW := (contentW - gapX*float64(ticketColumns-1)) / float64(ticketColumns)
	ticketH := (contentH - gapY*float64(ticketRows-1)) / float64(ticketRows)

	for i := 0; i < opts.Total; i++ {
		if i%ticketsPerPage == 0 {
			pdf.AddPage()
		}

		slot := i % ticketsPerPage
		col := slot % ticketColumns
		row := slot / ticketColumns

		x := left + float64(col)*(ticketW+gapX)
		y := top + float64(row)*(ticketH+gapY)
		queueNumber := opts.Start + i

		drawTicket(pdf, opts, i+1, queueNumber, x, y, ticketW, ticketH)
	}
}

func drawTicket(pdf *gofpdf.Fpdf, opts Options, index, queueNumber int, x, y, w, h float64) {
	accentR, accentG, accentB := 12, 115, 125
	headerH := 13.0
	bodyInset := 4.0

	pdf.SetDrawColor(12, 34, 56)
	pdf.SetLineWidth(0.45)
	pdf.Rect(x, y, w, h, "D")

	pdf.SetFillColor(accentR, accentG, accentB)
	pdf.Rect(x, y, w, headerH, "F")

	fontmanager.Set(pdf, opts.Lang, "B", 15)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetXY(x+bodyInset, y+2.5)
	pdf.CellFormat(w-bodyInset*2, 7, ellipsisToWidth(pdf, opts.Title, w-bodyInset*2), "", 0, "C", false, 0, "")

	// fontmanager.Set(pdf, opts.Lang, "", 8.5)
	// pdf.SetTextColor(85, 98, 114)
	// pdf.SetXY(x+bodyInset, y+headerH+3)
	// pdf.CellFormat(w-bodyInset*2, 4, ellipsisToWidth(pdf, opts.Subtitle, w-bodyInset*2), "", 0, "C", false, 0, "")

	fontmanager.Set(pdf, opts.Lang, "B", 10)
	pdf.SetTextColor(12, 34, 56)
	pdf.SetXY(x+bodyInset, y+headerH+11)
	pdf.CellFormat(w-bodyInset*2, 5, opts.QueueLabel, "", 0, "C", false, 0, "")

	queueText := formatQueueNumber(queueNumber, opts.Prefix, opts.Digits)
	fontmanager.Set(pdf, "en", "B", 34)
	pdf.SetTextColor(accentR, accentG, accentB)
	pdf.SetXY(x+bodyInset, y+h*0.42)
	pdf.CellFormat(w-bodyInset*2, 16, queueText, "", 0, "C", false, 0, "")

	pdf.SetDrawColor(202, 212, 224)
	pdf.SetLineWidth(0.2)
	lineY := y + h - 19
	pdf.Line(x+bodyInset, lineY, x+w-bodyInset, lineY)

	// fontmanager.Set(pdf, opts.Lang, "", 8)
	// pdf.SetTextColor(85, 98, 114)
	// pdf.SetXY(x+bodyInset, lineY+2)
	// pdf.CellFormat(w-bodyInset*2, 4, fmt.Sprintf("ลำดับ %d จาก %d", index, opts.Total), "", 0, "L", false, 0, "")

	// fontmanager.Set(pdf, "en", "", 8)
	// pdf.SetXY(x+bodyInset, lineY+2)
	// pdf.CellFormat(w-bodyInset*2, 4, opts.Now.Format("02 Jan 2006 15:04"), "", 0, "R", false, 0, "")

	// pdf.SetTextColor(0, 0, 0)
}

func formatQueueNumber(number int, prefix string, digits int) string {
	formatted := fmt.Sprintf("%0*d", digits, number)
	if prefix == "" {
		return formatted
	}

	return prefix + formatted
}

func ellipsisToWidth(pdf *gofpdf.Fpdf, s string, maxW float64) string {
	if pdf.GetStringWidth(s) <= maxW {
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
