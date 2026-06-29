package payslip_html

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

const defaultPDFTimeout = 30 * time.Second

type PDFOptions struct {
	Timeout time.Duration
}

func GeneratePDF(html string) ([]byte, error) {
	return GeneratePDFWithOptions(html, PDFOptions{})
}

func GeneratePDFWithOptions(html string, opts PDFOptions) ([]byte, error) {
	if strings.TrimSpace(html) == "" {
		return nil, fmt.Errorf("html is empty")
	}

	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = defaultPDFTimeout
	}

	allocatorCtx, cancelAllocator := chromedp.NewExecAllocator(
		context.Background(),
		append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.Headless,
			chromedp.DisableGPU,
			chromedp.NoFirstRun,
			chromedp.NoDefaultBrowserCheck,
		)...,
	)
	defer cancelAllocator()

	ctx, cancel := chromedp.NewContext(allocatorCtx)
	defer cancel()

	ctx, cancelTimeout := context.WithTimeout(ctx, timeout)
	defer cancelTimeout()

	var pdf []byte
	if err := chromedp.Run(ctx,
		chromedp.Navigate("about:blank"),
		chromedp.ActionFunc(func(ctx context.Context) error {
			frameTree, err := page.GetFrameTree().Do(ctx)
			if err != nil {
				return err
			}
			return page.SetDocumentContent(frameTree.Frame.ID, html).Do(ctx)
		}),
		chromedp.WaitReady("body", chromedp.ByQuery),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			pdf, _, err = page.PrintToPDF().
				WithPrintBackground(true).
				WithPreferCSSPageSize(true).
				Do(ctx)
			return err
		}),
	); err != nil {
		return nil, fmt.Errorf("generate payslip pdf from html: %w", err)
	}

	return pdf, nil
}
