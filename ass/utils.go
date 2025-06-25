package ass

import (
	"fmt"
	"io"
	"strconv"
	"strings"
)

// 判断字符串是否有前缀（不区分大小写）
func startWith(raw string, prefix string) bool {
	return strings.HasPrefix(strings.ToLower(raw), strings.ToLower(prefix))
}

// 解析格式定义行（Format:）
func ParseFormat(line string) (*FormatInfo, error) {
	// Format: Name, Fontname, Fontsize, PrimaryColour, SecondaryColour, OutlineColour, BackColour, Bold, Italic, Underline, StrikeOut, ScaleX, ScaleY, Spacing, Angle, BorderStyle, Outline, Shadow, Alignment, MarginL, MarginR, MarginV, Encoding
	parts := strings.SplitN(line, ":", 2)
	if len(parts) != 2 {
		return nil, ErrInvalidStyleFormat
	}

	fieldsStr := strings.TrimSpace(parts[1])
	fieldNames := strings.Split(fieldsStr, ",")

	// 清理字段名称
	for i := range fieldNames {
		fieldNames[i] = strings.TrimSpace(fieldNames[i])
	}

	return &FormatInfo{Fields: fieldNames}, nil
}

// 解析数据行（Style: 或 Dialogue:）并返回字段映射
func ParseDataLine(line string, format *FormatInfo) (map[string]string, error) {
	// Style: Default,方正准圆_GBK,48,&H00FFFFFF,&HF0000000,&H00665806,&H0058281B,0,0,0,0,100,100,1,0,1,2,0,2,30,30,10,1
	// Dialogue: 1,0:56:02.80,0:56:08.34,OP-JP,,0,0,10,,{\an2\c&HFFFFFF&\bord4\blur3\fs50\fax-0.1\3c&HA0350D&}突然降る夕立　あぁ傘もないや嫌

	// 先按冒号分割
	parts := strings.SplitN(line, ":", 2)
	if len(parts) != 2 {
		return nil, ErrInvalidStyleFormat
	}

	fieldCount := len(format.Fields)
	values := strings.SplitN(strings.TrimSpace(parts[1]), ",", fieldCount)

	result := make(map[string]string)

	// 将分割的值与对应的字段名进行映射
	for i := 0; i < fieldCount && i < len(values); i++ {
		result[format.Fields[i]] = strings.TrimSpace(values[i])
	}

	return result, nil
}

func findTag(code []rune, pos int) (string, int) {
	start := pos
	for pos < len(code) {
		if code[pos] == '\\' {
			break
		}
		pos++
	}
	return string(code[start:pos]), pos
}

// 根据传入的字符串判断并返回对应的“粗体”数值
// 转换失败时返回默认粗细大小 400
// "1"和"-1"被认为是启用粗体返回 700
// "0" 返回默认粗细大小
// 否则返回其数值大小
func calculateBold(raw string) (uint, error) {
	value, err := strconv.Atoi(raw)
	if err != nil {
		return defaultFontSize, err
	}

	switch value {
	case 1, -1:
		return defaultBoldFontSize, nil
	case 0:
		return defaultFontSize, nil
	default:
		if value < 0 {
			return defaultFontSize, ErrInvalidBoldValue
		}
		return uint(value), nil
	}
}

// 仅"1"和"-1"被认为是启用斜体
func calculateItalic(raw string) (uint, error) {
	value, err := strconv.Atoi(raw)
	if err != nil {
		return defaultItalic, err
	}
	switch value {
	case 1, -1:
		return defaultItalicSlant, nil
	case 0:
		return defaultItalic, nil
	default:
		if value < 0 {
			return defaultItalic, ErrInvalidItalicValue
		}
		return uint(value), nil
	}
}

// 将二进制数据嵌入到文本文件中
// data：需要被嵌入的二进制内容
// writer: 需要写入的对象
// insertLinebreaks: 控制是否每 80 个字符插入换行
func UUEncode(data []byte, writer io.Writer, insertLinebreaks bool) error {
	var err error
	size := len(data)
	written := 0

	for pos := 0; pos < size; pos += 3 {
		src := [3]byte{0, 0, 0}
		n := copy(src[:], data[pos:min(pos+3, size)])

		dst := [4]byte{
			src[0] >> 2,
			((src[0]&0x3)<<4 | (src[1]&0xF0)>>4),
			((src[1]&0xF)<<2 | (src[2]&0xC0)>>6),
			src[2] & 0x3F,
		}

		for i := 0; i < min(n+1, 4); i++ {
			b := dst[i] + 33
			if _, err = writer.Write([]byte{b}); err != nil {
				goto fail
			}
			written++
			if insertLinebreaks && written == 80 && pos+3 < size {
				if _, err = writer.Write([]byte{'\n'}); err != nil {
					goto fail
				}
				written = 0
			}
		}
	}
	return nil
fail:
	return fmt.Errorf("write error when UUencoding: %w", err)
}
