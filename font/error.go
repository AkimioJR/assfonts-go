package font

import (
	"errors"
	"fmt"
	"github/Akimio521/assfonts-go/ass"
)

var (
	ErrNoValidFontName = errors.New("no valid font names found")
	ErrNoValidFontFace = errors.New("no valid font face found")
	ErrNoContainFace   = errors.New("font contains no valid faces")
	ErrEmptySubsetData = errors.New("no subset font data collected")
	ErrNoFontFileFound = errors.New("no font files found")
)

type ErrUnsupportedID struct {
	platformID uint16
}

func (e *ErrUnsupportedID) Error() string {
	return fmt.Sprintf("skipping name with ID %d", e.platformID)
}

type ErrUnsupportedPlatform struct {
	platformID uint16
}

func (e *ErrUnsupportedPlatform) Error() string {
	return fmt.Sprintf("skipping name with platform ID %d", e.platformID)
}

type ErrOpenFontFace struct {
	path    string
	idx     uint
	errCode int
}

func NewErrOpenFontFace(p string, i uint, c int) *ErrOpenFontFace {
	return &ErrOpenFontFace{
		path:    p,
		idx:     i,
		errCode: c,
	}
}

func (e *ErrOpenFontFace) Error() string {
	return fmt.Sprintf("failed to create font face: \"%s\"[%d], error code: %d]", e.path, e.idx, e.errCode)
}

type ErrGetSFNTName struct {
	faceIdx uint
	nameIdx uint
	c       int
}

func NewErrGetSFNTName(faceIdx uint, nameIdx uint, c int) *ErrGetSFNTName {
	return &ErrGetSFNTName{
		faceIdx: faceIdx,
		nameIdx: nameIdx,
		c:       c,
	}
}

func (e *ErrGetSFNTName) Error() string {
	return fmt.Sprintf("failed to get SFNT name for face#%d name#%d: code %d", e.faceIdx, e.nameIdx, e.c)
}

type ErrMissCodepoints struct {
	fontDesc          *ass.FontDesc
	source            *FontFaceLocation
	missingCodepoints []rune
}

func (e *ErrMissCodepoints) Error() string {
	return fmt.Sprintf(`"%s"[%d] missing codepoints for face "%s" (%d,%d): %v`, e.source.Path, e.source.Index, e.fontDesc.FontName, e.fontDesc.Bold, e.fontDesc.Italic, formatCodepoints(e.missingCodepoints))
}

func NewErrMissCodepoints(fontDesc *ass.FontDesc, source *FontFaceLocation, missingCodepoints []rune) *ErrMissCodepoints {
	return &ErrMissCodepoints{
		fontDesc:          fontDesc,
		source:            source,
		missingCodepoints: missingCodepoints,
	}
}

type ErrUnSupportEncode string

func NewErrUnSupportEncode(s string) *ErrUnSupportEncode {
	var e = ErrUnSupportEncode(s)
	return &e
}

func (e ErrUnSupportEncode) Error() string {
	return fmt.Sprintf(`unsupport encoding type: "%s"`, string(e))
}

type WarningMsg string

func NewWarningMsg(format string, a ...any) *WarningMsg {
	w := WarningMsg(fmt.Sprintf(format, a...))
	return &w
}

func (w WarningMsg) Error() string {
	return string(w)
}

var _ error = (*ErrUnSupportEncode)(nil)
var _ error = (*WarningMsg)(nil)
