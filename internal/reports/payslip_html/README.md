# Payslip HTML

HTML payslip renderer with one folder per payslip type.

## Routes

```sh
curl "http://localhost:8080/report/payslip/html?type=default"
curl "http://localhost:8080/report/payslip/html?type=thai_delmar"
curl "http://localhost:8080/report/payslip/html?type=cp"
curl -o payslip.pdf "http://localhost:8080/report/payslip/pdf?type=default&download=1"
curl -X POST "http://localhost:8080/report/payslip" \
  -H "Content-Type: application/json" \
  --data @payload.json
```

The POST report route accepts the real report request payload:
`templateCode`, `companyCode`, `format`, and `data.employees`. `format`
currently supports `html` and `pdf`.

Supported type values:

- `default` (Tigersoft)
- `thai_delmar`
- `cp`

## Template Layout

```text
templates/
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

`template.html` is the renderer entrypoint. A template can either keep its CSS
in a sibling `style.css` file, or keep CSS inline in `template.html`.
`config.json` provides the type metadata, default preview data, and optional
data-model selection.
