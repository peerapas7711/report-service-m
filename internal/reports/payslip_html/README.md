# Payslip HTML

HTML payslip renderer with one folder per payslip type.

## Routes

```sh
curl "http://localhost:8080/report/payslip/html?type=default"
curl "http://localhost:8080/report/payslip/html?type=thai_demar"
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
- `thai_demar`
- `cp`

## Template Layout

```text
templates/
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

`template.html` renders the slip markup, `style.css` is embedded into the HTML,
and `config.json` provides the type metadata and default preview data.
