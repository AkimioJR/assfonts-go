package font

import (
	"os/user"
	"path/filepath"
)

func getDefaultFontPaths() []string {
	var paths []string
	macPaths := []string{
		"/Library/Fonts",
		"/Network/Library/Fonts",
		"/System/Library/Fonts",
	}
	for _, p := range macPaths {
		if isDir(p) {
			paths = append(paths, p)
		}
	}
	if usr, err := user.Current(); err == nil {
		userFont := filepath.Join(usr.HomeDir, "Library", "Fonts")
		if isDir(userFont) {
			paths = append(paths, userFont)
		}
	}
	return paths
}
