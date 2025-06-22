package font

import (
	"errors"
	"fmt"
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

type FontInfo struct {
	Families      []string  `json:"families"`        // 字体家族
	Fullnames     []string  `json:"fullnames"`       // 全名
	PSNames       []string  `json:"psnames"`         // PostScript 名称
	Weight        int       `json:"weight"`          // 字重，默认400
	Slant         int       `json:"slant"`           // 斜体角度，默认0
	Index         int64     `json:"index"`           // 字体索引，C++ long 对应 Go int64
	LastWriteTime time.Time `json:"last_write_time"` // 最后写入时间
	Path          string    `json:"path"`            // 字体文件路径
}
