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

## Structure

- `cmd/report-service`: service entrypoint
- `internal/config`: environment configuration
- `internal/httpserver`: Fiber routes and handlers
- `internal/reports`: PDF rendering packages
- `assets/fonts`: bundled fonts used by the PDF renderers
- `mock`: local preview payloads
