package payslip_html

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

func GeneratePDF(html string) ([]byte, error) {
	if strings.TrimSpace(html) == "" {
		return nil, fmt.Errorf("html is empty")
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

	ctx, cancelTimeout := context.WithTimeout(ctx, 30*time.Second)
	defer cancelTimeout()

	var pdf []byte
	if err := chromedp.Run(ctx,
		chromedp.Navigate(htmlDataURL(html)),
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

func htmlDataURL(html string) string {
	encoded := base64.StdEncoding.EncodeToString([]byte(html))
	return "data:text/html;charset=utf-8;base64," + url.PathEscape(encoded)
}
