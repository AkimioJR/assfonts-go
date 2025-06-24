package font

import (
	"encoding/json"
	"fmt"
	"github/Akimio521/assfonts-go/ass"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type FontDataBase struct {
	BigMemoryMode bool
	lib           *FreeTypeLibrary          // FreeType 库实例
	internalLib   bool                      // 是否内部创建 FreeType 库实例
	data          map[string][]FontFaceInfo // path -> []FontFaceInfo
	fontData      map[string][]byte
}

// 创建一个新的 FontDataBase 对象
// 如果传入 FreeTypeLibrary 为 nil，则会创建一个内部的 FreeTypeLibrary 实例
// 如果传入的 FreeTypeLibrary 不为 nil，则使用该实例
// 注意：如果传入的 FreeTypeLibrary 是内部创建的，需要调用 Close() 方法
func NewFontDataBase(lib *FreeTypeLibrary) (*FontDataBase, error) {
	var db = FontDataBase{
		lib:         lib,
		internalLib: false,
		data:        make(map[string][]FontFaceInfo),
		fontData:    make(map[string][]byte),
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

func (db *FontDataBase) BuildDB(fontsDirs []string, withSystemFontPath bool, fn CheckErrFn) error {
	fontPaths, err := findFontFiles(fontsDirs, withSystemFontPath)
	if err != nil {
		return fmt.Errorf("failed to find font files: %w", err)
	}

	for _, fontPath := range fontPaths {
		fontFaceInfos, err := db.lib.ParseFont(fontPath, fn)
		if err != nil {
			if fn != nil { // 仅提示错误，不终止程序
				fn(NewWarningMsg("failed to parse font %s: %s", fontPath, err))
			}
			continue
		}
		if db.BigMemoryMode {
			data, err := os.ReadFile(fontPath)
			if err != nil {
				if fn != nil { // 仅提示错误，不终止程序
					fn(NewWarningMsg("failed to read font file %s: %s", fontPath, err))
				}
				continue
			}
			db.fontData[fontPath] = data
		}
		db.data[fontPath] = fontFaceInfos
	}
	return nil
}

func (db *FontDataBase) SaveDB(dbPath string) error {

	data, err := json.MarshalIndent(db.data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal font data: %w", err)
	}
	if err := os.WriteFile(dbPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write font database to %s: %w", dbPath, err)
	}
	return nil
}

// LoadDB 加载字体数据库
func (db *FontDataBase) LoadDB(dbPath string) error {
	data, err := os.ReadFile(dbPath)
	if err != nil {
		return fmt.Errorf(`cannot read fonts database: "%s"`, dbPath)
	}

	if err := json.Unmarshal(data, &db.data); err != nil {
		return fmt.Errorf(`cannot load fonts database: "%s"`, dbPath)
	}
	if db.BigMemoryMode {
		for path := range db.data {
			fontData, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf(`cannot read font file "%s" for big memory mode: %w`, path, err)
			}
			db.fontData[path] = fontData
		}
	}
	return nil
}

func (db *FontDataBase) getFontData(path string) ([]byte, error) {
	if db.BigMemoryMode {
		if data, ok := db.fontData[path]; ok {
			return data, nil
		}
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read font file %s: %w", path, err)
	}
	if db.BigMemoryMode {
		db.fontData[path] = data
	}
	return data, nil
}

// 读取 ass.ASSParser 中的所有字形，在 FontDataBase 中查找对应字体的路径，通过 CreatSubfont 子集化后返回子集化后的字体文件
func (db *FontDataBase) Subset(ap *ass.ASSParser, opts ...SubsetOption) (map[string][]byte, error) {
	config := &subsetConfig{
		concurrent: false,
		checkGlyph: false,
		fn:         nil,
	}
	for _, opt := range opts {
		opt(config)
	}

	subsetFontInfos, err := db.parseSubsetFontInfos(ap, config.fn)
	if err != nil {
		return nil, fmt.Errorf("parse sub set info failed: %w", err)
	}

	var subFontDatas map[string][]byte
	if config.concurrent {
		subFontDatas, err = db.subsetConcurrent(subsetFontInfos, config)
	} else {
		subFontDatas, err = db.subsetSequential(subsetFontInfos, config)
	}

	if err != nil {
		return nil, err
	}
	if len(subFontDatas) == 0 {
		return nil, ErrEmptySubsetData
	}
	return subFontDatas, nil
}

func (db *FontDataBase) subsetSequential(subsetFontInfos []SubsetFontInfo, c *subsetConfig) (map[string][]byte, error) {
	subFontDatas := make(map[string][]byte)

	for _, sfi := range subsetFontInfos {
		key, subFontData, err := db.subset(&sfi, c)
		if err != nil {
			if c.fn != nil && !c.fn(err) {
				return nil, err
			}
			continue
		}
		subFontDatas[key] = subFontData
	}

	return subFontDatas, nil
}
func (db *FontDataBase) subsetConcurrent(subsetFontInfos []SubsetFontInfo, c *subsetConfig) (map[string][]byte, error) {
	type result struct {
		key  string
		data []byte
	}

	results := make(chan result, len(subsetFontInfos))

	var wg sync.WaitGroup
	wg.Add(len(subsetFontInfos))
	for _, sfi := range subsetFontInfos {
		go func() {
			defer wg.Done()
			key, subFontData, err := db.subset(&sfi, c)
			if err != nil && c.fn != nil {
				c.fn(err)
				return
			}
			results <- result{key, subFontData}
		}()
	}

	// 关闭结果通道当所有工作完成时
	go func() {
		wg.Wait()
		close(results)
	}()

	subFontDatas := make(map[string][]byte, len(subsetFontInfos))
	for r := range results {
		subFontDatas[r.key] = r.data
	}

	return subFontDatas, nil
}

func (db *FontDataBase) subset(sfi *SubsetFontInfo, c *subsetConfig) (string, []byte, error) {
	fontData, err := db.getFontData(sfi.Source.Path)
	if err != nil {
		return "", nil, fmt.Errorf("failed to get font data for %s: %w", sfi.Source.Path, err)
	}
	if c.fn != nil && c.checkGlyph {
		go func() {
			err := db.CheckGlyph(fontData, &sfi.Source, sfi.Codepoints, &sfi.FontsDesc)
			if err != nil {
				c.fn(err)
			}
		}()
	}
	subFontData, err := db.CreatSubfont(sfi, fontData)
	if err != nil {
		return "", nil, err
	}
	return sfi.FontsDesc.FontName + filepath.Ext(sfi.Source.Path), subFontData, nil
}

func (db *FontDataBase) parseSubsetFontInfos(ap *ass.ASSParser, fn CheckErrFn) ([]SubsetFontInfo, error) {
	subsetFontInfos := make([]SubsetFontInfo, 0, len(ap.FontSets))

	for fontDesc, fontSet := range ap.FontSets {
		// fmt.Println(fontDesc)
		codepointSet := make(ass.CodepointSet)
		fontPath, errValue, err := db.FindFont(&fontDesc, fontSet)
		if err != nil {
			return nil, err
		}
		if fn != nil {
			fn(NewInfoMsg(`"%s" (%d,%d) ---> "%s"[%d], error value: %d`, fontDesc.FontName, fontDesc.Bold, fontDesc.Italic, fontPath.Path, fontPath.Index, errValue))
		}

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
	ttfExts = []string{".ttf", ".ttc"}
	otfExts = []string{".otf", ".otc"}
)

func (db *FontDataBase) FindFont(fontDesc *ass.FontDesc, fontSet ass.CodepointSet) (*FontFaceLocation, int, *ErrMissingFontFaceFound) {
	targetName := strings.ToLower(fontDesc.FontName)

	find := func(acceptExts []string) (*FontFaceLocation, int) {
		minErr := math.MaxInt // 当前最小误差
		var best = &FontFaceLocation{}

		for path, fontFaceInfos := range db.data {
			if !contains(acceptExts, strings.ToLower(filepath.Ext(path))) {
				continue
			}
			for _, fontFaceInfo := range fontFaceInfos {
				var currentErr int = math.MaxInt // 当前误差
				found := false

				if contains(fontFaceInfo.Name.FullNames, targetName) || contains(fontFaceInfo.Name.PSNames, targetName) { // 精确匹配，全名一致
					found = true
					currentErr = 0
				} else if contains(fontFaceInfo.Name.FamilyNames, targetName) { // 检查家族名
					currentErr = abs(int(fontDesc.Bold)-int(fontFaceInfo.Weight)) + abs(int(fontDesc.Italic)-int(fontFaceInfo.Slant))
					found = true
				}

				if !found {
					continue
				}
				if currentErr < minErr {
					minErr = currentErr
					best = &fontFaceInfo.Source
				}
				if currentErr == 0 {
					break
				}
			}
		}
		return best, minErr
	}

	var bestSource *FontFaceLocation = nil
	var errValue = math.MaxInt
	ttfSource, ttfErr := find(ttfExts)
	otfSource, otfErr := find(otfExts)

	switch {
	case ttfErr == 0:
		bestSource = ttfSource
		errValue = 0
	case otfErr == 0:
		bestSource = otfSource
		errValue = 0
	case ttfErr <= otfErr:
		bestSource = ttfSource
		errValue = ttfErr
	case otfErr < ttfErr:
		bestSource = otfSource
		errValue = otfErr
	case ttfErr < math.MaxInt:
		bestSource = ttfSource
		errValue = ttfErr
	case otfErr < math.MaxInt:
		bestSource = otfSource
		errValue = otfErr
	}

	if bestSource == nil {
		return nil, 0, NewErrMissingFontFaceFound(*fontDesc)
	}
	return bestSource, errValue, nil
}
