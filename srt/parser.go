package srt

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type ContentInfo struct {
	index uint64
	start string
	end   string
	text  string
}

type SRTParser struct {
	rawContent []string
	Contents   []ContentInfo
}

func NewSRTParser(reader io.Reader) (*SRTParser, error) {
	p := SRTParser{
		rawContent: make([]string, 0),
		Contents:   make([]ContentInfo, 0),
	}
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		p.rawContent = append(p.rawContent, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to new SRTParser: %w", err)
	}
	return &p, nil
}

func (p *SRTParser) Parse() error {
	for i := 0; i < len(p.rawContent); {
		line := strings.TrimSpace(p.rawContent[i])
		if line == "" {
			i++
			continue
		}

		// 尝试解析索引行
		idx, err := strconv.ParseUint(line, 10, 64)
		if err != nil {
			return fmt.Errorf("failed to parse SRT content at line %d: %w", i+1, err)
		}

		i++ // 移动到时间行
		if i >= len(p.rawContent) {
			return fmt.Errorf("unexpected end of file after index at line %d", i)
		}

		timeLine := strings.TrimSpace(p.rawContent[i])
		if !strings.Contains(timeLine, "-->") {
			return fmt.Errorf("expected time line at line %d", i+1)
		}

		// 解析时间行
		parts := strings.Split(timeLine, "-->")
		if len(parts) != 2 {
			return fmt.Errorf("invalid time line format at line %d", i+1)
		}
		start := strings.TrimSpace(parts[0])
		end := strings.TrimSpace(parts[1])

		i++ // 移动到文本行
		// 收集所有文本行直到空行或文件结束
		var textLines []string
		for i < len(p.rawContent) {
			line = strings.TrimSpace(p.rawContent[i])
			if line == "" {
				continue
			}

			// 检查是否意外遇到下一个索引行
			if _, err := strconv.ParseUint(line, 10, 64); err == nil {
				i--   // 如果是索引行，回退一个位置
				break // 结束当前字幕的文本收集
			}

			textLines = append(textLines, line)
			i++
		}

		// 合并多行文本
		text := strings.Join(textLines, "\n")

		p.Contents = append(p.Contents, ContentInfo{
			index: idx,
			start: start,
			end:   end,
			text:  text,
		})
	}
	return nil
}

func (p *SRTParser) ToASS(writer io.Writer, header string) error {
	if len(p.Contents) == 0 {
		return fmt.Errorf("no content to convert")
	}

	if _, err := writer.Write([]byte(header)); err != nil { // 写入ASS头部
		return err
	}

	// 写入每条字幕
	for _, content := range p.Contents {
		start := convertTime(content.start) // 转换时间格式 (HH:MM:SS.ff)
		end := convertTime(content.end)

		_, err := fmt.Fprintf(
			writer,
			"Dialogue: 0,%s,%s,Default,,0,0,0,,%s\n",
			start,
			end,
			strings.ReplaceAll(content.text, "\n", "\\N"),
		)
		if err != nil {
			return err
		}
	}

	return nil
}
