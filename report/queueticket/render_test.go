package queueticket

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestRenderReturnsPDFForSevenHundredTickets(t *testing.T) {
	pdfBytes, err := Render(Options{
		Title: "บัตรคิว",
		Total: 700,
		Now:   time.Date(2026, time.April, 2, 10, 30, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("render queue tickets: %v", err)
	}

	if len(pdfBytes) == 0 {
		t.Fatal("render returned empty pdf")
	}

	if !bytes.HasPrefix(pdfBytes, []byte("%PDF")) {
		t.Fatal("render output is not a pdf")
	}
}

func TestNormalizeOptionsAppliesDefaults(t *testing.T) {
	opts := normalizeOptions(Options{})

	if opts.Title != defaultTitle {
		t.Fatalf("unexpected default title: %s", opts.Title)
	}
	if opts.Total != defaultTotal {
		t.Fatalf("unexpected default total: %d", opts.Total)
	}
	if opts.Start != defaultStart {
		t.Fatalf("unexpected default start: %d", opts.Start)
	}
	if opts.Digits != defaultDigits {
		t.Fatalf("unexpected default digits: %d", opts.Digits)
	}
}

func TestFormatQueueNumber(t *testing.T) {
	got := formatQueueNumber(7, "A", 3)
	if got != "A007" {
		t.Fatalf("unexpected formatted queue number: %s", got)
	}
}

func TestNormalizeOptionsExpandsDigitsForLargeNumbers(t *testing.T) {
	opts := normalizeOptions(Options{
		Start: 9950,
		Total: 120,
	})

	if opts.Digits != 5 {
		t.Fatalf("unexpected digits: %d", opts.Digits)
	}
}

func TestEllipsisToWidthShortensLongText(t *testing.T) {
	pdfBytes, err := Render(Options{
		Title:    strings.Repeat("บัตรคิวสำหรับผู้ติดต่อ ", 10),
		Subtitle: strings.Repeat("Queue Ticket ", 8),
		Total:    1,
	})
	if err != nil {
		t.Fatalf("render with long title: %v", err)
	}

	if !bytes.HasPrefix(pdfBytes, []byte("%PDF")) {
		t.Fatal("render output is not a pdf")
	}
}
