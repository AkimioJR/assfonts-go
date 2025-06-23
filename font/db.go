package font

import (
	"encoding/json"
	"fmt"
	"github/Akimio521/assfonts-go/ass"
	"math"
	"os"
	"path/filepath"
	"strings"
)

type FontDataBase struct {
	lib          *FreeTypeLibrary          // FreeType 库实例
	internalLib  bool                      // 是否内部创建 FreeType 库实例
	FontListInDB map[string][]FontFaceInfo // path -> []FontFaceInfo
}

// 创建一个新的 FontDataBase 对象
// 如果传入 FreeTypeLibrary 为 nil，则会创建一个内部的 FreeTypeLibrary 实例
// 如果传入的 FreeTypeLibrary 不为 nil，则使用该实例
// 注意：如果传入的 FreeTypeLibrary 是内部创建的，需要调用 Close() 方法
func NewFontDataBase(lib *FreeTypeLibrary) (*FontDataBase, error) {
	var db = FontDataBase{
		lib:          lib,
		internalLib:  false,
		FontListInDB: make(map[string][]FontFaceInfo),
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
	var fis []FontFaceInfo
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
	fdb.FontListInDB = make(map[string][]FontFaceInfo)

	var fis []FontFaceInfo
	if err := json.Unmarshal(data, &fis); err != nil {
		return fmt.Errorf(`cannot load fonts database: "%s"`, dbPath)
	}

	for _, fi := range fis {
		if fdb.FontListInDB[fi.Source.Path] == nil {
			fdb.FontListInDB[fi.Source.Path] = make([]FontFaceInfo, 0)
		}
		fdb.FontListInDB[fi.Source.Path] = append(fdb.FontListInDB[fi.Source.Path], fi)
	}
	return nil
}

// 读取 ass.ASSParser 中的所有字形，在 FontDataBase 中查找对应字体的路径，通过 CreatSubfont 子集化后返回子集化后的字体文件
func (db *FontDataBase) Subset(ap *ass.ASSParser) (map[string][]byte, error) {
	subsetFontInfos, err := db.parseSubsetFontInfos(ap)
	if err != nil {
		return nil, fmt.Errorf("parse sub set info failed: %w", err)
	}

	subFontDatas := make(map[string][]byte)

	for _, sfi := range subsetFontInfos {
		subFontData, err := CreatSubfont(&sfi)
		if err != nil {
			return nil, err
		}
		subFontDatas[sfi.FontsDesc.FontName+filepath.Ext(sfi.Source.Path)] = subFontData
	}
	return subFontDatas, nil
}

func (db *FontDataBase) parseSubsetFontInfos(ap *ass.ASSParser) ([]SubsetFontInfo, error) {
	subsetFontInfos := make([]SubsetFontInfo, 0, len(ap.FontSets))

	for fontDesc, fontSet := range ap.FontSets {
		// fmt.Println(fontDesc)
		codepointSet := make(ass.CodepointSet)
		fontPath, err := db.FindFont(&fontDesc, fontSet)
		if err != nil {
			return nil, fmt.Errorf(`missing the font face "%s" (%d,%d): %w`, fontDesc.FontName, fontDesc.Bold, fontDesc.Italic, err)
		}
		// err = fs.CheckGlyph(db.lib, fontPath, fontSet, strings.ToLower(fontDesc.FontName), fontDesc.Bold, fontDesc.Italic)
		// if err != nil {
		// 	return fmt.Errorf(`check font face "%s" (%d,%d) of %s error: %w`, fontDesc.FontName, fontDesc.Bold, fontDesc.Italic, fontPath.Path, err)
		// }

		for wch := range fontSet {
			codepointSet[wch] = struct{}{}
		}
		for _, ch := range additionalCodepoints {
			codepointSet[ch] = struct{}{}
		}
		subsetFontInfos = append(subsetFontInfos, SubsetFontInfo{
			FontsDesc:  fontDesc,
			Codepoints: codepointSet,
			Source:     *fontPath,
		})
	}
	return subsetFontInfos, nil
}

var (
	ttfExts = map[string]struct{}{".ttf": {}, ".ttc": {}}
	otfExts = map[string]struct{}{".otf": {}, ".otc": {}}
)

func (db *FontDataBase) FindFont(fontDesc *ass.FontDesc, fontSet ass.CodepointSet) (*FontFaceLocation, error) {
	fontname := strings.ToLower(fontDesc.FontName)

	find := func(acceptExts map[string]struct{}) (*FontFaceLocation, int) {
		minErr := math.MaxInt // 当前最小误差
		var best = &FontFaceLocation{}

		for path, fontInfos := range db.FontListInDB {
			if minErr == 0 {
				break
			}

			ext := strings.ToLower(filepath.Ext(path))
			if _, ok := acceptExts[ext]; !ok {
				continue
			}
			for _, fontInfo := range fontInfos {
				var currentErr int // 当前误差
				if contains(fontInfo.Families, fontname) {
					currentErr = abs(int(fontDesc.Bold)-int(fontInfo.Weight)) + abs(int(fontDesc.Italic)-int(fontInfo.Slant))
					// fmt.Println("find", fontname, "in", fontInfo.Families, "with score: ", score, "path: ", path)
				} else if contains(fontInfo.FullNames, fontname) || contains(fontInfo.PSNames, fontname) {
					currentErr = 0
				} else {
					continue
				}
				if currentErr < minErr {
					minErr = currentErr
					best.Path = path
					best.Index = fontInfo.Source.Index
				}
				if currentErr == 0 {
					break
				}
			}
		}
		return best, minErr
	}

	var bestSource *FontFaceLocation
	ttfSource, ttfErr := find(ttfExts)
	otcSource, otcErr := find(otfExts)

	// 优先 ttf/ttc
	if ttfErr < math.MaxInt || otcErr < math.MaxInt {
		if ttfErr <= otcErr {
			bestSource = ttfSource
		} else {
			bestSource = otcSource
		}
	}

	if bestSource == nil {
		return nil, fmt.Errorf("no valid font found for %s", fontDesc.FontName)
	}
	// fmt.Printf("find font \"%s\" (%d,%d) ---> \"%s\"[%d]\n", fontDesc.FontName, fontDesc.Bold, fontDesc.Italic, bestPath.Path, bestPath.Index)
	return bestSource, nil
}
