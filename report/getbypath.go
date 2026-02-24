package report

import "strings"

func Getbypath(data map[string]any, path string) any {
	if path == "" {
		return nil
	}

	parts := strings.Split(path, ".")
	var current any = data
	for _, p := range parts {
		m, ok := current.(map[string]any)
		if !ok {
			return nil
		}

		next, exists := m[p]
		if !exists {
			return nil
		}

		current = next
	}

	return current
}
