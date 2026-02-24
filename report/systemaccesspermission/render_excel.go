package systemaccesspermission

import (
	"fmt"

	"github.com/xuri/excelize/v2"
)

func SystemAccessPermissionXLSX(data SystemAccess, lang string) ([]byte, error) {
	f := excelize.NewFile()
	sheet := f.GetSheetName(0)

	// --- helpers ---
	title := pickTitle(data.Report.Title, lang)
	company := data.Report.Company.Code + " - " + pickName(data.Report.Company.Name, lang)
	group := pickName(data.Report.Group.Name, lang)

	hCode := map[string]string{"th": "รหัสเมนู", "en": "Menu Code", "my": "မီနူးကုဒ်"}[lang]
	if hCode == "" {
		hCode = "Menu Code"
	}
	hName := map[string]string{"th": "ชื่อเมนู", "en": "Menu Name", "my": "မီနူးအမည်"}[lang]
	if hName == "" {
		hName = "Menu Name"
	}

	// --- column widths (ปรับได้ตามใจ) ---
	_ = f.SetColWidth(sheet, "A", "A", 18)
	_ = f.SetColWidth(sheet, "B", "B", 45)
	_ = f.SetColWidth(sheet, "C", "C", 10)

	// --- styles ---
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 16},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	underlineStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Underline: "single", Size: 11},
	})
	sectionStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Underline: "single", Size: 10},
	})

	borderAll := []excelize.Border{
		{Type: "left", Color: "000000", Style: 1},
		{Type: "right", Color: "000000", Style: 1},
		{Type: "top", Color: "000000", Style: 1},
		{Type: "bottom", Color: "000000", Style: 1},
	}
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 11},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border:    borderAll,
	})
	cellStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 11},
		Alignment: &excelize.Alignment{Vertical: "center"},
		Border:    borderAll,
	})
	cellCenterStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 11},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border:    borderAll,
	})

	// --- layout ---
	// Title (A1:C1)
	_ = f.MergeCell(sheet, "A1", "C1")
	_ = f.SetCellValue(sheet, "A1", title)
	_ = f.SetCellStyle(sheet, "A1", "A1", titleStyle)
	_ = f.SetRowHeight(sheet, 1, 24)

	// Company (A3:C3)
	_ = f.MergeCell(sheet, "A3", "C3")
	_ = f.SetCellValue(sheet, "A3", company)
	_ = f.SetCellStyle(sheet, "A3", "A3", underlineStyle)

	// Group (A5:C5)
	_ = f.MergeCell(sheet, "A5", "C5")
	_ = f.SetCellValue(sheet, "A5", group)
	_ = f.SetCellStyle(sheet, "A5", "A5", sectionStyle)

	// Table header row = 7
	_ = f.SetCellValue(sheet, "A7", hCode)
	_ = f.SetCellValue(sheet, "B7", hName)
	_ = f.SetCellValue(sheet, "C7", "View")
	_ = f.SetCellStyle(sheet, "A7", "C7", headerStyle)
	_ = f.SetRowHeight(sheet, 7, 20)

	// Data rows start = 8
	startRow := 8
	for i, m := range data.Report.Menus {
		r := startRow + i
		a := fmt.Sprintf("A%d", r)
		b := fmt.Sprintf("B%d", r)
		c := fmt.Sprintf("C%d", r)

		_ = f.SetCellValue(sheet, a, m.Code)
		_ = f.SetCellValue(sheet, b, pickName(m.Name, lang))
		if m.View {
			_ = f.SetCellValue(sheet, c, "☑")
		} else {
			_ = f.SetCellValue(sheet, c, "☐")
		}

		_ = f.SetCellStyle(sheet, a, a, cellStyle)
		_ = f.SetCellStyle(sheet, b, b, cellStyle)
		_ = f.SetCellStyle(sheet, c, c, cellCenterStyle)
		_ = f.SetRowHeight(sheet, r, 20)
	}

	// (Optional) ซ่อน gridlines ให้ดูเหมือน report
	show := false
	_ = f.SetSheetView(sheet, 0, &excelize.ViewOptions{ShowGridLines: &show})

	// output bytes (เหมือน pdf.Output(&buf))
	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
