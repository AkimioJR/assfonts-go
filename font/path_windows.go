package font

import (
	"os"
	"path/filepath"
)

func getDefaultFontPaths() []string {
	var paths []string
	windir := os.Getenv("WINDIR")
	if windir != "" {
		fontDir := filepath.Join(windir, "Fonts")
		if isDir(fontDir) {
			paths = append(paths, fontDir)
		}
	}
	return paths
}
