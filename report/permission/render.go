package permission

import (
	"bytes"

	"github.com/jung-kurt/gofpdf"
)

func RenderPermissionPDF(data Report) ([]byte, error) {
	pdf := gofpdf.New("L", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.SetAutoPageBreak(true, 15)
	pdf.AddPage()

	// โหลดฟอนต์ไทย
	pdf.AddUTF8Font("TH", "", "assets/fonts/Noto_Sans_Myanmar/NotoSansMyanmar-Regular.ttf")
	if !pdf.Ok() {
		return nil, pdf.Error()
	}
	pdf.AddUTF8Font("TH", "B", "assets/fonts/Noto_Sans_Myanmar/NotoSansMyanmar-Bold.ttf")
	if !pdf.Ok() {
		return nil, pdf.Error()
	}

	// 1) Header
	drawHeader(pdf, data)

	// 2) User info
	// drawUserInfo(pdf, data)

	// 3) Sections
	drawChecklistBox(pdf, "โปรแกรม/เงินเดือน", data.User.Programs)
	drawTableCheckBox(pdf, "บริษัท", "บริษัท", data.User.Companies)
	drawTableCheckBox(pdf, "ประเภทพนักงาน", "ประเภทพนักงาน", data.User.EmpTypes)

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
