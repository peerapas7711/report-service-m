package report

type Template struct {
	Title    string    `json:"title"`
	Sections []Section `json:"sections"`
}

type Section struct {
	Type  string `json:"type"`  // ตอนนี้รองรับแค่ "kv"
	Label string `json:"label"` // ชื่อที่จะแสดง
	Key   string `json:"key"`   // path ใน data เช่น "employee.salary"
}
