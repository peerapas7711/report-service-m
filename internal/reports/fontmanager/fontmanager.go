package fontmanager

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/jung-kurt/gofpdf"
)

const (
	FontEN = "EN"
	FontTH = "TH"
	FontMM = "MM"
)

// โหลดฟอนต์ทั้งหมด (เรียกครั้งเดียวต่อ pdf)
func LoadAll(pdf *gofpdf.Fpdf) error {
	fonts := []struct {
		name  string
		style string
		path  string
	}{
		// English
		{FontEN, "", "assets/fonts/Poppins/Poppins-Regular.ttf"},
		{FontEN, "B", "assets/fonts/Poppins/Poppins-Bold.ttf"},
		{FontEN, "I", "assets/fonts/Poppins/Poppins-Italic.ttf"},
		{FontEN, "BI", "assets/fonts/Poppins/Poppins-BoldItalic.ttf"},

		// Thai
		{FontTH, "", "assets/fonts/Noto_Sans_Thai/NotoSansThai-Regular.ttf"},
		{FontTH, "B", "assets/fonts/Noto_Sans_Thai/NotoSansThai-Bold.ttf"},
		{FontTH, "I", "assets/fonts/Noto_Sans_Thai/NotoSansThai-Italic.ttf"},
		{FontTH, "BI", "assets/fonts/Noto_Sans_Thai/NotoSansThai-BoldItalic.ttf"},

		// Myanmar (Italic บางทีไม่มี)
		{FontMM, "", "assets/fonts/Noto_Sans_Myanmar/NotoSansMyanmar-Regular.ttf"},
		{FontMM, "B", "assets/fonts/Noto_Sans_Myanmar/NotoSansMyanmar-Bold.ttf"},
		{FontMM, "I", "assets/fonts/Noto_Sans_Myanmar/NotoSansMyanmar-Italic.ttf"},
		{FontMM, "BI", "assets/fonts/Noto_Sans_Myanmar/NotoSansMyanmar-BoldItalic.ttf"},
	}

	for _, f := range fonts {
		fontPath := resolveFontPath(f.path)
		if fontBytes, err := os.ReadFile(fontPath); err == nil {
			pdf.AddUTF8FontFromBytes(f.name, f.style, fontBytes)
			if !pdf.Ok() {
				return pdf.Error()
			}
		}
	}
	return nil
}

func Set(pdf *gofpdf.Fpdf, lang, style string, size float64) {
	var font string
	switch lang {
	case "th":
		font = FontTH
	case "my":
		font = FontMM
	default:
		font = FontEN
	}

	style = normalizeStyle(lang, style)

	if style == "" {
		pdf.SetFont(font, "", size)
		return
	}

	pdf.SetFont(font, style, size)

	// ถ้า set แล้ว error ให้ fallback เป็น regular
	if !pdf.Ok() {
		pdf.SetFont(font, "", size)
	}
}

func resolveFontPath(path string) string {
	if _, err := os.Stat(path); err == nil {
		return path
	}

	if wd, err := os.Getwd(); err == nil {
		if root := findProjectRoot(wd); root != "" {
			candidate := filepath.Join(root, path)
			if _, err := os.Stat(candidate); err == nil {
				return candidate
			}
		}
	}

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		return path
	}

	if !filepath.IsAbs(currentFile) {
		if absCurrent, err := filepath.Abs(currentFile); err == nil {
			currentFile = absCurrent
		}
	}

	if root := findProjectRoot(filepath.Dir(currentFile)); root != "" {
		candidate := filepath.Join(root, path)
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	return path
}

func findProjectRoot(start string) string {
	dir := filepath.Clean(start)

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

func normalizeStyle(lang, style string) string {
	if style == "" {
		return ""
	}

	if lang == "th" || lang == "my" {
		style = strings.ReplaceAll(style, "I", "")
	}

	return style
}
