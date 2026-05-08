package payslip

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"strings"
	"sync"
)

//go:embed templates/*.json
var layoutTemplateFS embed.FS

type LayoutTemplate struct {
	ID       string          `json:"id"`
	Aliases  []string        `json:"aliases,omitempty"`
	Page     LayoutPage      `json:"page"`
	Elements []LayoutElement `json:"elements"`
}

type LayoutPage struct {
	Margin             float64 `json:"margin"`
	DefaultOrientation string  `json:"default_orientation,omitempty"`
}

type LayoutElement struct {
	Type       string          `json:"type"`
	X          float64         `json:"x,omitempty"`
	Y          float64         `json:"y,omitempty"`
	W          float64         `json:"w,omitempty"`
	H          float64         `json:"h,omitempty"`
	X2         float64         `json:"x2,omitempty"`
	Y2         float64         `json:"y2,omitempty"`
	Radius     float64         `json:"radius,omitempty"`
	Corners    string          `json:"corners,omitempty"`
	Source     string          `json:"source,omitempty"`
	Text       string          `json:"text,omitempty"`
	Align      string          `json:"align,omitempty"`
	ValueAlign string          `json:"value_align,omitempty"`
	Padding    float64         `json:"padding,omitempty"`
	LineHeight float64         `json:"line_height,omitempty"`
	RowGap     float64         `json:"row_gap,omitempty"`
	MinRowGap  float64         `json:"min_row_gap,omitempty"`
	MaxRowGap  float64         `json:"max_row_gap,omitempty"`
	MaxRows    int             `json:"max_rows,omitempty"`
	LabelRatio float64         `json:"label_ratio,omitempty"`
	LineWidth  float64         `json:"line_width,omitempty"`
	Multiline  bool            `json:"multiline,omitempty"`
	Stacked    bool            `json:"stacked,omitempty"`
	ShowBorder *bool           `json:"show_border,omitempty"`
	Font       *FontSpec       `json:"font,omitempty"`
	LabelFont  *FontSpec       `json:"label_font,omitempty"`
	ValueFont  *FontSpec       `json:"value_font,omitempty"`
	DrawColor  *ColorSpec      `json:"draw_color,omitempty"`
	FillColor  *ColorSpec      `json:"fill_color,omitempty"`
	TextColor  *ColorSpec      `json:"text_color,omitempty"`
	Children   []LayoutElement `json:"children,omitempty"`
}

type FontSpec struct {
	Lang  string  `json:"lang,omitempty"`
	Style string  `json:"style,omitempty"`
	Size  float64 `json:"size,omitempty"`
}

type ColorSpec struct {
	R int `json:"r"`
	G int `json:"g"`
	B int `json:"b"`
}

var (
	layoutTemplates     map[string]LayoutTemplate
	layoutTemplateAlias map[string]string
	layoutTemplateErr   error
	layoutTemplateOnce  sync.Once
)

func loadLayoutTemplate(templateID string) (LayoutTemplate, error) {
	layoutTemplateOnce.Do(initLayoutTemplates)
	if layoutTemplateErr != nil {
		return LayoutTemplate{}, layoutTemplateErr
	}

	normalized := normalizeTemplateID(templateID)
	if tmpl, ok := layoutTemplates[normalized]; ok {
		return tmpl, nil
	}

	tmpl, ok := layoutTemplates[defaultTemplate]
	if !ok {
		return LayoutTemplate{}, fmt.Errorf("default payslip template %q not found", defaultTemplate)
	}

	return tmpl, nil
}

func normalizeTemplateID(value string) string {
	layoutTemplateOnce.Do(initLayoutTemplates)

	key := strings.ToLower(strings.TrimSpace(value))
	if key == "" {
		return defaultTemplate
	}

	if layoutTemplateAlias != nil {
		if normalized, ok := layoutTemplateAlias[key]; ok {
			return normalized
		}
	}

	return defaultTemplate
}

func initLayoutTemplates() {
	layoutTemplates = make(map[string]LayoutTemplate)
	layoutTemplateAlias = make(map[string]string)

	paths, err := fs.Glob(layoutTemplateFS, "templates/*.json")
	if err != nil {
		layoutTemplateErr = err
		return
	}

	for _, path := range paths {
		raw, err := layoutTemplateFS.ReadFile(path)
		if err != nil {
			layoutTemplateErr = err
			return
		}

		var tmpl LayoutTemplate
		if err := json.Unmarshal(raw, &tmpl); err != nil {
			layoutTemplateErr = fmt.Errorf("parse %s: %w", path, err)
			return
		}

		tmpl.ID = strings.ToLower(strings.TrimSpace(tmpl.ID))
		if tmpl.ID == "" {
			layoutTemplateErr = fmt.Errorf("template %s missing id", path)
			return
		}

		layoutTemplates[tmpl.ID] = tmpl
		layoutTemplateAlias[tmpl.ID] = tmpl.ID
		for _, alias := range tmpl.Aliases {
			normalizedAlias := strings.ToLower(strings.TrimSpace(alias))
			if normalizedAlias == "" {
				continue
			}
			layoutTemplateAlias[normalizedAlias] = tmpl.ID
		}
	}

	if _, ok := layoutTemplates[defaultTemplate]; !ok {
		layoutTemplateErr = fmt.Errorf("default payslip template %q not found", defaultTemplate)
	}
}
