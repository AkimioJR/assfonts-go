package ass

import (
	"fmt"
	"io"
	"strings"
)

// 判断字符串是否有前缀（不区分大小写）
func startWith(raw string, prefix string) bool {
	return strings.HasPrefix(strings.ToLower(raw), strings.ToLower(prefix))
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
