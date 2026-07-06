package excel

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

type Workbook struct {
	Sheets []Sheet
}

type Sheet struct {
	Name    string
	Headers []string
	Rows    [][]any
}

func Render(workbook Workbook) ([]byte, error) {
	if len(workbook.Sheets) == 0 {
		return nil, errors.New("at least one sheet is required")
	}

	file := excelize.NewFile()
	defer func() {
		_ = file.Close()
	}()

	if err := file.SetDocProps(&excelize.DocProperties{
		Creator:        "report-service-m",
		LastModifiedBy: "report-service-m",
	}); err != nil {
		return nil, err
	}

	sheetNames := uniqueSheetNames(workbook.Sheets)
	for i, sheet := range workbook.Sheets {
		name := sheetNames[i]
		if i == 0 {
			if err := file.SetSheetName("Sheet1", name); err != nil {
				return nil, err
			}
		} else if _, err := file.NewSheet(name); err != nil {
			return nil, err
		}

		if err := writeSheet(file, name, sheet); err != nil {
			return nil, err
		}
	}

	index, err := file.GetSheetIndex(sheetNames[0])
	if err != nil {
		return nil, err
	}
	file.SetActiveSheet(index)

	buffer, err := file.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func writeSheet(file *excelize.File, name string, sheet Sheet) error {
	headers, rows := normalizedGrid(sheet)
	if len(headers) == 0 {
		return errors.New("sheet must have at least one column")
	}

	headerStyle, bodyStyle, err := createStyles(file)
	if err != nil {
		return err
	}

	headerRow := stringSliceToAny(headers)
	if err := file.SetSheetRow(name, "A1", &headerRow); err != nil {
		return err
	}

	for rowIdx, row := range rows {
		cell, err := excelize.CoordinatesToCellName(1, rowIdx+2)
		if err != nil {
			return err
		}
		if err := file.SetSheetRow(name, cell, &row); err != nil {
			return err
		}
	}

	lastCell, err := excelize.CoordinatesToCellName(len(headers), len(rows)+1)
	if err != nil {
		return err
	}
	lastHeaderCell, err := excelize.CoordinatesToCellName(len(headers), 1)
	if err != nil {
		return err
	}

	if err := file.SetCellStyle(name, "A1", lastHeaderCell, headerStyle); err != nil {
		return err
	}
	if len(rows) > 0 {
		if err := file.SetCellStyle(name, "A2", lastCell, bodyStyle); err != nil {
			return err
		}
	}
	if err := file.SetRowHeight(name, 1, 24); err != nil {
		return err
	}
	if err := setColumnWidths(file, name, headers, rows); err != nil {
		return err
	}
	if err := file.AutoFilter(name, "A1:"+lastCell, nil); err != nil {
		return err
	}
	return file.SetPanes(name, &excelize.Panes{
		Freeze:      true,
		YSplit:      1,
		TopLeftCell: "A2",
		ActivePane:  "bottomLeft",
		Selection: []excelize.Selection{
			{SQRef: "A2", ActiveCell: "A2", Pane: "bottomLeft"},
		},
	})
}

func createStyles(file *excelize.File) (int, int, error) {
	borders := []excelize.Border{
		{Type: "left", Color: "CBD5E1", Style: 1},
		{Type: "right", Color: "CBD5E1", Style: 1},
		{Type: "top", Color: "CBD5E1", Style: 1},
		{Type: "bottom", Color: "CBD5E1", Style: 1},
	}

	headerStyle, err := file.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:  true,
			Color: "FFFFFF",
			Size:  11,
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"C7C7C7"},
			Pattern: 1,
		},
		Border: borders,
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
			WrapText:   true,
		},
	})
	if err != nil {
		return 0, 0, err
	}

	bodyStyle, err := file.NewStyle(&excelize.Style{
		Border: borders,
		Alignment: &excelize.Alignment{
			Vertical: "center",
		},
	})
	if err != nil {
		return 0, 0, err
	}

	return headerStyle, bodyStyle, nil
}

func setColumnWidths(file *excelize.File, sheet string, headers []string, rows [][]any) error {
	for i, header := range headers {
		width := displayWidth(header)
		for _, row := range rows {
			if i < len(row) {
				width = max(width, displayWidth(cellDisplayText(row[i])))
			}
		}
		width = min(max(width+2, 8), 36)

		col, err := excelize.ColumnNumberToName(i + 1)
		if err != nil {
			return err
		}
		if err := file.SetColWidth(sheet, col, col, float64(width)); err != nil {
			return err
		}
	}
	return nil
}

func normalizedGrid(sheet Sheet) ([]string, [][]any) {
	colCount := len(sheet.Headers)
	for _, row := range sheet.Rows {
		if len(row) > colCount {
			colCount = len(row)
		}
	}

	headers := make([]string, colCount)
	for i := 0; i < colCount; i++ {
		if i < len(sheet.Headers) && strings.TrimSpace(sheet.Headers[i]) != "" {
			headers[i] = sheet.Headers[i]
			continue
		}
		headers[i] = fmt.Sprintf("Column %d", i+1)
	}

	rows := make([][]any, len(sheet.Rows))
	for i, row := range sheet.Rows {
		normalized := make([]any, colCount)
		for j, value := range row {
			normalized[j] = normalizeCellValue(value)
		}
		rows[i] = normalized
	}

	return headers, rows
}

func normalizeCellValue(value any) any {
	number, ok := value.(json.Number)
	if !ok {
		return value
	}

	text := number.String()
	if strings.ContainsAny(text, ".eE") {
		if value, err := number.Float64(); err == nil {
			return value
		}
		return text
	}
	if value, err := number.Int64(); err == nil {
		return value
	}
	if value, err := number.Float64(); err == nil {
		return value
	}
	return text
}

func stringSliceToAny(values []string) []any {
	out := make([]any, len(values))
	for i, value := range values {
		out[i] = value
	}
	return out
}

func cellDisplayText(value any) string {
	switch v := value.(type) {
	case nil:
		return ""
	case json.Number:
		return v.String()
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case bool:
		if v {
			return "TRUE"
		}
		return "FALSE"
	default:
		return fmt.Sprint(v)
	}
}

func displayWidth(value string) int {
	width := 0
	for _, r := range value {
		if r > 127 {
			width += 2
			continue
		}
		width++
	}
	return width
}

func SanitizeSheetName(name string, index int) string {
	name = strings.TrimSpace(name)
	if name == "" {
		name = fmt.Sprintf("Sheet%d", index)
	}

	replacer := strings.NewReplacer(
		"[", " ",
		"]", " ",
		":", " ",
		"*", " ",
		"?", " ",
		"/", " ",
		"\\", " ",
	)
	name = strings.Join(strings.Fields(replacer.Replace(name)), " ")
	runes := []rune(name)
	if len(runes) > 31 {
		name = string(runes[:31])
	}
	if strings.TrimSpace(name) == "" {
		return fmt.Sprintf("Sheet%d", index)
	}
	return name
}

func uniqueSheetNames(sheets []Sheet) []string {
	names := make([]string, len(sheets))
	used := make(map[string]struct{}, len(sheets))
	for i, sheet := range sheets {
		base := SanitizeSheetName(sheet.Name, i+1)
		name := base
		for suffix := 2; ; suffix++ {
			key := strings.ToLower(name)
			if _, ok := used[key]; !ok {
				used[key] = struct{}{}
				names[i] = name
				break
			}

			suffixText := fmt.Sprintf(" (%d)", suffix)
			runes := []rune(base)
			maxBaseRunes := 31 - len([]rune(suffixText))
			if len(runes) > maxBaseRunes {
				base = string(runes[:maxBaseRunes])
			}
			name = base + suffixText
		}
	}
	return names
}
