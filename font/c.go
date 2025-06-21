package font

/*
#cgo CFLAGS: -I${SRCDIR}/clib/arm64-osx/include
#cgo LDFLAGS: -L${SRCDIR}/clib/arm64-osx/lib -lfreetype -lbz2 -lbrotlidec -lbrotlicommon -lbrotlienc -lz -lpng16

#include <ft2build.h>
#include FT_FREETYPE_H // <freetype/freetype.h>
#include FT_TYPE1_TABLES_H // <freetype/t1tables.h>
#include FT_SFNT_NAMES_H // <freetype/ttnameid.h>
#include FT_TRUETYPE_TABLES_H // <freetype/tttables.h>
#include FT_SFNT_NAMES_H // <freetype/ftsnames.h>
*/
import "C"

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"unsafe"
)

type FreeTypeLibrary struct {
	ptr C.FT_Library // FreeType库句柄指针
}

// 初始化FreeType库
func NewFreeTypeLibrary() (*FreeTypeLibrary, error) {

	var lib C.FT_Library

	// 调用C函数初始化库
	errCode := C.FT_Init_FreeType(&lib)
	if errCode != 0 {
		return nil, fmt.Errorf("init error, error code: %d", int(errCode))
	}

	return &FreeTypeLibrary{ptr: lib}, nil
}

// 关闭FreeType库
func (lib *FreeTypeLibrary) Close() error {
	if lib.ptr == nil {
		return nil
	}
	errCode := C.FT_Done_FreeType(lib.ptr)
	lib.ptr = nil // 防止重复释放
	if errCode != 0 {
		return fmt.Errorf("release error, error code: %d", int(errCode))
	}
	return nil
}

// 解析字体文件
func (lib *FreeTypeLibrary) ParseFont(fontPath string, ignoreError bool) ([]FontInfo, error) {
	fileInfo, err := os.Stat(fontPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("font file does not exist: %s", fontPath)
		}
		return nil, fmt.Errorf("error accessing font file: %s, error: %w", fontPath, err)
	}

	cFontPath := C.CString(fontPath)
	defer C.free(unsafe.Pointer(cFontPath))

	var face C.FT_Face

	// 调用C函数解析字体
	errCode := C.FT_New_Face(lib.ptr, cFontPath, 0, &face) // 0表示加载第一个字体
	if errCode != 0 {
		return nil, fmt.Errorf("parse font error, error code: %d", int(errCode))
	}
	// defer C.FT_Done_Face(face) // 解析完成后释放face资源

	facesNum := int(face.num_faces)            // 获取字体文件中的字体数量
	fontInfos := make([]FontInfo, 0, facesNum) // 初始化字体信息切片
	for idx := range facesNum {
		// fmt.Println(idx, facesNum)
		fontInfo, err := lib.parseFace(cFontPath, face, idx)
		if err != nil && ignoreError {
			return nil, fmt.Errorf("error parsing face at index %d: %w", idx, err)
		}
		if fontInfo != nil {
			fontInfo.Index = int64(idx)                             // 设置字体索引
			fontInfo.LastWriteTime = formatTime(fileInfo.ModTime()) // 设置最后写入时间
			fontInfo.Path = fontPath                                // 设置字体文件路径
			fontInfos = append(fontInfos, *fontInfo)                // 添加到字体信息切片
		}
	}

	if len(fontInfos) == 0 {
		return nil, ErrNoValidFontFace
	}
	return fontInfos, nil
}

func (lib *FreeTypeLibrary) parseFace(cFontPath *C.char, face C.FT_Face, idx int) (*FontInfo, error) {
	var (
		openArgs  C.FT_Open_Args = newOpenArgs(cFontPath)
		families  []string       = make([]string, 0) // 字体家族
		fullnames []string       = make([]string, 0) // 字体全名
		psnames   []string       = make([]string, 0) // PostScript字体名称
	)

	if idx != 0 {
		if errCode := C.FT_Open_Face(lib.ptr, &openArgs, C.FT_Long(idx), &face); errCode != 0 {
			return nil, fmt.Errorf("error opening face at index %d with error code %d", idx, errCode)
		}
	}

	if C.FT_Has_PS_Glyph_Names(face) != 0 { // 检查是否有PostScript字形名称
		fontInfo := (C.PS_FontInfo)(C.malloc(C.sizeof_PS_FontInfo))
		defer C.free(unsafe.Pointer(fontInfo))

		if C.FT_Get_PS_Font_Info(face, fontInfo) == 0 { // 获取PostScript字体信息
			family := C.GoString((*C.char)(fontInfo.family_name))
			fullname := C.GoString((*C.char)(fontInfo.full_name))
			families = append(families, family)
			fullnames = append(fullnames, fullname)
		}
		// 	fmt.Println("Error getting PS font info:", err)

	}

	// 	fmt.Println("No PS Glyph Names found in the font.")

	for i := range C.FT_Get_Sfnt_Name_Count(face) {
		if err := parseFontName(face, i, &families, &fullnames, &psnames); err != nil {
			if _, ok := err.(*UnsupportedIDError); ok {
				// fmt.Println(err)
				continue // 跳过不支持的名称ID
			} else if _, ok := err.(*UnsupportedPlatformError); ok {
				// fmt.Println(err)
				continue // 跳过不支持的平台ID
			}
			return nil, fmt.Errorf("error parsing font name at index %d: %w", idx, err)
		}
	}

	if len(families) == 0 && len(fullnames) == 0 && len(psnames) == 0 {
		return nil, ErrNoValidFontName
	}

	// 	fmt.Printf("Families: %v\n", families)
	// 	fmt.Printf("Fullnames: %v\n", fullnames)
	// 	fmt.Printf("PS Names: %v\n", psnames)

	fontInfo := FontInfo{
		Families:  families,
		Fullnames: fullnames,
		PSNames:   psnames,
		Weight:    getAssFaceWeight(face),                             // 字重
		Slant:     110 * int(face.style_flags&C.FT_STYLE_FLAG_ITALIC), // 0或110，斜体角度
		Index:     0,
	}
	if fontInfo.Slant < 0 || fontInfo.Slant > 110 {
		fontInfo.Slant = 0 // 如果斜体角度不在0-110范围内，设置为默认值0
	}
	if fontInfo.Weight < 100 || fontInfo.Weight > 900 {
		fontInfo.Weight = 400 // 如果字重不在100-900范围内，设置为默认值400
	}
	return &fontInfo, nil
}

// 设置 FT_Open_Args 结构体
func newOpenArgs(cFontPath *C.char) C.FT_Open_Args {
	var args C.FT_Open_Args

	args.flags = C.FT_OPEN_PATHNAME
	args.memory_base = nil
	args.memory_size = 0
	args.pathname = cFontPath
	args.stream = nil
	args.driver = nil
	args.num_params = 0
	args.params = nil
	return args
}

// 解析字体名称
func parseFontName(
	ftFace C.FT_Face,
	nameIdx C.uint,
	families, fullnames, psnames *[]string,
) error {
	var name C.FT_SfntName

	if C.FT_Get_Sfnt_Name(ftFace, nameIdx, &name) != 0 {
		return fmt.Errorf("error getting name for index %d", nameIdx)
	}
	// fmt.Println(name)

	// 检查名称ID和平台ID，只处理特定的名称ID
	// 0->TT_NAME_ID_COPYRIGHT版权
	// 1->TT_NAME_ID_FONT_FAMILY字体家族名称
	// 2->TT_NAME_ID_SUBFAMILY子家族名称
	// 3->TT_NAME_ID_UNIQUE_ID唯一ID
	// 4->TT_NAME_ID_FULL_NAME字体全名
	// 5->TT_NAME_ID_VERSION版本
	// 6->TT_NAME_ID_PS_NAME PostScript字体名称
	if name.name_id != TT_NAME_ID_FULL_NAME &&
		name.name_id != TT_NAME_ID_FONT_FAMILY &&
		name.name_id != TT_NAME_ID_PS_NAME {
		return &UnsupportedPlatformError{uint16(name.name_id)}
	}
	// fmt.Println("no skip name id:", name.name_id)

	// 检查平台ID，只处理微软平台
	// 0->TT_PLATFORM_UNICODE Unicode平台
	// 1->TT_PLATFORM_MAC Macintosh平台
	// 2->TT_PLATFORM_ISO ISO平台
	// 3->TT_PLATFORM_MICROSOFT Microsoft平台
	// 4->TT_PLATFORM_CUSTOM Custom平台
	// 5->TT_PLATFORM_ADOBE Adobe平台
	if name.platform_id != TT_PLATFORM_MICROSOFT {
		return &UnsupportedPlatformError{uint16(name.platform_id)}
	}
	// fmt.Println("no skip platform id:", name.platform_id)

	// 拷贝原始字节
	wbuf := C.GoBytes(unsafe.Pointer(name.string), C.int(name.string_len))
	buf := ""

	switch name.encoding_id {
	case TT_MS_ID_PRC: // 微软简体中文编码
		wbufn := bytes.ReplaceAll(wbuf, []byte{0}, nil)
		if !IconvConvert(wbufn, &buf, "GB2312", "UTF-8") {
			if !IconvConvert(wbuf, &buf, "UTF-16BE", "UTF-8") {
				return nil
			}
		}
	case TT_MS_ID_BIG_5: // 微软繁体中文编码
		wbufn := bytes.ReplaceAll(wbuf, []byte{0}, nil)
		if !IconvConvert(wbufn, &buf, "BIG-5", "UTF-8") {
			if !IconvConvert(wbuf, &buf, "UTF-16BE", "UTF-8") {
				return nil
			}
		}
	default:
		if !IconvConvert(wbuf, &buf, "UTF-16BE", "UTF-8") {
			return nil
		}
	}

	// 去除末尾的 '\0'
	buf = strings.TrimRight(buf, "\x00")
	if buf == "" {
		return nil
	}
	// fmt.Println("wbuf:", string(wbuf), "buf:", buf)

	switch name.name_id { // 根据名称ID处理不同的名称
	case TT_NAME_ID_FONT_FAMILY: // 字体家族名称
		if !contains(*families, buf) {
			family := strings.ToLower(buf)
			if family != "" && family != "undefined" {
				*families = append(*families, family)
			}
		}
	case TT_NAME_ID_FULL_NAME: // 字体全名
		if !contains(*fullnames, buf) {
			fullname := strings.ToLower(buf)
			if fullname != "" && fullname != "undefined" {
				*fullnames = append(*fullnames, fullname)
			}
		}
	case TT_NAME_ID_PS_NAME: // PostScript字体名称
		if !contains(*psnames, buf) {
			psname := strings.ToLower(buf)
			if psname != "" && psname != "undefined" {
				*psnames = append(*psnames, psname)
			}
		}
	}
	return nil
}

func getAssFaceWeight(face C.FT_Face) int {
	os2 := C.FT_Get_Sfnt_Table(face, C.FT_SFNT_OS2) // 获取OS/2表

	var os2Weight C.FT_UShort = 400 // 默认字重为400

	if os2 != nil {
		os2Weight = (*C.TT_OS2)(os2).usWeightClass // 需要将 os2 转换为 *C.TT_OS2 类型
	}

	switch os2Weight { // 根据OS/2表的字重值返回对应的字重
	case 0:
		bold := 0
		if (face.style_flags & C.FT_STYLE_FLAG_BOLD) != 0 {
			bold = 1
		}
		return 300*bold + 400
	case 1:
		return 100
	case 2:
		return 200
	case 3:
		return 300
	case 4:
		return 350
	case 5:
		return 400
	case 6:
		return 600
	case 7:
		return 700
	case 8:
		return 800
	case 9:
		return 900
	default:
		return int(os2Weight)
	}
}
