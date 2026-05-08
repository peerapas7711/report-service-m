package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"report-service-m/report"
	"report-service-m/report/payslip"
	"report-service-m/report/queueticket"
	"report-service-m/report/systemaccesspermission"

	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/gofiber/fiber/v2"
)

func main() {
	// อ่านไฟล์ + Unmarshal
	// partReadFileAndUnmarshal()

	//  ตัวอย่าง slice / map / struct
	// partBasicTypes()

	//  JSON -> map + ลองดึง nested + type
	// partJSONMapAndTypes()

	//  getByPath แบบ split "."
	// partGetByPath()

	//  HTTP Server
	// partHTTPServer()

	// Render report (Text/PDF)
	// partReportRender()

	// chromedp HTML -> PDF
	// partChromeDP()

	// APi templant
	// mux := http.NewServeMux()
	// mux.HandleFunc("/render/pdf", renderPDFHandler)

	// addr := ":8080"
	// log.Println("Report API listening on", addr)
	// log.Fatal(http.ListenAndServe(addr, mux))

	// ?---
	// mux := http.NewServeMux()

	// mux.HandleFunc("/report/permission/mock", func(w http.ResponseWriter, r *http.Request) {
	// 	if r.Method != http.MethodPost && r.Method != http.MethodGet {
	// 		w.WriteHeader(http.StatusMethodNotAllowed)
	// 		return
	// 	}

	// 	q := r.URL.Query()
	// 	programs := mustInt(q.Get("programs"), 50)
	// 	companies := mustInt(q.Get("companies"), 40)
	// 	empTypes := mustInt(q.Get("emptypes"), 20)

	// 	rep := permission.MockReport(programs, companies, empTypes)

	// 	pdfBytes, err := permission.RenderPermissionPDF(rep)
	// 	if err != nil {
	// 		w.Header().Set("Content-Type", "application/json; charset=utf-8")
	// 		w.WriteHeader(http.StatusInternalServerError)
	// 		_ = json.NewEncoder(w).Encode(map[string]any{"error": err.Error()})
	// 		return
	// 	}

	// 	w.Header().Set("Content-Type", "application/pdf")
	// 	w.Header().Set("Content-Disposition", `attachment; filename="permission_report.pdf"`)
	// 	w.WriteHeader(http.StatusOK)
	// 	_, _ = w.Write(pdfBytes)
	// })

	// addr := ":8080"
	// log.Println("listening on", addr)
	// log.Fatal(http.ListenAndServe(addr, mux))

	// !!!!!!

	// report.StypePDF()

	// data := systemaccesspermission.MustSystemAccessFromFile("mock/permissionreport.json")
	// systemaccesspermission.SystemAccessPermission(data, "en")

	// !!!

	// http.HandleFunc("/preview/system-access", func(w http.ResponseWriter, r *http.Request) {

	// 	orientationStr := r.URL.Query().Get("orientation")
	// 	if orientationStr == "" {
	// 		orientationStr = "P"
	// 	}

	// 	lang := r.URL.Query().Get("lang")
	// 	if lang == "" {
	// 		lang = "th"
	// 	}

	// 	// โหลด mock data
	// 	data := systemaccesspermission.MustSystemAccessFromFile("mock/permissionreport.json")

	// 	// render เป็น bytes
	// 	pdfBytes, err := systemaccesspermission.SystemAccessPermission(data, orientationStr, lang)
	// 	if err != nil {
	// 		http.Error(w, err.Error(), http.StatusInternalServerError)
	// 		return
	// 	}

	// 	// กัน cache เวลา refresh
	// 	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	// 	w.Header().Set("Pragma", "no-cache")

	// 	// ส่งเป็น PDF ให้เปิดใน browser
	// 	w.Header().Set("Content-Type", "application/pdf")
	// 	w.Header().Set("Content-Disposition", "inline; filename=preview.pdf")
	// 	w.WriteHeader(http.StatusOK)
	// 	_, _ = w.Write(pdfBytes)
	// })

	// addr := ":8080"
	// log.Println("Preview server listening on", addr)
	// log.Fatal(http.ListenAndServe(addr, nil))

	//? Fiber ----------------

	app := fiber.New()

	app.Get("/preview/system-access", func(c *fiber.Ctx) error {
		orientationStr := c.Query("orientation", "P") // ?orientation=P|L
		lang := c.Query("lang", "th")                 // ?lang=th|en|my

		data := systemaccesspermission.MustSystemAccessFromFile("mock/permissionreport.json")

		pdfBytes, err := systemaccesspermission.SystemAccessPermission(data, orientationStr, lang)
		if err != nil {
			return c.Status(500).SendString(err.Error())
		}

		c.Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
		c.Set("Pragma", "no-cache")

		c.Set("Content-Type", "application/pdf")
		c.Set("Content-Disposition", "inline; filename=preview.pdf")

		return c.Send(pdfBytes)
	})

	app.Get("/preview/payslip", func(c *fiber.Ctx) error {
		mockName := strings.TrimSpace(c.Query("mock", "hopinn"))
		mockPath, ok := resolvePayslipMockPath(mockName)
		if !ok {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error":           "unknown payslip mock",
				"available_mocks": []string{"hopinn", "tigersoft", "bluewave", "kubota", "1", "2", "3", "4"},
			})
		}

		data := payslip.MustPayslipFromFile(mockPath)

		if companyName := strings.TrimSpace(c.Query("company_name")); companyName != "" {
			data.Report.Company.Name = companyName
		}

		logoURL := strings.TrimSpace(c.Query("logo"))
		if logoURL == "" {
			logoURL = strings.TrimSpace(c.Query("logo_url"))
		}
		if logoURL != "" {
			data.Report.Company.Logo = logoURL
		}

		if templateID := strings.TrimSpace(c.Query("template")); templateID != "" {
			data.TemplateID = templateID
		}

		pdfBytes, err := payslip.Render(data, c.Query("orientation", "P"))
		if err != nil {
			return c.Status(http.StatusInternalServerError).SendString(err.Error())
		}

		c.Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
		c.Set("Pragma", "no-cache")
		c.Set("Content-Type", "application/pdf")
		c.Set("Content-Disposition", "inline; filename=payslip_preview.pdf")

		return c.Send(pdfBytes)
	})

	app.Get("/preview/queue-ticket", func(c *fiber.Ctx) error {
		start := mustInt(c.Query("start"), 1)
		total := mustInt(c.Query("total"), 700)
		if total <= 0 {
			total = 700
		}

		pdfBytes, err := queueticket.Render(queueticket.Options{
			Title:      strings.TrimSpace(c.Query("title", "บัตรคิว")),
			Subtitle:   strings.TrimSpace(c.Query("subtitle", "Queue Ticket")),
			QueueLabel: strings.TrimSpace(c.Query("label", "หมายเลขคิว")),
			Prefix:     strings.TrimSpace(c.Query("prefix")),
			Lang:       strings.TrimSpace(c.Query("lang", "th")),
			Start:      start,
			Total:      total,
			Digits:     mustInt(c.Query("digits"), 0),
			Now:        time.Now(),
		})
		if err != nil {
			return c.Status(http.StatusInternalServerError).SendString(err.Error())
		}

		disposition := "inline"
		if isTruthy(c.Query("download")) {
			disposition = "attachment"
		}

		filename := fmt.Sprintf(
			"queue_tickets_%04d_%04d.pdf",
			start,
			start+total-1,
		)

		c.Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
		c.Set("Pragma", "no-cache")
		c.Set("Content-Type", "application/pdf")
		c.Set("Content-Disposition", fmt.Sprintf(`%s; filename="%s"`, disposition, filename))

		return c.Send(pdfBytes)
	})

	addr := ":8080"
	log.Println("Preview server listening on", addr)
	log.Fatal(app.Listen(addr))

}
func mustInt(s string, def int) int {
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 0 {
		return def
	}
	return n
}

func resolvePayslipMockPath(name string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "", "1", "hopinn":
		return "mock/payslip_hopinn.json", true
	case "2", "tigersoft":
		return "mock/payslip_tigersoft.json", true
	case "3", "bluewave":
		return "mock/payslip_bluewave.json", true
	case "4", "kubota":
		return "mock/payslip_kubota.json", true
	default:
		return "", false
	}
}

func isTruthy(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "y", "on":
		return true
	default:
		return false
	}
}

// อ่านไฟล์ + Unmarshal

func partReadFileAndUnmarshal() {
	b, err := os.ReadFile("mock/data.json")
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	// A) Unmarshal เป็น list
	var list []mockreport
	err = json.Unmarshal(b, &list)
	if err != nil {
		fmt.Println("Error unmarshaling JSON (list):", err)
		return
	}
	for _, r := range list {
		showreport(r)
	}

	// B) Unmarshal เป็น object เดี่ยว
	// หมายเหตุ: ถ้าไฟล์เป็น list จริง อันนี้จะ error (ปกติ)
	var mr mockreport
	err = json.Unmarshal(b, &mr)
	if err != nil {
		fmt.Println("Error unmarshaling JSON (single):", err)
		// ไม่ return ก็ได้ เพื่อให้ส่วนอื่นรันต่อ
		return
	}

	r := mockreport{
		Name: mr.Name,
		Year: mr.Year,
	}

	_ = mockreport{
		Name: "Updated",
		Year: 2026,
	}

	showreport(r)
}

// slice / map / struct

func partBasicTypes() {
	num := []int{1, 2, 3, 4, 5}
	num = append(num, 6)
	fmt.Println(num)

	user := map[string]any{
		"name":   "John Doe",
		"age":    30,
		"active": true,
	}
	fmt.Println(user)

	u := person{
		Name: "Alice",
		Age:  28,
	}
	fmt.Println(u.Name)
}

// JSON -> map, ดึงค่า nested, ดู type

func partJSONMapAndTypes() {
	raw := []byte(`{
		"employee": {
			"profile": {
				"name": "Peerapas"
			},
			"salary": 45000
		}
	}`)

	// 1) raw -> map
	var data map[string]any
	_ = json.Unmarshal(raw, &data)

	// 2) raw -> struct (จะไม่ได้ค่า เพราะ JSON ไม่มี name/age ตรง ๆ)
	var p person
	err := json.Unmarshal(raw, &p)
	if err != nil {
		fmt.Println("Error unmarshaling JSON:", err)
		return
	}
	fmt.Println("Unmarshaled person:", p.Name, p.Age)

	// โชว์ map
	fmt.Println(data)
	fmt.Println(data["employee"])

	// ตรงนี้ของเดิม "employee.name" ไม่มีจริง (ชื่ออยู่ใน employee.profile.name)
	// เลยทำแบบถูก path:
	emp := data["employee"].(map[string]any)
	profile := emp["profile"].(map[string]any)
	fmt.Println(profile["name"])

	salary := emp["salary"]
	fmt.Printf("%T\n", salary)

	// อีก JSON ชุด
	rawtet := []byte(`{
	  "company": {
	    "name": "TigerSoft",
	    "location": "Bangkok"
	  }
	}`)

	var data2 map[string]any
	_ = json.Unmarshal(rawtet, &data2)

	com := data2["company"].(map[string]any)
	fmt.Println(com["name"])
}

// getByPath แบบ split "."

func partGetByPath() {
	raw := []byte(`{
		"employee": {
			"profile": { "name": "Peerapas" },
			"salary": 45000
		}
	}`)

	var result map[string]any
	_ = json.Unmarshal(raw, &result)

	path := "employee.profile.name"
	paths := strings.Split(path, ".")

	var current any = result
	for _, p := range paths {
		m := current.(map[string]any)
		current = m[p]
	}

	fmt.Println(current)
}

// HTTP Server

func partHTTPServer() {
	http.HandleFunc("/report", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Report Service is running")
	})

	fmt.Println("Starting server at :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("Error starting server:", err)
	}
}

// RenderText + RenderPDF

func partReportRender() {
	templateJSON := `{
		"title": "Employee Report",
		"sections": [
			{ "type": "kv", "label": "ชื่อ", "key": "employee.profile.name" },
			{ "type": "kv", "label": "เงินเดือน", "key": "employee.salary" }
		]
	}`

	dataJSON := `{
		"employee": {
			"profile": { "name": "Peerapas" },
			"salary": 45000
		}
	}`

	t := mustTemplate(templateJSON)
	d := mustData(dataJSON)

	out := report.RenderText(t, d)
	fmt.Println(out)

	tf := mustTemplateFromFile("mock/template.json")
	df := mustDataFromFile("mock/data.json")

	outf := report.RenderText(tf, df)
	fmt.Println(outf)

	pdfBytes, err := report.RenderPDF(tf, df)
	if err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile("report.pdf", pdfBytes, 0644); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Generated: report.pdf")
}

// chromedp HTML -> PDF

func partChromeDP() {
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(),
		append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.Flag("headless", true),
		)...,
	)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	html := `<!doctype html>
<html>
<head>
  <meta charset="utf-8" />
  <style>
    body { font-family: Arial; margin: 24px; }
    .box { border: 1px solid #333; padding: 12px; }
  </style>
</head>
<body>
  <h1>Employee Report</h1>
  <div class="box">Hello PDF</div>
</body>
</html>`

	var pdfBuf []byte

	err := chromedp.Run(ctx,
		chromedp.Navigate("about:blank"),
		chromedp.ActionFunc(func(ctx context.Context) error {
			root, err := dom.GetDocument().Do(ctx)
			if err != nil {
				return err
			}
			return dom.SetOuterHTML(root.NodeID, html).Do(ctx)
		}),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			pdfBuf, _, err = page.PrintToPDF().
				WithPrintBackground(true).
				WithPaperWidth(8.27).
				WithPaperHeight(11.69).
				Do(ctx)
			return err
		}),
	)
	if err != nil {
		log.Fatal(err)
	}

	if err := os.WriteFile("out.pdf", pdfBuf, 0644); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Generated: out.pdf")
}

// TYPES + HELPERS
type mockreport struct {
	Name string `json:"name"`
	Year int    `json:"year"`
}

func showreport(r mockreport) {
	fmt.Println("Report Name:", r.Name)
	fmt.Println("Report Year:", r.Year)
}

type person struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func mustTemplate(row string) report.Template {
	var t report.Template
	err := json.Unmarshal([]byte(row), &t)
	if err != nil {
		panic(err)
	}
	return t
}

func mustData(row string) map[string]any {
	var data map[string]any
	err := json.Unmarshal([]byte(row), &data)
	if err != nil {
		panic(err)
	}
	return data
}

func mustTemplateFromFile(path string) report.Template {
	b, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}

	var t report.Template
	if err := json.Unmarshal(b, &t); err != nil {
		log.Fatal(err)
	}
	return t
}

func mustDataFromFile(path string) map[string]any {
	b, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}

	var d map[string]any
	if err := json.Unmarshal(b, &d); err != nil {
		log.Fatal(err)
	}
	return d
}

type RenderPDFRequest struct {
	Template report.Template `json:"template"`
	Data     map[string]any  `json:"data"`
	FileName string          `json:"fileName,omitempty"` // optional
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func renderPDFHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "method not allowed"})
		return
	}

	// จำกัดขนาด body กันยิงใหญ่เกิน (เช่น 5MB)
	r.Body = http.MaxBytesReader(w, r.Body, 5<<20)

	var req RenderPDFRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields() // กันพิมพ์ field ผิดแล้วหลุดเงียบ ๆ
	if err := dec.Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid json: " + err.Error()})
		return
	}

	// กัน template ว่าง
	if req.Template.Title == "" {
		req.Template.Title = "Report"
	}

	pdfBytes, err := report.RenderPDF(req.Template, req.Data)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "render pdf failed: " + err.Error()})
		return
	}

	fn := req.FileName
	if fn == "" {
		fn = "report.pdf"
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", `attachment; filename="`+fn+`"`)
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(pdfBytes)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":   true,
		"time": time.Now().Format(time.RFC3339),
	})
}
