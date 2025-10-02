package font

import (
	"errors"
	"fmt"

	"github.com/AkimioJR/assfonts-go/ass"
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

type ErrMissingFontFaceFound ass.FontDesc

func NewErrMissingFontFaceFound(f ass.FontDesc) *ErrMissingFontFaceFound {
	e := ErrMissingFontFaceFound(f)
	return &e
}

func (e *ErrMissingFontFaceFound) Error() string {
	return fmt.Sprintf(`missing font face found for "%s" (%d,%d)`, e.FontName, e.Bold, e.Italic)
}

type ErrSubsetInputCreate struct {
	s FontFaceLocation
}

func NewErrSubsetInputCreate(s *FontFaceLocation) *ErrSubsetInputCreate {
	return &ErrSubsetInputCreate{s: *s}
}

func (e *ErrSubsetInputCreate) Error() string {
	return fmt.Sprintf(`failed to create HarfBuzz subset input for font: "%s"[%d]`, e.s.Path, e.s.Index)
}

type ErrSubsetFail struct {
	s              FontFaceLocation
	codepointCount int
}

func NewErrSubsetFail(s *FontFaceLocation, codepointCount int) *ErrSubsetFail {
	return &ErrSubsetFail{s: *s, codepointCount: codepointCount}
}

func (e *ErrSubsetFail) Error() string {
	return fmt.Sprintf(`failed to create font subset for font "%s"[%d] with %d codepoints`, e.s.Path, e.s.Index, e.codepointCount)
}

type ErrSubsetDataGet struct {
	s          FontFaceLocation
	dataLength uint
}

func NewErrSubsetDataGet(s *FontFaceLocation, dataLength uint) *ErrSubsetDataGet {
	return &ErrSubsetDataGet{s: *s, dataLength: dataLength}
}

func (e *ErrSubsetDataGet) Error() string {
	return fmt.Sprintf(`failed to get subset font data for font  "%s"[%d], data length: %d`, e.s.Path, e.s.Index, e.dataLength)
}

type WarningMsg string

func NewWarningMsg(format string, a ...any) *WarningMsg {
	w := WarningMsg(fmt.Sprintf(format, a...))
	return &w
}

func (w WarningMsg) Error() string {
	return string(w)
}

type InfoMsg string

func NewInfoMsg(format string, a ...any) *InfoMsg {
	w := InfoMsg(fmt.Sprintf(format, a...))
	return &w
}

func (i InfoMsg) Error() string {
	return string(i)
}

var _ error = (*ErrUnSupportEncode)(nil)
var _ error = (*ErrSubsetInputCreate)(nil)
var _ error = (*ErrSubsetFail)(nil)
var _ error = (*ErrSubsetDataGet)(nil)
var _ error = (*WarningMsg)(nil)
var _ error = (*InfoMsg)(nil)
