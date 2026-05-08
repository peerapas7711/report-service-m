package httpserver

import (
	"os"
	"path/filepath"
)

func projectFile(path string) string {
	if _, err := os.Stat(path); err == nil {
		return path
	}

	wd, err := os.Getwd()
	if err != nil {
		return path
	}

	root := findProjectRoot(wd)
	if root == "" {
		return path
	}

	return filepath.Join(root, path)
}

func findProjectRoot(start string) string {
	dir := filepath.Clean(start)

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}
