package report

import (
	"fmt"
	"strings"
)

func RenderText(t Template, data map[string]any) string {
	var b strings.Builder

	b.WriteString("=== " + t.Title + " ===\n")

	for _, s := range t.Sections {
		switch s.Type {
		case "kv":
			v := Getbypath(data, s.Key)
			b.WriteString(fmt.Sprintf("%s: %v\n", s.Label, v))
		default:
			b.WriteString("Unknown section type: " + s.Type + "\n")
		}
	}

	return b.String()
}
