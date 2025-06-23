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
