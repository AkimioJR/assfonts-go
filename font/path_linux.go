package font

import (
	"os"
	"path/filepath"
)

func getDefaultFontPaths() []string {
	var paths []string
	linuxPaths := []string{
		"/usr/share/fonts",
	}
	for _, p := range linuxPaths {
		if isDir(p) {
			paths = append(paths, p)
		}
	}
	if home := os.Getenv("HOME"); home != "" {
		userFont := filepath.Join(home, ".local", "share", "fonts")
		if isDir(userFont) {
			paths = append(paths, userFont)
		}
	}
	return paths
}
