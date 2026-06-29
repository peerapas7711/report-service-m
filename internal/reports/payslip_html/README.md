# Payslip HTML Flow

เอกสารนี้อธิบายการออก payslip แบบ HTML ใน package `internal/reports/payslip_html` ตั้งแต่รับ request, โหลดข้อมูล, render HTML, และแปลงเป็น PDF

## ภาพรวม

Flow ปัจจุบันของ HTML slip เป็น preview/report route ที่รับข้อมูลผ่าน query string แล้วไปดึง payslip จาก repository mock แบบ SQL-like ก่อน render ด้วย HTML template

```text
HTTP request
  -> internal/httpserver/report_handlers.go
  -> loadPreviewPayslipData()
  -> payslipdata.Repository.FindPayslip()
  -> payslip_html.RenderHTML()
  -> ส่งกลับ text/html

ถ้าเป็น PDF:
  -> payslip_html.GeneratePDF()
  -> ส่งกลับ application/pdf
```

ไฟล์หลักที่เกี่ยวข้อง:

- `internal/httpserver/routes.go` ลง route `/report/payslip/html` และ `/report/payslip/pdf`
- `internal/httpserver/report_handlers.go` รับ request, query ข้อมูล, override บาง field, แล้วเรียก renderer
- `internal/datasources/payslipdata` เป็น repository สำหรับหา payslip จาก mock data แบบจำลอง table/SQL
- `internal/reports/payslip/model.go` เป็น data model กลางของ payslip
- `internal/reports/payslip_html/renderer.go` render HTML จาก template
- `internal/reports/payslip_html/pdf.go` แปลง HTML เป็น PDF ด้วย headless Chrome/chromedp
- `internal/reports/payslip_html/templates` เก็บ HTML templates และ CSS

## Endpoint ที่ใช้

### ออก HTML

```http
GET /report/payslip/html
```

ตัวอย่าง:

```sh
curl "http://localhost:8080/report/payslip/html?mock=hopinn"
curl "http://localhost:8080/report/payslip/html?mock=tigersoft&template=tigersoft"
curl "http://localhost:8080/report/payslip/html?company_id=company_hopinn&employee_id=HOP-240117&slip_no=1"
```

ผลลัพธ์:

- `Content-Type: text/html; charset=utf-8`
- body เป็น HTML ที่ render เสร็จแล้ว

### ออก PDF จาก HTML

```http
GET /report/payslip/pdf
```

ตัวอย่าง:

```sh
curl -o payslip.pdf "http://localhost:8080/report/payslip/pdf?mock=hopinn"
curl -o payslip.pdf "http://localhost:8080/report/payslip/pdf?mock=2&download=1"
```

ผลลัพธ์:

- `Content-Type: application/pdf`
- `Content-Disposition` เป็น `inline` ตาม default
- ถ้าส่ง `download=1` จะเปลี่ยนเป็น `attachment`

## ส่งข้อมูลยังไง

สำหรับ flow HTML/PDF ปัจจุบัน ข้อมูลไม่ได้ส่งเป็น JSON body โดยตรง แต่ส่งตัวกรองผ่าน query string เพื่อให้ server ไปหา payslip จาก `payslipdata.Repository`

Query ที่ใช้เลือกข้อมูล:

| Query | ความหมาย |
| --- | --- |
| `mock` | เลือก mock data ตามชื่อหรือ alias เช่น `hopinn`, `tigersoft`, `1`, `2` |
| `slip_id` | เลือก payslip ด้วย slip id เช่น `slip_hopinn` |
| `company_id` | กรอง company id เช่น `company_hopinn` |
| `employee_id` | กรอง employee id เช่น `HOP-240117` |
| `period_id` | กรอง period id เช่น `period_hopinn` |
| `period` | กรอง period label เช่น `16/2026 : February - Center` |
| `slip_no` | กรองเลข slip เช่น `1` |

Query ที่ใช้ override การแสดงผล:

| Query | ความหมาย |
| --- | --- |
| `template` | เลือก template เช่น `default`, `modern`, `tigersoft`, `tiger_soft` |
| `company_name` | override ชื่อบริษัทก่อน render |
| `logo` หรือ `logo_url` | override logo บริษัทก่อน render |
| `download` | ใช้กับ `/report/payslip/pdf`; ถ้าเป็นค่าจริง เช่น `1`, `true`, `yes` จะส่ง PDF แบบ download |

ถ้าไม่ส่ง `mock` และไม่ส่งตัวกรอง SQL-like เลย ระบบจะ default เป็น `mock=hopinn`

Mock data ที่โหลดตอนเริ่ม handler:

| mock | file | aliases |
| --- | --- | --- |
| `hopinn` | `mock/payslip_hopinn.json` | `1`, `default`, `modern` |
| `tigersoft` | `mock/payslip_tigersoft.json` | `2`, `tiger_soft` |
| `bluewave` | `mock/payslip_bluewave.json` | `3` |
| `kubota` | `mock/payslip_kubota.json` | `4` |

## รูปแบบข้อมูล payslip

Data model อยู่ที่ `internal/reports/payslip/model.go` และ HTML renderer ใช้ type เดียวกันคือ `payslip.Payslip`

โครงสร้างหลัก:

```json
{
  "template_id": "modern",
  "report": {
    "company": {
      "name": "Company Name",
      "logo": "https://example.com/logo.png"
    },
    "document": {
      "title": "PAY SLIP",
      "confidential_title": "PRIVATE & CONFIDENTIAL",
      "confidential_subtitle": "(TO BE OPENED BY ADDRESSEE ONLY)"
    },
    "payroll": {
      "period": "16/2026 : February - Center",
      "pay_date": "27/02/2026",
      "slip_no": "1"
    },
    "employee": {
      "name": "Employee Name",
      "emp_id": "EMP-001",
      "joined_date": "01/10/2024",
      "division": "Operations",
      "department": "Front Office",
      "section": "Center",
      "position": "Staff",
      "tax_id": "1103700123456"
    },
    "earnings": [
      { "label": "Salary", "value": "30,000.00" }
    ],
    "deductions": [
      { "label": "Social Security", "value": "750.00" }
    ],
    "work_stats": [
      { "label": "Works", "value": "30:00:00" }
    ],
    "accumulations": [
      { "label": "Y-T-D Income", "value": "245,000.00" }
    ],
    "totals": {
      "total_income": "30,000.00",
      "total_deduct": "750.00",
      "net_pay": "29,250.00",
      "bank_account_no": "123-4-56789-0"
    }
  }
}
```

หมายเหตุ:

- `value` ใน `LineItem` และ `Totals` เป็น string เพราะระบบรับค่าที่ format มาแล้ว เช่น comma, decimal, หรือเวลา
- `template_id` ใน data จะถูกใช้เป็นค่า default ของ template ถ้า request ไม่ส่ง `template`
- `company.logo` เป็น URL หรือ data URL ได้ ถ้าว่าง renderer จะใช้ `templates/assets/logo.png` เป็น fallback

## การทำงานใน HTTP handler

1. `routes.go` map route:
   - `/report/payslip/html` -> `previewPayslipHTML`
   - `/report/payslip/pdf` -> `previewPayslipHTMLPDF`
2. `previewPayslipHTML` และ `previewPayslipHTMLPDF` เรียก `loadPreviewPayslipData(c)`
3. `loadPreviewPayslipData` สร้าง `payslipdata.Query` จาก query string ด้วย `payslipQuery(c)`
4. `payslipRepo.FindPayslip()` หา payslip จาก repository
5. ถ้ามี query `company_name`, `logo`/`logo_url`, หรือ `template` จะ override ลง data ก่อน render
6. เรียก `payslip_html.RenderHTML(c.Query("template", data.TemplateID), data)`
7. ถ้าเป็น `/report/payslip/html` จะส่ง HTML กลับทันที
8. ถ้าเป็น `/report/payslip/pdf` จะส่ง HTML เข้า `payslip_html.GeneratePDF()` แล้วส่ง PDF กลับ

กรณีไม่เจอข้อมูล จะตอบ `400` พร้อม JSON:

```json
{
  "error": "payslip mock not found",
  "available_mocks": ["1", "2", "3", "4", "bluewave", "default", "hopinn", "kubota", "modern", "tiger_soft", "tigersoft"]
}
```

## การทำงานของ Repository

`internal/datasources/payslipdata/mock_sql_repository.go` แปลง mock JSON ให้เป็น row จำลองหลาย table:

- `Companies`
- `Employees`
- `PayrollPeriods`
- `Payslips`
- `LineItems`
- `Texts`
- `Aliases`

ตอน query:

1. ถ้ามี `slip_id` จะหา payslip ด้วย slip id ก่อน
2. ถ้ามี `mock` จะ resolve alias เป็น slip id แล้วหา payslip
3. ถ้าไม่มีสองตัวบน จะวนหา payslip ด้วย SQL-like filters เช่น company, employee, period, slip no
4. เมื่อเจอ row แล้ว `assemblePayslip()` จะประกอบกลับเป็น `payslip.Payslip`

นี่ทำให้ route preview ใช้วิธีเรียกใกล้เคียงกับการต่อ database จริงในอนาคต แต่ยังอ่านจาก mock file อยู่

## การทำงานของ RenderHTML

`RenderHTML(templateName string, data Payslip) (string, error)` ทำงานตามลำดับนี้:

1. Normalize ชื่อ template ด้วย `normalizeTemplateName()`
   - ค่าว่าง, `default`, `modern`, `payslip_default` -> `payslip_default`
   - `tigersoft`, `tiger_soft`, `payslip_tigersoft` -> `payslip_tigersoft`
   - ค่าอื่นจะพยายามหาไฟล์ `templates/<name>.html`
2. Parse template จาก embedded filesystem ด้วย `template.ParseFS`
3. สร้าง view model ด้วย `buildViewModel()`
   - ใส่ fallback logo ถ้า company logo ว่าง
   - อ่าน `templates/assets/report.css`
   - แทน path font Sarabun ให้เป็น data URL จาก `assets/fonts/Sarabun/*.ttf`
   - สร้าง `EmployeeFields` และ `PayrollFields` สำหรับแสดงใน `<dl>`
   - สร้าง flag เช่น `HasEarnings`, `HasDeductions`, `HasWorkStats`, `HasAccumulation`
4. Execute template แล้ว return HTML string

Templates ที่มีตอนนี้:

| template | file | จุดเด่น |
| --- | --- | --- |
| `payslip_default` | `templates/payslip_default.html` | layout default, แสดง earnings/deductions, summary, work stats, accumulations |
| `payslip_tigersoft` | `templates/payslip_tigersoft.html` | layout accent สีแดง, summary แบบ Tigersoft |

CSS อยู่ที่ `templates/assets/report.css` โดยกำหนด `@page size: A4`, margin, font Sarabun, screen preview style, print style, และ rule ป้องกัน table/card ถูกตัดกลางหน้า

## การทำงานของ GeneratePDF

`GeneratePDF(html string) ([]byte, error)` ใน `pdf.go` ใช้ `chromedp`:

1. ตรวจว่า HTML ไม่ว่าง
2. เปิด headless Chrome context พร้อม option เช่น headless, disable GPU
3. แปลง HTML เป็น `data:text/html;charset=utf-8;base64,...`
4. Navigate ไปที่ data URL
5. รอ `body` พร้อม
6. เรียก Chrome DevTools `page.PrintToPDF()`
   - `WithPrintBackground(true)` เพื่อให้สีพื้นหลัง/ตารางติดไปใน PDF
   - `WithPreferCSSPageSize(true)` เพื่อใช้ขนาดจาก CSS `@page`
7. return bytes ของ PDF

## ทำยังไงจนออกได้

ขั้นตอน local preview:

```sh
go run ./cmd/report-service
```

เปิด HTML:

```text
http://localhost:8080/report/payslip/html?mock=hopinn
```

เปิด PDF จาก HTML:

```text
http://localhost:8080/report/payslip/pdf?mock=hopinn
```

เลือก template:

```text
http://localhost:8080/report/payslip/html?mock=tigersoft&template=tigersoft
```

เลือกข้อมูลแบบ SQL-like:

```text
http://localhost:8080/report/payslip/html?company_id=company_hopinn&employee_id=HOP-240117&slip_no=1
```

Override logo/name:

```text
http://localhost:8080/report/payslip/html?mock=hopinn&company_name=Demo%20Company&logo_url=https://example.com/logo.png
```

## ถ้าต้องเพิ่ม template ใหม่

1. เพิ่มไฟล์ `internal/reports/payslip_html/templates/<template_name>.html`
2. ใช้ `{{ define "<template_name>.html" }}` ให้ชื่อตรงกับไฟล์
3. ใช้ field จาก view model เช่น `.Report`, `.EmployeeFields`, `.PayrollFields`, `.Stylesheet`
4. เพิ่มชื่อใน `AvailableTemplates()` ถ้าต้องการให้ระบบอื่น list ได้
5. เพิ่ม case ใน `normalizeTemplateName()` ถ้าต้องการ alias เช่น `new_brand` หรือ `brand-a`
6. ทดสอบด้วย `/report/payslip/html?mock=hopinn&template=<template_name>`

## ข้อควรรู้

- Package นี้ render เฉพาะ HTML/PDF จาก HTML ส่วน endpoint `POST /v1/reports/payslip/render` ยังใช้ renderer PDF คนละชุดใน `internal/reports/payslip`
- HTML template ใช้ `html/template` จึง escape ข้อมูล string ตามปกติ
- PDF generation ต้องมี Chrome/Chromium ที่ `chromedp` ใช้งานได้ใน environment ที่รัน service
- ถ้า font Sarabun ใน `assets/fonts/Sarabun` หาไม่เจอ CSS จะเหลือ path เดิม ทำให้ environment ปลายทางต้อง serve font path เองหรือใช้ fallback font
