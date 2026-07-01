# Report Service

Small Go microservice for payslip HTML rendering and Chrome PDF export.

## Run

```sh
go run ./cmd/report-service
```

Environment:

- `PORT`, default `8080`
- `APP_ENV`, default `local`
- `BODY_LIMIT_MB`, default `10`

## HTTP API

- `GET /health`
- `GET /ready`
- `POST /report/payslip`
- `GET /report/payslip/html`
- `POST /report/payslip/html`
- `GET /report/payslip/pdf`
- `POST /report/payslip/pdf`

Payslip type is selected with the `type` query parameter. When omitted,
Tigersoft is used as the default.

- `GET /report/payslip/html?type=default` (Tigersoft)
- `GET /report/payslip/html?type=thai_delmar`
- `GET /report/payslip/html?type=cp`
- `GET /report/payslip/pdf?type=default&download=1`

Use `POST /report/payslip` when another service sends real report data. The
handler selects the template from `templateCode`, validates `companyCode`, and
returns the response format from `format` (`html` or `pdf`). The
`/report/payslip/html` and `/report/payslip/pdf` POST routes accept the same
body but force the output format from the path.

```sh
curl -X POST "http://localhost:8083/report/payslip" \
  -H "Content-Type: application/json" \
  --data '{
    "templateCode": "PAYROLL/PAYSLIP_THAI_DELMAR",
    "format": "pdf",
    "companyCode": "1",
    "data": {
      "company": "บริษัท ตังค์ จำกัด",
      "companyEn": "Tung Co. , Ltd.",
      "address": "700/359 ม.6",
      "tel": "038-743923",
      "logoUrl": "",
      "employees": [
        {
          "empCode": "0160",
          "empName": "Mr.Prasert  TEST",
          "position": "02-16 : UF-ALL / AF-ALL",
          "department": "MANU - Production Time : Production",
          "empType": "",
          "startDate": "28/08/2543",
          "bankName": "Bank of Ayudhaya Public Company Limited",
          "bankAccount": "4474029128",
          "year": "2569",
          "month": "June",
          "periodNo": "01",
          "payDate": "30/Jun/2026",
          "earnings": [{ "name": "Salary", "amount": 25324 }],
          "deductions": [],
          "totalEarnings": 25324,
          "totalDeductions": 0,
          "netPay": 25324,
          "ytdIncome": 66247.31,
          "ytdTax": 0,
          "ytdSocialSecurity": 1625,
          "ytdProvidentFund": 5064.8,
          "ytdProvidentFundCompany": 3038.88
        },
         {
          "empCode": "0160111",
          "empName": "Mr.Prasert  TEST",
          "position": "02-16 : UF-ALL / AF-ALL",
          "department": "MANU - Production Time : Production",
          "empType": "",
          "startDate": "28/08/2543",
          "bankName": "Bank of Ayudhaya Public Company Limited",
          "bankAccount": "4474029128",
          "year": "2569",
          "month": "June",
          "periodNo": "01",
          "payDate": "30/Jun/2026",
          "earnings": [{ "name": "Salary", "amount": 25324 }],
          "deductions": [],
          "totalEarnings": 25324,
          "totalDeductions": 0,
          "netPay": 25324,
          "ytdIncome": 66247.31,
          "ytdTax": 0,
          "ytdSocialSecurity": 1625,
          "ytdProvidentFund": 5064.8,
          "ytdProvidentFundCompany": 3038.88
        }
      ]
    }
  }'
```

Use `count` or `total` on the GET preview routes to generate a local batch from
the selected type's config data. POST report requests render one payslip per
entry in `data.employees`.

## Templates

Payslip templates live under `internal/reports/payslip_html/templates`:

```text
payslip/
  default/
    template.html
    style.css
    config.json
  thai_delmar/
    template.html
    config.json
  cp/
    config.json
```

Each `config.json` owns the sample data for that type. Templates can either use
an external `style.css` that is embedded by the renderer, or keep CSS inline in
`template.html`. `cp` currently reuses the default template assets.

## Structure

- `cmd/report-service`: service entrypoint
- `internal/config`: environment configuration
- `internal/httpserver`: Fiber routes and handlers
- `internal/reports/payslip_html`: payslip HTML renderer, PDF export, templates
- `assets/fonts`: bundled fonts used by the HTML renderer
