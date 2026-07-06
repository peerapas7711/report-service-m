package httpserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"report-service-m/internal/reports/excel"

	"github.com/gofiber/fiber/v2"
)

type excelReportRequest struct {
	Filename string
	Workbook excel.Workbook
}

type excelReportPayload struct {
	Filename  string              `json:"filename"`
	SheetName string              `json:"sheetName"`
	Name      string              `json:"name"`
	Headers   []string            `json:"headers"`
	Rows      []json.RawMessage   `json:"rows"`
	Data      json.RawMessage     `json:"data"`
	Sheets    []excelSheetPayload `json:"sheets"`
}

type excelSheetPayload struct {
	Name      string            `json:"name"`
	SheetName string            `json:"sheetName"`
	Headers   []string          `json:"headers"`
	Rows      []json.RawMessage `json:"rows"`
}

type decodedExcelRow struct {
	array  []any
	object map[string]any
}

func loadExcelReportRequest(c *fiber.Ctx) (excelReportRequest, error) {
	body := bytes.TrimSpace(c.Body())
	if len(body) == 0 {
		return excelReportRequest{}, fiber.NewError(fiber.StatusBadRequest, "request body is required")
	}

	var payload excelReportPayload
	if err := decodeJSONUseNumber(body, &payload); err != nil {
		return excelReportRequest{}, fiber.NewError(fiber.StatusBadRequest, "invalid excel report request json: "+err.Error())
	}

	if len(payload.Sheets) == 0 && len(payload.Rows) == 0 && len(payload.Data) > 0 {
		wrapperFilename := payload.Filename
		wrapperSheetName := firstNonEmpty(payload.SheetName, payload.Name)
		var nested excelReportPayload
		if err := decodeJSONUseNumber(payload.Data, &nested); err != nil {
			return excelReportRequest{}, fiber.NewError(fiber.StatusBadRequest, "invalid excel report data json: "+err.Error())
		}
		if strings.TrimSpace(nested.Filename) == "" {
			nested.Filename = wrapperFilename
		}
		if strings.TrimSpace(nested.SheetName) == "" && strings.TrimSpace(nested.Name) == "" {
			nested.SheetName = wrapperSheetName
		}
		payload = nested
	}

	filename := firstNonEmpty(c.Query("filename"), payload.Filename, "excel_report.xlsx")
	if !strings.HasSuffix(strings.ToLower(filename), ".xlsx") {
		filename += ".xlsx"
	}

	workbook, err := buildExcelWorkbook(payload)
	if err != nil {
		return excelReportRequest{}, err
	}

	return excelReportRequest{
		Filename: filename,
		Workbook: workbook,
	}, nil
}

func buildExcelWorkbook(payload excelReportPayload) (excel.Workbook, error) {
	if len(payload.Sheets) > 0 {
		sheets := make([]excel.Sheet, 0, len(payload.Sheets))
		for i, payloadSheet := range payload.Sheets {
			sheet, err := buildExcelSheet(payloadSheet, fmt.Sprintf("Sheet%d", i+1))
			if err != nil {
				return excel.Workbook{}, fmt.Errorf("sheets[%d]: %w", i, err)
			}
			sheets = append(sheets, sheet)
		}
		return excel.Workbook{Sheets: sheets}, nil
	}

	sheet, err := buildExcelSheet(excelSheetPayload{
		Name:      firstNonEmpty(payload.SheetName, payload.Name),
		SheetName: payload.SheetName,
		Headers:   payload.Headers,
		Rows:      payload.Rows,
	}, "AfterProcess")
	if err != nil {
		return excel.Workbook{}, err
	}

	return excel.Workbook{Sheets: []excel.Sheet{sheet}}, nil
}

func buildExcelSheet(payload excelSheetPayload, fallbackName string) (excel.Sheet, error) {
	if len(payload.Rows) == 0 {
		return excel.Sheet{}, fiber.NewError(fiber.StatusBadRequest, "rows is required")
	}

	headers, rows, err := normalizeExcelRows(payload.Headers, payload.Rows)
	if err != nil {
		return excel.Sheet{}, err
	}
	if len(headers) == 0 {
		return excel.Sheet{}, fiber.NewError(fiber.StatusBadRequest, "headers or row columns are required")
	}

	return excel.Sheet{
		Name:    firstNonEmpty(payload.SheetName, payload.Name, fallbackName),
		Headers: headers,
		Rows:    rows,
	}, nil
}

func normalizeExcelRows(headers []string, rawRows []json.RawMessage) ([]string, [][]any, error) {
	normalizedHeaders := append([]string(nil), headers...)
	decoded := make([]decodedExcelRow, 0, len(rawRows))
	extraObjectKeys := map[string]struct{}{}
	maxArrayColumns := len(normalizedHeaders)

	for i, rawRow := range rawRows {
		var arrayRow []any
		if err := decodeJSONUseNumber(rawRow, &arrayRow); err == nil {
			decoded = append(decoded, decodedExcelRow{array: arrayRow})
			if len(arrayRow) > maxArrayColumns {
				maxArrayColumns = len(arrayRow)
			}
			continue
		}

		var objectRow map[string]any
		if err := decodeJSONUseNumber(rawRow, &objectRow); err != nil {
			return nil, nil, fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("rows[%d] must be an array or object", i))
		}
		decoded = append(decoded, decodedExcelRow{object: objectRow})
		for key := range objectRow {
			extraObjectKeys[key] = struct{}{}
		}
	}

	if len(normalizedHeaders) == 0 && len(extraObjectKeys) > 0 {
		normalizedHeaders = sortedKeys(extraObjectKeys)
	} else if len(extraObjectKeys) > 0 {
		known := make(map[string]struct{}, len(normalizedHeaders))
		for _, header := range normalizedHeaders {
			known[header] = struct{}{}
		}
		for _, key := range sortedKeys(extraObjectKeys) {
			if _, ok := known[key]; !ok {
				normalizedHeaders = append(normalizedHeaders, key)
			}
		}
	}

	for len(normalizedHeaders) < maxArrayColumns {
		normalizedHeaders = append(normalizedHeaders, fmt.Sprintf("Column %d", len(normalizedHeaders)+1))
	}

	rows := make([][]any, 0, len(decoded))
	for _, row := range decoded {
		if row.object != nil {
			values := make([]any, len(normalizedHeaders))
			for i, header := range normalizedHeaders {
				values[i] = row.object[header]
			}
			rows = append(rows, values)
			continue
		}
		rows = append(rows, row.array)
	}

	return normalizedHeaders, rows, nil
}

func sortedKeys(values map[string]struct{}) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func decodeJSONUseNumber(data []byte, v any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	return decoder.Decode(v)
}
