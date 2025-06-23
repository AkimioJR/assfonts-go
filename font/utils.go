package font

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

func isDir(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && fi.IsDir()
}

func iconvConvert(in []byte, fromCode, toCode string) (string, error) {
	var decoder *encoding.Decoder
	var encoder *encoding.Encoder

	// 选择解码器
	switch strings.ToUpper(fromCode) {
	case "UTF-16BE":
		decoder = unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM).NewDecoder()
	case "GB2312":
		decoder = simplifiedchinese.HZGB2312.NewDecoder()
	case "BIG-5", "BIG5":
		decoder = traditionalchinese.Big5.NewDecoder()
	case "ISO-8859-1":
		decoder = charmap.ISO8859_1.NewDecoder()
	case "UTF-8":
		decoder = encoding.Nop.NewDecoder()
	default:
		return "", NewErrUnSupportEncode(fromCode) // 不支持的编码
	}

	// 选择编码器
	switch strings.ToUpper(toCode) {
	case "UTF-8":
		encoder = encoding.Nop.NewEncoder()
	default:
		return "", ErrUnSupportEncode(toCode) // 这里只实现转 UTF-8
	}

	// 解码
	reader := transform.NewReader(bytes.NewReader(in), decoder)
	decoded, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("read encoder error: %w", err)
	}

	// 编码
	writer := &bytes.Buffer{}
	encoderWriter := transform.NewWriter(writer, encoder)
	defer encoderWriter.Close()
	_, err = encoderWriter.Write(decoded)
	if err != nil {
		return "", fmt.Errorf("write encoder error: %w", err)
	}

	return writer.String(), nil
}

func contains[T comparable](list []T, s T) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

var fontPattern = regexp.MustCompile(`(?i).+\.(ttf|otf|ttc|otc)$`)

// 查找字体文件
func findFontFiles(fontsDirs []string, withSystemFontPath bool) ([]string, error) {
	if withSystemFontPath {
		fontsDirs = append(fontsDirs, getDefaultFontPaths()...)
	}
	fontsPath := make([]string, 0, 10)
	for _, dir := range fontsDirs {
		filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil // 忽略错误
			}
			if !d.IsDir() && fontPattern.MatchString(d.Name()) {
				absPath, err := filepath.Abs(path)
				if err != nil {
					fontsPath = append(fontsPath, path)
				} else {
					fontsPath = append(fontsPath, absPath)
				}
			}
			return nil
		})
	}
	if len(fontsPath) == 0 {
		return nil, ErrNoFontFileFound
	}
	return fontsPath, nil
}

type Signed interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

func abs[T Signed](x T) T {
	if x < 0 {
		return -x
	}
	return x
}

// 辅助函数：格式化缺失码点
func formatCodepoints(codepoints []rune) string {
	var b strings.Builder
	for _, cp := range codepoints {
		// b.WriteString("0x" + fmt.Sprintf("%04X", cp) + "  ")
		b.WriteRune(cp)
	}
	return b.String()
}
