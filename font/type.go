package font

import (
	"errors"
	"fmt"
	"github/Akimio521/assfonts-go/ass"
	"time"
)

const (
	TT_NAME_ID_FONT_FAMILY = 1  // 字体家族名称
	TT_NAME_ID_FULL_NAME   = 4  // 字体全名
	TT_NAME_ID_PS_NAME     = 6  // PostScript字体名称
	TT_PLATFORM_MICROSOFT  = 3  // 微软平台ID
	TT_MS_ID_PRC           = 21 // 微软简体中文编码ID
	TT_MS_ID_BIG_5         = 3  // 微软繁体中文编码ID
)

var (
	ErrNoValidFontName = errors.New("no valid font names found")
	ErrNoValidFontFace = errors.New("no valid font face found")
	ErrNoContainFace   = errors.New("font contains no valid faces")
)

// 这是标准 ASCII 可打印字符区间，包括空格（0x20）、数字、英文字母、常用标点符号等，一直到波浪号 ~（0x7e）
// 以及全角 ASCII 字符（0xff01 到 0xff5e），总共 95 个字符。
var additionalCodepoints []rune

func init() {
	additionalCodepoints = make([]rune, 0, 95)
	for ch := rune(0x0020); ch <= 0x007e; ch++ {
		additionalCodepoints = append(additionalCodepoints, ch)
	}
	for ch := rune(0xff01); ch <= 0xff5e; ch++ {
		additionalCodepoints = append(additionalCodepoints, ch)
	}
}

type UnsupportedIDError struct {
	platformID uint16
}

func (e *UnsupportedIDError) Error() string {
	return fmt.Sprintf("skipping name with ID %d", e.platformID)
}

type UnsupportedPlatformError struct {
	platformID uint16
}

func (e *UnsupportedPlatformError) Error() string {
	return fmt.Sprintf("skipping name with platform ID %d", e.platformID)
}

type ErrMissCodepoints struct {
	fontDesc          ass.FontDesc
	missingCodepoints []rune
}

func (e *ErrMissCodepoints) Error() string {
	return fmt.Sprintf(`missing codepoints for face "%s" (%d,%d) : %v`, e.fontDesc.FontName, e.fontDesc.Bold, e.fontDesc.Italic, formatCodepoints(e.missingCodepoints))
}

func NewErrMissCodepoints(fontDesc *ass.FontDesc, missingCodepoints []rune) *ErrMissCodepoints {
	return &ErrMissCodepoints{
		fontDesc:          *fontDesc,
		missingCodepoints: missingCodepoints,
	}
}

type FontFaceLocation struct {
	Path  string // 字体文件路径
	Index uint   // 字体集合中的索引位置
}

type FontFaceInfo struct {
	Source    FontFaceLocation `json:"source"`    // 字体来源信息
	Families  []string         `json:"families"`  // 字体家族名称列表
	FullNames []string         `json:"fullnames"` // 字体完整名称列表
	PSNames   []string         `json:"psnames"`   // PostScript 名称列表
	Weight    uint             `json:"weight"`    // 字重
	Slant     uint             `json:"slant"`     // 倾斜角度
	Modified  time.Time        `json:"modified"`  // 字体文件最后修改时间
}
type SubsetFontInfo struct {
	FontsDesc  ass.FontDesc     // 字体描述列表
	Codepoints ass.CodepointSet // 码点集合
	Source     FontFaceLocation // 字体路径及索引
}
