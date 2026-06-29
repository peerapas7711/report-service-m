# Report Service

Small Go microservice for rendering report PDFs.

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
- `GET /payslip`
- `POST /v1/reports/payslip/render`
- `POST /v1/reports/system-access-permission/render`

Preview routes remain for local checks:

- `GET /preview/payslip`
- `GET /preview/system-access`

Payslip preview data currently uses a SQL-like mock repository. Existing mock
aliases still work:

- `GET /report/payslip/html?mock=hopinn`
- `GET /report/payslip/html?mock=tigersoft`
- `GET /report/payslip/pdf?mock=1`
- `GET /report/payslip/pdf?mock=tigersoft&count=1000`

The same repository also accepts SQL-like filters for later DB wiring:

- `company_id`
- `employee_id`
- `period_id`
- `period`
- `slip_no`
- `slip_id`

Use `count` or `total` on `/report/payslip/html` and `/report/payslip/pdf`
to generate a local batch from one mock payslip. The batch is capped at 1000
reports.

Example:

- `GET /report/payslip/html?company_id=company_hopinn&employee_id=HOP-240117&slip_no=1`

## Structure

- `cmd/report-service`: service entrypoint
- `internal/config`: environment configuration
- `internal/datasources`: report data repositories and mock data sources
- `internal/httpserver`: Fiber routes and handlers
- `internal/reports`: PDF rendering packages
- `assets/fonts`: bundled fonts used by the PDF renderers
- `mock`: local preview payloads
