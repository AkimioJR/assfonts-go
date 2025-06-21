package font

import (
	"encoding/json"
	"fmt"
	"os"
)

type FontDataBase struct {
	FontListInDB map[string][]FontInfo // path -> []FontInfo
}

func NewFontDataBase() *FontDataBase {
	return &FontDataBase{
		FontListInDB: make(map[string][]FontInfo),
	}
}

func (fp *FontDataBase) BuildDB(lib *FreeTypeLibrary, fontsDirs []string, withDefault bool, ignoreError bool) error {
	fontPaths, err := findFontFiles(fontsDirs, withDefault)
	if err != nil {
		return fmt.Errorf("failed to find font files: %w", err)
	}
	// lib, err := NewFreeTypeLibrary()
	// if err != nil {
	// 	return fmt.Errorf("failed to create FreeType library: %w", err)
	// }
	// defer lib.Close()

	for _, fontPath := range fontPaths {
		fontInfos, err := lib.ParseFont(fontPath, ignoreError)
		if err != nil && !ignoreError {
			return fmt.Errorf("failed to parse font %s: %w", fontPath, err)
		}
		fp.FontListInDB[fontPath] = fontInfos
	}
	return nil
}

func (fp *FontDataBase) SaveDB(dbPath string) error {
	var fis []FontInfo
	for _, fontInfos := range fp.FontListInDB {
		fis = append(fis, fontInfos...)
	}
	data, err := json.MarshalIndent(fis, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal font data: %w", err)
	}
	if err := os.WriteFile(dbPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write font database to %s: %w", dbPath, err)
	}
	return nil
}

// LoadDB 加载字体数据库
func (fdb *FontDataBase) LoadDB(dbPath string) error {
	data, err := os.ReadFile(dbPath)
	if err != nil {
		return fmt.Errorf(`cannot read fonts database: "%s"`, dbPath)
	}
	fdb.FontListInDB = make(map[string][]FontInfo)

	var fis []FontInfo
	if err := json.Unmarshal(data, &fis); err != nil {
		return fmt.Errorf(`cannot load fonts database: "%s"`, dbPath)
	}

	for _, fi := range fis {
		if fdb.FontListInDB[fi.Path] == nil {
			fdb.FontListInDB[fi.Path] = make([]FontInfo, 0)
		}
		fdb.FontListInDB[fi.Path] = append(fdb.FontListInDB[fi.Path], fi)
	}
	return nil
}
