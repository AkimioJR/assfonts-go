package font

import (
	"encoding/json"
	"fmt"
	"os"
)

type FontDataBase struct {
	lib          *FreeTypeLibrary      // FreeType 库实例
	internalLib  bool                  // 是否内部创建 FreeType 库实例
	FontListInDB map[string][]FontInfo // path -> []FontInfo
}

// 创建一个新的 FontDataBase 对象
// 如果传入 FreeTypeLibrary 为 nil，则会创建一个内部的 FreeTypeLibrary 实例
// 如果传入的 FreeTypeLibrary 不为 nil，则使用该实例
// 注意：如果传入的 FreeTypeLibrary 是内部创建的，需要调用 Close() 方法
func NewFontDataBase(lib *FreeTypeLibrary) (*FontDataBase, error) {
	var db = FontDataBase{
		lib:          lib,
		internalLib:  false,
		FontListInDB: make(map[string][]FontInfo),
	}
	if lib == nil {
		lib, err := NewFreeTypeLibrary()
		if err != nil {
			return nil, fmt.Errorf("create FontDataBase faild due to create internal FreeTypeLibrary: %w", err)
		}
		db.lib = lib
		db.internalLib = true
	}
	return &db, nil
}

// 关闭 FontDataBase
// 如果 FontDataBase 是内部创建的 FreeTypeLibrary 实例，则会关闭该实例
// 如果传入的 FreeTypeLibrary 是外部创建的，则不会关闭该实例
func (fdb *FontDataBase) Close() error {
	if fdb.internalLib && fdb.lib != nil {
		err := fdb.lib.Close()
		fdb.lib = nil
		if err != nil {
			return fmt.Errorf("failed to close FontDataBase due to FreeType library close error: %v", err)
		}
		return nil
	}
	return nil
}

func (fdb *FontDataBase) BuildDB(fontsDirs []string, withSystemFontPath bool, ignoreError bool) error {
	fontPaths, err := findFontFiles(fontsDirs, withSystemFontPath)
	if err != nil {
		return fmt.Errorf("failed to find font files: %w", err)
	}

	for _, fontPath := range fontPaths {
		fontInfos, err := fdb.lib.ParseFont(fontPath, ignoreError)
		if err != nil {
			if ignoreError {
				continue
			}
			return fmt.Errorf("failed to parse font %s: %w", fontPath, err)
		}
		fdb.FontListInDB[fontPath] = fontInfos
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
