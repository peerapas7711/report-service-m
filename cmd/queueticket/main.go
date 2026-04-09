package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"report-service-m/report/queueticket"
)

func main() {
	var (
		title      = flag.String("title", "บัตรคิว", "title shown on each ticket")
		subtitle   = flag.String("subtitle", "Queue Ticket", "subtitle shown on each ticket")
		queueLabel = flag.String("label", "หมายเลขคิว", "label shown above the queue number")
		prefix     = flag.String("prefix", "", "optional prefix before queue numbers")
		lang       = flag.String("lang", "th", "font language: th, en, my")
		start      = flag.Int("start", 1, "starting queue number")
		total      = flag.Int("total", 700, "number of tickets to generate")
		digits     = flag.Int("digits", 0, "zero-pad digits; 0 means automatic")
		output     = flag.String("output", "queue_tickets_700.pdf", "output pdf file path")
	)

	flag.Parse()

	pdfBytes, err := queueticket.Render(queueticket.Options{
		Title:      *title,
		Subtitle:   *subtitle,
		QueueLabel: *queueLabel,
		Prefix:     *prefix,
		Lang:       *lang,
		Start:      *start,
		Total:      *total,
		Digits:     *digits,
		Now:        time.Now(),
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "render queue tickets: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(*output, pdfBytes, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "write output file: %v\n", err)
		os.Exit(1)
	}

	absPath, err := filepath.Abs(*output)
	if err != nil {
		absPath = *output
	}

	fmt.Printf("generated %d queue tickets at %s\n", *total, absPath)
}
