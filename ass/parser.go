package ass

import (
	"bufio"
	"fmt"
	"io"
	"sort"
	"strings"
	"unicode"
)

type ASSParser struct {
	Contents   []ContentInfo             // 元素内容
	StyleTable *StyleTable               // 样式表
	EventTable *EventTable               // 事件表
	FontSets   map[FontDesc]CodepointSet // 字体集
}

func NewASSParser(reader io.Reader) (*ASSParser, error) {
	ap := &ASSParser{
		Contents:   make([]ContentInfo, 0, 200),
		StyleTable: NewStyleTable(make(map[string]FontDesc)),
		EventTable: &EventTable{Rows: make([]DialogueInfo, 0)},
		FontSets:   make(map[FontDesc]CodepointSet),
	}

	var lineNum uint = 0
	var inFontsSection = false
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		lineNum++
		line := scanner.Text() // 读取一行
		temp := strings.TrimSpace(strings.ToLower(line))

		switch temp {
		case "[fonts]":
			inFontsSection = true // 设置标志位
			continue              // 跳过 [Fonts] 行
		case "[events]", "[script info]", "[v4 styles]", "[v4+ styles]", "[graphics]":
			inFontsSection = false // 清除标志位
		}
		if !inFontsSection {
			ap.Contents = append(ap.Contents, ContentInfo{LineNum: lineNum, RawContent: line})
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to new ASSParser: %w", err)
	}
	return ap, nil
}

func (ap *ASSParser) Parse() error {
	var s parseState
	var err error

	for i := range ap.Contents {
		s, err = ap.parseContent(i, s)
		if err != nil {
			return fmt.Errorf("failed to parse ass content at line %d: %w", ap.Contents[i].LineNum, err)
		}
	}

	// 验证必要区块
	if !s.hasStyle {
		return ErrStyleParseFailed
	}
	if !s.hasEvent {
		return ErrEventParseFailed
	}
	ap.cleanFontSets()
	return nil
}

func (ap *ASSParser) parseContent(i int, s parseState) (parseState, error) {
	ci := ap.Contents[i]
	// 检查区块开始
	switch {
	case startWith(ci.RawContent, "[V4+ Styles]"), startWith(ci.RawContent, "[V4 Styles]"):
		s.inStyleSection = true
		s.inEventSection = false
		ap.StyleTable.Format = nil // 重置格式定义
		return s, nil

	case startWith(ci.RawContent, "[Events]"):
		s.inEventSection = true
		s.inStyleSection = false
		ap.EventTable.Format = nil // 重置格式定义
		return s, nil
	case startWith(ci.RawContent, "["):
		s.inStyleSection = false
		s.inEventSection = false
	}

	// 根据当前状态处理行
	switch {
	case s.inStyleSection && startWith(ci.RawContent, "Format:"):
		// 解析样式格式定义
		format, err := ParseFormat(ci.RawContent)
		if err != nil {
			return s, err
		}
		ap.StyleTable.Format = format

	case s.inStyleSection && startWith(ci.RawContent, "Style:"):
		if ap.StyleTable.Format == nil {
			return s, ErrMissingFormat
		}
		si, err := parseStyleLine(&ap.Contents[i], ap.StyleTable.Format)
		if err != nil {
			return s, err
		}
		ap.StyleTable.Append(*si)
		s.hasStyle = true

	case s.inEventSection && startWith(ci.RawContent, "Format:"):
		// 解析事件格式定义
		format, err := ParseFormat(ci.RawContent)
		if err != nil {
			return s, err
		}
		ap.EventTable.Format = format

	case s.inEventSection && (startWith(ci.RawContent, "Dialogue:") || startWith(ci.RawContent, "Comment:")):
		if ap.EventTable.Format == nil {
			return s, ErrMissingFormat
		}
		di, err := parseEventLine(&ap.Contents[i], ap.EventTable.Format)
		if err != nil {
			return s, err
		}
		ap.EventTable.Rows = append(ap.EventTable.Rows, *di)
		s.hasEvent = true
	}
	return s, nil
}

// 统计每种字体样式实际用到的字符集合
func (ap *ASSParser) ParseDialogue(dialogue *DialogueInfo) error {
	initialFD, err := ap.getFontDescStyle(dialogue)
	if err != nil {
		return fmt.Errorf("failed to get font description style for dialogue at line %d: %w", dialogue.content.LineNum, err)
	}

	// 初始化字体集合
	if _, ok := ap.FontSets[initialFD]; !ok {
		ap.FontSets[initialFD] = make(CodepointSet)
	}

	// 获取对话文本内容
	text, ok := dialogue.Fields["Text"]
	if !ok || text == "" {
		return nil // 如果没有对话文本内容就跳过
	}

	runes := []rune(text)
	currentFD := initialFD // 当前对话使用的字体描述

	idx := 0
	for idx < len(runes) {
		idx = ap.gatherCharacter(runes, idx, &currentFD, &initialFD, dialogue.content)
	}
	return nil
}

// 删除空的字体集合
func (ap *ASSParser) cleanFontSets() {
	keysForDel := []FontDesc{}
	for fontDesc, set := range ap.FontSets {
		if len(set) == 0 {
			keysForDel = append(keysForDel, fontDesc)
		}
	}
	for _, key := range keysForDel {
		delete(ap.FontSets, key)
	}
}

// 获取对话对应的字体样式
// 不会对传入的 DialogueInfo 进行修改
func (ap *ASSParser) getFontDescStyle(dialogue *DialogueInfo) (FontDesc, error) {
	// Dialogue: 0,0:00:31.43,0:00:34.45,Default,NTP,0,0,0,,反复读了很多遍之后让我明白了不少事情

	var styleName string = defaultFontName // 默认样式
	if style, ok := dialogue.Fields["Style"]; ok && style != "" {
		styleName = style
	}

	fd := ap.StyleTable.GetFontDescByName(styleName)
	if fd == nil {
		return FontDesc{}, fmt.Errorf("style '%s' not found", styleName)
	}
	return *fd, nil
}

// 处理对话文本中的每个字符，收集字体用到的字符
// 返回下一个未处理字符的索引
// fd 是当前对话使用的字体描述（不会进行修改）
func (ap *ASSParser) gatherCharacter(runes []rune, idx int, currentFD *FontDesc, initialFD *FontDesc, ci *ContentInfo) int {
	if idx < len(runes)-1 && runes[idx] == '\\' {
		switch runes[idx+1] {
		case 'h', 'n', 'N': // 跳过 \h \n \N
			return idx + 2
		case '{', '}': // 转译 \{ \}
			if currentFD.FontName != "" {
				if _, ok := ap.FontSets[*currentFD]; !ok {
					ap.FontSets[*currentFD] = make(CodepointSet)
				}
				ap.FontSets[*currentFD][runes[idx+1]] = struct{}{}
			}
			return idx + 2 // 跳过 \{ \}
		}
	}

	// 样式覆盖段 {...}
	// Dialogue: 0,0:20:01.88,0:20:06.05,mianze,NTP,0,0,0,,{\fade(500,500)}本字幕由动漫国字幕组制作(dmguo.org)\N仅供试看,请支持购买正版音像制品
	if runes[idx] == '{' {
		endIdx := idx + 1
		for endIdx < len(runes) && runes[endIdx] != '}' {
			endIdx++
		}
		if endIdx >= len(runes) { // 没有找到 '}'，直接加入当前字符

			if currentFD.FontName != "" {
				if _, ok := ap.FontSets[*currentFD]; !ok {
					ap.FontSets[*currentFD] = make(CodepointSet)
				}
				ap.FontSets[*currentFD][runes[idx]] = struct{}{}
			}
			return idx + 1
		} else { // 处理样式覆盖
			// \fad(500,0)\fnB3CJROEU\fs22\frz19.65\c&H6C6D6F&\pos(468,349)
			ap.StyleOverride(runes[idx+1:endIdx], currentFD, initialFD, ci)
			return endIdx + 1
		}
	}
	// 普通字符
	if currentFD.FontName != "" {
		if _, ok := ap.FontSets[*currentFD]; !ok {
			ap.FontSets[*currentFD] = make(CodepointSet)
		}
		ap.FontSets[*currentFD][runes[idx]] = struct{}{}
	}
	return idx + 1
}

func (ap *ASSParser) StyleOverride(code []rune, currentFD *FontDesc, initialFD *FontDesc, ci *ContentInfo) {
	currentFDCopy := *currentFD // 创建当前字体描述的副本

	pos := 0
	for pos < len(code) {
		// 查找下一个标签开始位置
		if code[pos] != '\\' {
			pos++
			continue
		}

		pos++                 // 跳过 '\'
		if pos >= len(code) { // 如果已经到达字符串末尾，退出循环
			break
		}
		tagChar := code[pos] // 获取标签的第一个字符
		pos++                // 移动到标签内容开始位置

		switch tagChar {
		case 'f': // 处理字体相关标签 (\fn, \fr, 等)
			switch code[pos] {
			case 'n': // \fn
				pos++ // 跳过 'n'
				if pos >= len(code) {
					break // 如果已经到达字符串末尾，退出循环
				}

				var fontName string
				fontName, pos = findTag(code, pos)
				fontName = strings.TrimPrefix(strings.TrimSpace(fontName), "@")
				if fontName != "" {
					currentFDCopy.FontName = fontName
				}
			}
		case 'b': // 处理粗体标签 (\b)
			if pos < len(code) && (unicode.IsDigit(rune(code[pos])) || code[pos] == '-' || code[pos] == ' ') {
				var boldStr string
				boldStr, pos = findTag(code, pos)
				boldStr = strings.TrimSpace(boldStr)
				if bold, err := calculateBold(boldStr); err == nil || err == ErrInvalidBoldValue {
					currentFDCopy.Bold = bold
				}
			}
		case 'i': // 处理斜体标签 (\i)
			if pos < len(code) && (unicode.IsDigit(rune(code[pos])) || code[pos] == '-' || code[pos] == ' ') {
				var italicStr string
				italicStr, pos = findTag(code, pos)
				italicStr = strings.TrimSpace(italicStr)
				if italic, err := calculateItalic(italicStr); err == nil || err == ErrInvalidItalicValue {
					currentFDCopy.Italic = italic
				}
			}

		case 'r': // 处理样式重置标签 (\r)
			// 检查是否是\rnd标签
			if pos < len(code) && pos+1 < len(code) && code[pos] == 'n' && code[pos+1] == 'd' {
				pos += 2 // 跳过 "nd"
				continue
			}

			var styleName string
			if pos < len(code) {
				styleName, pos = findTag(code, pos)
			}
			styleName = strings.TrimSpace(styleName)

			if styleName == "" { // 无样式名时重置为初始样式
				currentFDCopy = *initialFD
			} else if desc := ap.StyleTable.GetFontDescByName(styleName); desc != nil { // 找到指定样式，更新当前字体描述
				currentFDCopy = *desc
			} else {
				fmt.Printf("Style \"%s\" not found. (Line %d)\n", styleName, ci.LineNum)
			}
		}
	}
	*currentFD = currentFDCopy // 更新最终的字体描述
}

func (ap *ASSParser) WriteWithEmbeddedFonts(fontDatas map[string][]byte, writer io.Writer) error {
	insertedFonts := false
	var err error

	for _, ci := range ap.Contents {
		if !insertedFonts && strings.ToLower(strings.TrimSpace(ci.RawContent)) == "[events]" {
			if _, err = writer.Write([]byte("[Fonts]")); err != nil {
				goto fail
			}

			// 对字体名称进行排序以确保输出的确定性
			var fontNames []string
			for fontName := range fontDatas {
				fontNames = append(fontNames, fontName)
			}
			sort.Strings(fontNames)

			// 按排序后的顺序写入字体
			for _, fontName := range fontNames {
				fontData := fontDatas[fontName]
				if _, err = writer.Write([]byte("\nfontname: " + fontName + "\n")); err != nil {
					goto fail
				}
				if err = UUEncode(fontData, writer, true); err != nil {
					goto fail
				}
			}
			if _, err = writer.Write([]byte("\n")); err != nil {
				goto fail
			}
			insertedFonts = true
		}
		if _, err = writer.Write([]byte(ci.RawContent + "\n")); err != nil {
			goto fail
		}
	}
	return nil

fail:
	return fmt.Errorf("embed ass error when write to writer: %w", err)
}

// 将 ASS 内容转换为 SRT 格式并写入指定的 Writer
func (ap *ASSParser) ToSRT(writer io.Writer) error {
	for i, di := range ap.EventTable.Rows {
		_, err := fmt.Fprintf(
			writer,
			"%d\n%s --> %s\n%s\n\n",
			i+1,
			di.Fields["Start"],
			di.Fields["End"],
			CleanEffects(di.Fields["Text"]),
		)
		if err != nil {
			return fmt.Errorf("failed to write SRT content at ASS line %d: %w", di.content.LineNum, err)
		}
	}
	return nil
}
