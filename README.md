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
- `GET /report/payslip/html`
- `GET /report/payslip/pdf`

Payslip type is selected with the `type` query parameter. When omitted,
Tigersoft is used as the default.

- `GET /report/payslip/html?type=default` (Tigersoft)
- `GET /report/payslip/html?type=thai_demar`
- `GET /report/payslip/html?type=cp`
- `GET /report/payslip/pdf?type=default&download=1`

Use `count` or `total` on `/report/payslip/html` and `/report/payslip/pdf`
to generate a local batch from the selected type's config data. The batch is
capped at 1000 reports.

## Templates

Payslip templates live under `internal/reports/payslip_html/templates`:

```text
payslip/
  default/
    template.html
    style.css
    config.json
  thai_demar/
    template.html
    style.css
    config.json
  cp/
    template.html
    style.css
    config.json
```

Each `config.json` owns the sample data for that type. Each `style.css` is
embedded into the rendered HTML and inlined with Sarabun font data for PDF
generation.

## Structure

- `cmd/report-service`: service entrypoint
- `internal/config`: environment configuration
- `internal/httpserver`: Fiber routes and handlers
- `internal/reports/payslip_html`: payslip HTML renderer, PDF export, templates
- `assets/fonts`: bundled fonts used by the HTML renderer
