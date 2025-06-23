package font

/*
#cgo CFLAGS: -I${SRCDIR}/clib/arm64-osx/include
#cgo LDFLAGS: -L${SRCDIR}/clib/arm64-osx/lib -lfreetype -lharfbuzz -lharfbuzz-subset -lbz2 -lbrotlidec -lbrotlicommon -lbrotlienc -lz -lpng16

// freetype
#include <ft2build.h>
#include FT_FREETYPE_H // <freetype/freetype.h>
#include FT_TYPE1_TABLES_H // <freetype/t1tables.h>
#include FT_TRUETYPE_IDS_H // <freetype/ttnameid.h>
#include FT_TRUETYPE_TABLES_H // <freetype/tttables.h>
#include FT_SFNT_NAMES_H // <freetype/ftsnames.h>

// harfbuzz
#define HB_EXPERIMENTAL_API
#include <harfbuzz/hb-subset.h>
*/
import "C"

import (
	"bytes"
	"errors"
	"fmt"
	"github/Akimio521/assfonts-go/ass"
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
func (lib *FreeTypeLibrary) ParseFont(fontPath string, ignoreError bool) ([]FontFaceInfo, error) {
	fileInfo, err := os.Stat(fontPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat font file %s: %w", fontPath, err)
	}

	fontData, err := os.ReadFile(fontPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read font file %s: %w", fontPath, err)
	}
	cFontData := C.CBytes(fontData) // 将Go字节切片转换为C字节数组
	defer C.free(cFontData)         // 确保在函数结束时释放C字节数组

	var metaFace C.FT_Face

	// 调用C函数解析字体
	errCode := C.FT_New_Memory_Face(lib.ptr, (*C.FT_Byte)(cFontData), C.FT_Long(len(fontData)), 0, &metaFace) // 0表示加载第一个字体
	if errCode != 0 {
		return nil, fmt.Errorf("parse font error, error code: %d", int(errCode))
	}
	defer C.FT_Done_Face(metaFace)

	facesNum := int64(metaFace.num_faces)              // 获取字体文件中的字体数量
	fontFaceInfos := make([]FontFaceInfo, 0, facesNum) // 初始化字体信息切片
	for idx := 0; idx < int(facesNum); idx++ {
		var face C.FT_Face = nil
		errCode = C.FT_New_Memory_Face(lib.ptr, (*C.FT_Byte)(cFontData), C.FT_Long(len(fontData)), C.FT_Long(idx), &face)
		if errCode != 0 {
			if ignoreError {
				continue // 如果忽略错误，跳过当前字体
			}
			return nil, fmt.Errorf("error creating face for font %s at index %d with error code %d", fontPath, idx, int(errCode))
		} else {
			defer C.FT_Done_Face(face)
		}

		fontFaceInfo, err := lib.parseFace(face)
		if err != nil {
			if ignoreError {
				continue // 如果忽略错误，跳过当前字体
			}
			return nil, fmt.Errorf("error parsing face at index %d: %w", idx, err)
		}
		if fontFaceInfo != nil {
			fontFaceInfo.Source = FontFaceLocation{ // 设置字体来源为文件
				Path:  fontPath,  // 设置字体文件路径
				Index: uint(idx), // 设置字体索引
			}
			fontFaceInfo.Modified = fileInfo.ModTime()           // 设置最后写入时间
			fontFaceInfos = append(fontFaceInfos, *fontFaceInfo) // 添加到字体信息切片
		}

	}

	if len(fontFaceInfos) == 0 {
		return nil, ErrNoValidFontFace
	}
	return fontFaceInfos, nil
}

func (lib *FreeTypeLibrary) parseFace(face C.FT_Face) (*FontFaceInfo, error) {
	var (
		families  []string = make([]string, 0) // 字体家族
		fullnames []string = make([]string, 0) // 字体全名
		psnames   []string = make([]string, 0) // PostScript字体名称
	)

	// if C.FT_Has_PS_Glyph_Names(face) != 0 { // 检查是否有PostScript字形名称
	// 	// var fontInfo C.PS_FontInfo = nil
	// 	fontInfo := (C.PS_FontInfo)(C.malloc(C.sizeof_PS_FontInfo))
	// 	defer C.free(unsafe.Pointer(fontInfo))

	// 	if C.FT_Get_PS_Font_Info(face, fontInfo) == 0 { // 获取PostScript字体信息
	// 		if fontInfo != nil {
	// 			fmt.Println("PostScript Font Info found")
	// 			fmt.Scanf("%s", &fontInfo) // 调试用，等待用户输入
	// 			family := C.GoString((*C.char)(fontInfo.family_name))
	// 			fullname := C.GoString((*C.char)(fontInfo.full_name))
	// 			families = append(families, family)
	// 			fullnames = append(fullnames, fullname)
	// 		}
	// 	}
	// 	// 	fmt.Println("Error getting PS font info:", err)

	// }

	// 	fmt.Println("No PS Glyph Names found in the font.")
	namesNum := int(C.FT_Get_Sfnt_Name_Count(face)) // 获取字体名称数量
	for i := 0; i < namesNum; i++ {
		if err := parseFontName(face, C.uint(i), &families, &fullnames, &psnames); err != nil {
			if _, ok := err.(*UnsupportedIDError); ok {
				// fmt.Println(err)
				continue // 跳过不支持的名称ID
			} else if _, ok := err.(*UnsupportedPlatformError); ok {
				// fmt.Println(err)
				continue // 跳过不支持的平台ID
			}
			return nil, fmt.Errorf("error parsing font name at index %d: %w", i, err)
		}
	}

	if len(families) == 0 && len(fullnames) == 0 && len(psnames) == 0 {
		return nil, ErrNoValidFontName
	}

	// 	fmt.Printf("Families: %v\n", families)
	// 	fmt.Printf("Fullnames: %v\n", fullnames)
	// 	fmt.Printf("PS Names: %v\n", psnames)

	fontFaceInfo := FontFaceInfo{
		Families:  families,
		FullNames: fullnames,
		PSNames:   psnames,
		Weight:    getAssFaceWeight(face), // 字重
		Slant:     getAssFaceSlant(face),  // 0或110，斜体角度
	}
	return &fontFaceInfo, nil
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
	if name.name_id != C.TT_NAME_ID_FULL_NAME &&
		name.name_id != C.TT_NAME_ID_FONT_FAMILY &&
		name.name_id != C.TT_NAME_ID_PS_NAME {
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
	if name.platform_id != C.TT_PLATFORM_MICROSOFT {
		return &UnsupportedPlatformError{uint16(name.platform_id)}
	}
	// fmt.Println("no skip platform id:", name.platform_id)

	// 拷贝原始字节
	wbuf := C.GoBytes(unsafe.Pointer(name.string), C.int(name.string_len))
	buf := ""

	switch name.encoding_id {
	case C.TT_MS_ID_PRC: // 微软简体中文编码
		wbufn := bytes.ReplaceAll(wbuf, []byte{0}, nil)
		if !IconvConvert(wbufn, &buf, "GB2312", "UTF-8") {
			if !IconvConvert(wbuf, &buf, "UTF-16BE", "UTF-8") {
				return nil
			}
		}
	case C.TT_MS_ID_BIG_5: // 微软繁体中文编码
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
	case C.TT_NAME_ID_FONT_FAMILY: // 字体家族名称
		if !contains(*families, buf) {
			family := strings.ToLower(buf)
			if family != "" && family != "undefined" {
				*families = append(*families, family)
			}
		}
	case C.TT_NAME_ID_FULL_NAME: // 字体全名
		if !contains(*fullnames, buf) {
			fullname := strings.ToLower(buf)
			if fullname != "" && fullname != "undefined" {
				*fullnames = append(*fullnames, fullname)
			}
		}
	case C.TT_NAME_ID_PS_NAME: // PostScript字体名称
		if !contains(*psnames, buf) {
			psname := strings.ToLower(buf)
			if psname != "" && psname != "undefined" {
				*psnames = append(*psnames, psname)
			}
		}
	}
	return nil
}

func getAssFaceWeight(face C.FT_Face) uint {
	os2 := C.FT_Get_Sfnt_Table(face, C.FT_SFNT_OS2) // 获取OS/2表

	var os2Weight C.FT_UShort = 400 // 默认字重为400

	if os2 != nil {
		os2Weight = (*C.TT_OS2)(os2).usWeightClass // 需要将 os2 转换为 *C.TT_OS2 类型
	} else {
		return 400 // 如果OS/2表不存在，返回默认字重400
	}

	switch os2Weight { // 根据OS/2表的字重值返回对应的字重
	case 0:
		var bold uint = 0
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
		if os2Weight < 100 || os2Weight > 900 {
			os2Weight = 400 // 如果字重不在100-900范围内，设置为默认值400
		}
		return uint(os2Weight)
	}
}

func getAssFaceSlant(face C.FT_Face) uint {
	slant := 110 * int(face.style_flags&C.FT_STYLE_FLAG_ITALIC)
	if slant < 0 || slant > 110 {
		slant = 0 // 如果斜体角度不在0-110范围内，设置为默认值0
	}
	return uint(slant)
}

func (db *FontDataBase) CheckGlyph(ffl *FontFaceLocation, fontSet ass.CodepointSet, fontDesc *ass.FontDesc) error {
	var missingCodepoints []rune
	var face C.FT_Face
	fontData, err := os.ReadFile(ffl.Path)
	if err != nil {
		return err
	}
	cFontData := C.CBytes(fontData)
	defer C.free(cFontData)

	errCode := C.FT_New_Memory_Face(db.lib.ptr, (*C.FT_Byte)(cFontData), C.FT_Long(len(fontData)), C.FT_Long(ffl.Index), &face)
	if errCode != 0 {
		return fmt.Errorf("parse font error, error code: %d", int(errCode))
	}
	defer C.FT_Done_Face(face) // 释放元字体对象

	for codepoint := range fontSet {
		if codepoint == 0 {
			continue
		}
		if C.FT_Get_Char_Index(face, C.FT_ULong(codepoint)) != 0 { // 未找到对应的字形
			missingCodepoints = append(missingCodepoints, codepoint)
		}
	}
	if len(missingCodepoints) > 0 {
		return NewErrMissCodepoints(fontDesc, missingCodepoints)
	}
	return nil
}

func (db *FontDataBase) CreatSubfont(subsetFontInfo *SubsetFontInfo) ([]byte, error) {
	if subsetFontInfo == nil {
		return nil, fmt.Errorf("subsetInfo is nil")
	}
	fontData, err := db.getFontData(subsetFontInfo.Source.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to read font file %s: %w", subsetFontInfo.Source.Path, err)
	}
	cFontData := C.CBytes(fontData)
	defer C.free(cFontData)

	// 创建 HarfBuzz blob
	blob := C.hb_blob_create((*C.char)(cFontData), C.uint(len(fontData)), C.HB_MEMORY_MODE_WRITABLE, nil, nil)
	defer C.hb_blob_destroy(blob)

	// 创建 HarfBuzz face
	face := C.hb_face_create(blob, 0)
	defer C.hb_face_destroy(face)

	// 创建 codepoint set
	cpSet := C.hb_set_create()
	defer C.hb_set_destroy(cpSet)
	for cp := range subsetFontInfo.Codepoints {
		C.hb_set_add(cpSet, C.uint(cp))
	}

	// 创建 subset input
	input := C.hb_subset_input_create_or_fail()
	if input == nil {
		return nil, errors.New("hb subset input create failed")
	}
	defer C.hb_subset_input_destroy(input)
	inputCodepoints := C.hb_subset_input_set(input, C.HB_SUBSET_SETS_UNICODE)
	C.hb_set_union(inputCodepoints, cpSet)

	// 子集化
	subsetFace := C.hb_subset_or_fail(face, input)
	if subsetFace == nil {
		return nil, errors.New("hb_subset_or_fail failed")
	}
	defer C.hb_face_destroy(subsetFace)

	// 获取子集数据
	subsetBlob := C.hb_face_reference_blob(subsetFace)
	defer C.hb_blob_destroy(subsetBlob)
	var length C.uint
	subsetData := C.hb_blob_get_data(subsetBlob, &length)
	if subsetData == nil || length == 0 {
		return nil, errors.New("hb_blob_get_data failed")
	}

	// 将C字节数组转换为Go字节切片
	goData := C.GoBytes(unsafe.Pointer(subsetData), C.int(length))
	return goData, nil
}
