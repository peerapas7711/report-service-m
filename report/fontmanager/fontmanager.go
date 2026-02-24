package fontmanager

import (
	"os"

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
		if _, err := os.Stat(f.path); err == nil {
			pdf.AddUTF8Font(f.name, f.style, f.path)
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

	// fallback style ถ้าไม่มี (กัน Myanmar Italic พัง)
	if style != "" {
		if !pdf.Ok() {
			style = ""
		}
	}

	pdf.SetFont(font, style, size)

	// ถ้า set แล้ว error ให้ fallback เป็น regular
	if !pdf.Ok() {
		pdf.SetFont(font, "", size)
	}
}
