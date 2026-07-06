package excel

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
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

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	if err := addZipFile(zw, "[Content_Types].xml", contentTypesXML(len(workbook.Sheets))); err != nil {
		return nil, err
	}
	if err := addZipFile(zw, "_rels/.rels", rootRelsXML()); err != nil {
		return nil, err
	}
	if err := addZipFile(zw, "docProps/core.xml", corePropsXML()); err != nil {
		return nil, err
	}
	if err := addZipFile(zw, "docProps/app.xml", appPropsXML()); err != nil {
		return nil, err
	}
	if err := addZipFile(zw, "xl/workbook.xml", workbookXML(workbook.Sheets)); err != nil {
		return nil, err
	}
	if err := addZipFile(zw, "xl/_rels/workbook.xml.rels", workbookRelsXML(len(workbook.Sheets))); err != nil {
		return nil, err
	}
	if err := addZipFile(zw, "xl/styles.xml", stylesXML()); err != nil {
		return nil, err
	}

	for i, sheet := range workbook.Sheets {
		xml, err := worksheetXML(sheet)
		if err != nil {
			return nil, err
		}
		if err := addZipFile(zw, fmt.Sprintf("xl/worksheets/sheet%d.xml", i+1), xml); err != nil {
			return nil, err
		}
	}

	if err := zw.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func addZipFile(zw *zip.Writer, name string, data string) error {
	w, err := zw.Create(name)
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(data))
	return err
}

func contentTypesXML(sheetCount int) string {
	var b strings.Builder
	b.WriteString(xmlHeader())
	b.WriteString(`<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">`)
	b.WriteString(`<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>`)
	b.WriteString(`<Default Extension="xml" ContentType="application/xml"/>`)
	b.WriteString(`<Override PartName="/docProps/app.xml" ContentType="application/vnd.openxmlformats-officedocument.extended-properties+xml"/>`)
	b.WriteString(`<Override PartName="/docProps/core.xml" ContentType="application/vnd.openxmlformats-package.core-properties+xml"/>`)
	b.WriteString(`<Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/>`)
	b.WriteString(`<Override PartName="/xl/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.styles+xml"/>`)
	for i := 1; i <= sheetCount; i++ {
		fmt.Fprintf(&b, `<Override PartName="/xl/worksheets/sheet%d.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/>`, i)
	}
	b.WriteString(`</Types>`)
	return b.String()
}

func rootRelsXML() string {
	return xmlHeader() + `<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">` +
		`<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/>` +
		`<Relationship Id="rId2" Type="http://schemas.openxmlformats.org/package/2006/relationships/metadata/core-properties" Target="docProps/core.xml"/>` +
		`<Relationship Id="rId3" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/extended-properties" Target="docProps/app.xml"/>` +
		`</Relationships>`
}

func corePropsXML() string {
	now := time.Now().UTC().Format(time.RFC3339)
	return xmlHeader() + `<cp:coreProperties xmlns:cp="http://schemas.openxmlformats.org/package/2006/metadata/core-properties" ` +
		`xmlns:dc="http://purl.org/dc/elements/1.1/" ` +
		`xmlns:dcterms="http://purl.org/dc/terms/" ` +
		`xmlns:dcmitype="http://purl.org/dc/dcmitype/" ` +
		`xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">` +
		`<dc:creator>report-service-m</dc:creator>` +
		`<cp:lastModifiedBy>report-service-m</cp:lastModifiedBy>` +
		`<dcterms:created xsi:type="dcterms:W3CDTF">` + escapeXML(now) + `</dcterms:created>` +
		`<dcterms:modified xsi:type="dcterms:W3CDTF">` + escapeXML(now) + `</dcterms:modified>` +
		`</cp:coreProperties>`
}

func appPropsXML() string {
	return xmlHeader() + `<Properties xmlns="http://schemas.openxmlformats.org/officeDocument/2006/extended-properties" ` +
		`xmlns:vt="http://schemas.openxmlformats.org/officeDocument/2006/docPropsVTypes">` +
		`<Application>report-service-m</Application>` +
		`</Properties>`
}

func workbookXML(sheets []Sheet) string {
	var b strings.Builder
	sheetNames := uniqueSheetNames(sheets)
	b.WriteString(xmlHeader())
	b.WriteString(`<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" `)
	b.WriteString(`xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">`)
	b.WriteString(`<sheets>`)
	for i, sheetName := range sheetNames {
		fmt.Fprintf(
			&b,
			`<sheet name="%s" sheetId="%d" r:id="rId%d"/>`,
			escapeXML(sheetName),
			i+1,
			i+1,
		)
	}
	b.WriteString(`</sheets></workbook>`)
	return b.String()
}

func workbookRelsXML(sheetCount int) string {
	var b strings.Builder
	b.WriteString(xmlHeader())
	b.WriteString(`<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">`)
	for i := 1; i <= sheetCount; i++ {
		fmt.Fprintf(&b, `<Relationship Id="rId%d" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet%d.xml"/>`, i, i)
	}
	fmt.Fprintf(&b, `<Relationship Id="rId%d" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/>`, sheetCount+1)
	b.WriteString(`</Relationships>`)
	return b.String()
}

func stylesXML() string {
	return xmlHeader() + `<styleSheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">` +
		`<numFmts count="1"><numFmt numFmtId="164" formatCode="@"/></numFmts>` +
		`<fonts count="2">` +
		`<font><sz val="11"/><color theme="1"/><name val="Arial"/></font>` +
		`<font><b/><sz val="11"/><color rgb="FFFFFFFF"/><name val="Arial"/></font>` +
		`</fonts>` +
		`<fills count="3">` +
		`<fill><patternFill patternType="none"/></fill>` +
		`<fill><patternFill patternType="gray125"/></fill>` +
		`<fill><patternFill patternType="solid"><fgColor rgb="FF0F766E"/><bgColor indexed="64"/></patternFill></fill>` +
		`</fills>` +
		`<borders count="2">` +
		`<border><left/><right/><top/><bottom/><diagonal/></border>` +
		`<border><left style="thin"><color rgb="FFCBD5E1"/></left><right style="thin"><color rgb="FFCBD5E1"/></right><top style="thin"><color rgb="FFCBD5E1"/></top><bottom style="thin"><color rgb="FFCBD5E1"/></bottom><diagonal/></border>` +
		`</borders>` +
		`<cellStyleXfs count="1"><xf numFmtId="0" fontId="0" fillId="0" borderId="0"/></cellStyleXfs>` +
		`<cellXfs count="3">` +
		`<xf numFmtId="0" fontId="0" fillId="0" borderId="1" xfId="0" applyBorder="1"/>` +
		`<xf numFmtId="0" fontId="1" fillId="2" borderId="1" xfId="0" applyFont="1" applyFill="1" applyBorder="1" applyAlignment="1"><alignment horizontal="center" vertical="center" wrapText="1"/></xf>` +
		`<xf numFmtId="164" fontId="0" fillId="0" borderId="1" xfId="0" applyNumberFormat="1" applyBorder="1" applyAlignment="1"><alignment vertical="center"/></xf>` +
		`</cellXfs>` +
		`<cellStyles count="1"><cellStyle name="Normal" xfId="0" builtinId="0"/></cellStyles>` +
		`<dxfs count="0"/>` +
		`<tableStyles count="0" defaultTableStyle="TableStyleMedium2" defaultPivotStyle="PivotStyleLight16"/>` +
		`</styleSheet>`
}

func worksheetXML(sheet Sheet) (string, error) {
	headers, rows := normalizedGrid(sheet)
	if len(headers) == 0 {
		return "", errors.New("sheet must have at least one column")
	}

	rowCount := len(rows) + 1
	colCount := len(headers)
	lastRef := cellRef(rowCount, colCount)
	filterRef := "A1:" + lastRef

	var b strings.Builder
	b.WriteString(xmlHeader())
	b.WriteString(`<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" `)
	b.WriteString(`xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">`)
	fmt.Fprintf(&b, `<dimension ref="A1:%s"/>`, lastRef)
	b.WriteString(`<sheetViews><sheetView workbookViewId="0"><pane ySplit="1" topLeftCell="A2" activePane="bottomLeft" state="frozen"/></sheetView></sheetViews>`)
	b.WriteString(`<sheetFormatPr defaultRowHeight="18"/>`)
	writeColumns(&b, headers, rows)
	b.WriteString(`<sheetData>`)
	b.WriteString(`<row r="1" ht="24" customHeight="1">`)
	for col, header := range headers {
		writeStringCell(&b, 1, col+1, header, 1)
	}
	b.WriteString(`</row>`)
	for rowIdx, row := range rows {
		excelRow := rowIdx + 2
		fmt.Fprintf(&b, `<row r="%d">`, excelRow)
		for colIdx := 0; colIdx < colCount; colIdx++ {
			writeValueCell(&b, excelRow, colIdx+1, row[colIdx])
		}
		b.WriteString(`</row>`)
	}
	b.WriteString(`</sheetData>`)
	fmt.Fprintf(&b, `<autoFilter ref="%s"/>`, filterRef)
	b.WriteString(`<pageMargins left="0.7" right="0.7" top="0.75" bottom="0.75" header="0.3" footer="0.3"/>`)
	b.WriteString(`</worksheet>`)

	return b.String(), nil
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
		copy(normalized, row)
		rows[i] = normalized
	}

	return headers, rows
}

func writeColumns(b *strings.Builder, headers []string, rows [][]any) {
	b.WriteString(`<cols>`)
	for i, header := range headers {
		width := displayWidth(header)
		for _, row := range rows {
			if i < len(row) {
				width = max(width, displayWidth(cellDisplayText(row[i])))
			}
		}
		width = min(max(width+2, 8), 36)
		fmt.Fprintf(b, `<col min="%d" max="%d" width="%d" customWidth="1"/>`, i+1, i+1, width)
	}
	b.WriteString(`</cols>`)
}

func writeValueCell(b *strings.Builder, row, col int, value any) {
	switch v := value.(type) {
	case nil:
		fmt.Fprintf(b, `<c r="%s" s="2"/>`, cellRef(row, col))
	case json.Number:
		if _, err := v.Float64(); err == nil && strings.TrimSpace(v.String()) != "" {
			writeNumberCell(b, row, col, v.String())
			return
		}
		writeStringCell(b, row, col, v.String(), 2)
	case float64:
		if math.IsNaN(v) || math.IsInf(v, 0) {
			writeStringCell(b, row, col, strconv.FormatFloat(v, 'f', -1, 64), 2)
			return
		}
		writeNumberCell(b, row, col, strconv.FormatFloat(v, 'f', -1, 64))
	case float32:
		writeNumberCell(b, row, col, strconv.FormatFloat(float64(v), 'f', -1, 32))
	case int:
		writeNumberCell(b, row, col, strconv.Itoa(v))
	case int64:
		writeNumberCell(b, row, col, strconv.FormatInt(v, 10))
	case int32:
		writeNumberCell(b, row, col, strconv.FormatInt(int64(v), 10))
	case bool:
		value := "0"
		if v {
			value = "1"
		}
		fmt.Fprintf(b, `<c r="%s" t="b" s="0"><v>%s</v></c>`, cellRef(row, col), value)
	default:
		writeStringCell(b, row, col, fmt.Sprint(v), 2)
	}
}

func writeNumberCell(b *strings.Builder, row, col int, value string) {
	fmt.Fprintf(b, `<c r="%s" s="0"><v>%s</v></c>`, cellRef(row, col), escapeXML(value))
}

func writeStringCell(b *strings.Builder, row, col int, value string, style int) {
	ref := cellRef(row, col)
	if value == "" {
		fmt.Fprintf(b, `<c r="%s" s="%d"/>`, ref, style)
		return
	}
	preserve := ""
	if strings.TrimSpace(value) != value {
		preserve = ` xml:space="preserve"`
	}
	fmt.Fprintf(b, `<c r="%s" t="inlineStr" s="%d"><is><t%s>%s</t></is></c>`, ref, style, preserve, escapeXML(value))
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

func cellRef(row, col int) string {
	return columnName(col) + strconv.Itoa(row)
}

func columnName(col int) string {
	name := ""
	for col > 0 {
		col--
		name = string(rune('A'+col%26)) + name
		col /= 26
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

func escapeXML(value string) string {
	var b strings.Builder
	if err := xml.EscapeText(&b, []byte(value)); err != nil {
		return value
	}
	return b.String()
}

func xmlHeader() string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`
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
