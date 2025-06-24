package ass

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"
)

type ASSParser struct {
	Texts             []TextInfo                // 元素内容（不包含关于Font）
	Styles            []StyleInfo               // 包含哪些样式
	Dialogues         []DialogueInfo            // ASS 字幕 Dialogues 内容
	RenameInfos       []RenameInfo              // 记录字体调用位置
	FontSets          map[FontDesc]CodepointSet // 字体集
	HasFonts          bool                      // 是否包含字体样式
	HasDefaultStyle   bool                      // 是否有默认样式
	StyleNameFontDesc map[string]*FontDesc      // 样式描述
}

func NewASSParser(reader io.Reader) (*ASSParser, error) {
	ap := &ASSParser{
		Texts:             make([]TextInfo, 0, 200),
		RenameInfos:       make([]RenameInfo, 0, 10),
		FontSets:          make(map[FontDesc]CodepointSet),
		HasFonts:          false,
		HasDefaultStyle:   false,
		StyleNameFontDesc: make(map[string]*FontDesc),
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
			ap.HasFonts = true
			inFontsSection = true // 设置标志位
			continue              // 跳过 [Fonts] 行
		case "[events]", "[script info]", "[v4 styles]", "[v4+ styles]", "[graphics]":
			inFontsSection = false // 清除标志位
		}
		if !inFontsSection {
			ap.Texts = append(ap.Texts, TextInfo{LineNum: lineNum, Text: line})
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to new ASSParser: %w", err)
	}
	return ap, nil
}

func (ap *ASSParser) Parse() error {
	var (
		inStyleSection bool // 标记是否在样式区块中
		inEventSection bool // 标记是否在事件区块中
		hasStyle       bool // 标记是否解析到样式
		hasEvent       bool // 标记是否解析到事件
	)

	for _, ti := range ap.Texts {
		err := ap.parseTxet(&ti, &inStyleSection, &inEventSection, &hasStyle, &hasEvent)
		if err != nil {
			return fmt.Errorf("failed to parse text at line %d: %w", ti.LineNum, err)
		}
	}

	// 验证必要区块
	if !hasStyle {
		return ErrStyleParseFailed
	}
	if !hasEvent {
		return ErrEventParseFailed
	}
	ap.cleanFontSets()
	return nil
}

func (ap *ASSParser) parseTxet(text *TextInfo, inStyleSection *bool, inEventSection *bool, hasStyle *bool, hasEvent *bool) error {
	// 检查区块开始
	switch {
	case startWith(text.Text, "[V4+ Styles]"), startWith(text.Text, "[V4 Styles]"):
		*inStyleSection = true
		*inEventSection = false
		return nil

	case startWith(text.Text, "[Events]"):
		*inEventSection = true
		*inStyleSection = false
		return nil
	case startWith(text.Text, "["):
		*inStyleSection = false
		*inEventSection = false
	}

	// 根据当前状态处理行
	switch {
	case *inStyleSection && startWith(text.Text, "Style:"):
		err := ap.parseStyleLine(text)
		if err != nil {
			return ErrInvalidStyleFormat
		}
		*hasStyle = true

	case *inEventSection && startWith(text.Text, "Dialogue:"):
		err := ap.parseEventLine(text)
		if err != nil {
			return err
		}
		*hasEvent = true
	}
	return nil
}

// 解析单行样式
func (ap *ASSParser) parseStyleLine(text *TextInfo) error {
	fields := parseLine(text.Text, 10)
	if fields == nil {
		return ErrInvalidStyleFormat
	}

	si := StyleInfo{
		LineNum:    text.LineNum,
		RawContent: text.Text,
		Style:      fields,
	}
	ap.Styles = append(ap.Styles, si)
	ap.setStyleNameFontDesc(&si)
	return nil
}

// 解析单行事件
func (ap *ASSParser) parseEventLine(text *TextInfo) error {
	fields := parseLine(text.Text, 10)
	if fields == nil {
		return ErrInvalidEventFormat
	}
	di := DialogueInfo{
		LineNum:    text.LineNum,
		RawContent: text.Text,
		Dialogue:   fields,
	}
	ap.Dialogues = append(ap.Dialogues, di)
	return ap.parseDialogue(&di)
}

// 解析类似于 Style: 或 Dialogue: 这样的 ASS 行，将其各个属性分割出来
// 返回切片的长度不会小于 numField
func parseLine(line string, numField int) []string {
	// Style: Default,方正准圆_GBK,48,&H00FFFFFF,&HF0000000,&H00665806,&H0058281B,0,0,0,0,100,100,1,0,1,2,0,2,30,30,10,1

	// 先按冒号分割，再按逗号分割
	parts := strings.SplitN(line, ":", 2)
	if len(parts) < 2 {
		return nil
	}
	fields := []string{strings.TrimSpace(parts[0])}
	for _, f := range strings.SplitN(parts[1], ",", numField-1) {
		fields = append(fields, strings.TrimSpace(f))
	}
	if len(fields) < numField {
		return nil
	}
	return fields
}

// 根据传入的字符串判断并返回对应的“粗体”数值
// 转换失败时返回默认粗细大小 400
// "1"和"-1"被认为是启用粗体返回 700
// 否则返回其数值大小
func calculateBold(raw string) (uint, error) {
	value, err := strconv.Atoi(raw)
	if err != nil {
		return defaultFontSize, err
	}
	if value == 1 || value == -1 {
		return defaultBoldFontSize, nil
	} else if value <= 0 {
		return defaultFontSize, ErrInvalidBoldValue
	} else {
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

func (ap *ASSParser) setStyleNameFontDesc(style *StyleInfo) {
	// Format: Name, Fontname, Fontsize, PrimaryColour, SecondaryColour, OutlineColour, BackColour, Bold, Italic, Underline, StrikeOut, ScaleX, ScaleY, Spacing, Angle, BorderStyle, Outline, Shadow, Alignment, MarginL, MarginR, MarginV, Encoding
	// Style: Default,方正准圆_GBK,48,&H00FFFFFF,&HF0000000,&H00665806,&H0058281B,0,0,0,0,100,100,1,0,1,2,0,2,30,30,10,1

	if len(style.Style) > 1 && style.Style[1] == defaultFontName { // 检查是否为 Default 样式
		ap.HasDefaultStyle = true
	}

	styleName := style.Style[1]                         // 第二个字段是样式名称
	fontname := strings.TrimPrefix(style.Style[2], "@") // 第三个字段是字体名称，去掉前缀 @（如果有的话）
	fd := FontDesc{
		FontName: fontname,
		Bold:     400, // 默认粗细大小
		Italic:   0,   // 默认不斜体
	}
	if len(style.Style) > 8 {
		if bold, err := calculateBold(style.Style[8]); err == nil {
			fd.Bold = bold // 计算粗体大小
		} else {
			fd.Bold = defaultFontSize // 如果计算失败，使用默认值
		}
	}
	if len(style.Style) > 9 {
		if italic, err := calculateItalic(style.Style[9]); err == nil {
			fd.Italic = italic // 是否启用斜体
		} else {
			fd.Italic = defaultItalic // 如果计算失败，使用默认值
		}
	}
	ap.StyleNameFontDesc[styleName] = &fd // 保存样式名称对应的字体描述

	renameInfo := RenameInfo{
		FontName: fontname,
		LineNum:  style.LineNum,
		Begin:    uint(strings.Index(style.RawContent, fontname)),
		End:      uint(strings.Index(style.RawContent, fontname) + len(fontname)),
	}
	ap.RenameInfos = append(ap.RenameInfos, renameInfo)
}

// 统计每种字体样式实际用到的字符集合
func (ap *ASSParser) parseDialogue(dialogue *DialogueInfo) error {
	fd, err := ap.getFontDescStyle(dialogue)
	if err != nil {
		return fmt.Errorf("failed to get font description style for dialogue at line %d: %w", dialogue.LineNum, err)
	}
	localFD := *fd // 复制字体描述，避免修改原始数据

	// 初始化字体集合
	if _, ok := ap.FontSets[*fd]; !ok {
		ap.FontSets[*fd] = make(CodepointSet)
	}

	if len(dialogue.Dialogue) < 10 {
		// fmt.Println("Dialogue at line", dialogue.LineNum, "has no text content, skipping.", "\nrawContent:", dialogue.RawContent, "\nDialogue:", strings.Join(dialogue.Dialogue, "---"), "\nNum:", len(dialogue.Dialogue))
		return nil // 如果没有对话文本内容就跳过
	}

	runes := []rune(dialogue.Dialogue[9])
	idx := 0
	for idx < len(runes) {
		idx = ap.gatherCharacter(runes, idx, &localFD, dialogue.LineNum, &dialogue.RawContent)
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
func (p *ASSParser) getFontDescStyle(dialogue *DialogueInfo) (*FontDesc, error) {
	// Dialogue: 0,0:00:31.43,0:00:34.45,Default,NTP,0,0,0,,反复读了很多遍之后让我明白了不少事情

	var styleName string = defaultFontName // 默认样式
	if len(dialogue.Dialogue) > 4 {
		styleName = dialogue.Dialogue[4]
	}
	fd, ok := p.StyleNameFontDesc[styleName]
	if !ok {
		if p.HasDefaultStyle {
			fd = p.StyleNameFontDesc[defaultFontName]
		} else {
			return nil, fmt.Errorf("style '%s' not found", styleName)
		}
	}
	return fd, nil
}

// 处理对话文本中的每个字符，收集字体用到的字符
// 返回下一个未处理字符的索引
// fd 是当前对话使用的字体描述（不会进行修改）
// 注意：此方法会修改 idx 的值，指向下一个未处理的字符
func (ap *ASSParser) gatherCharacter(runes []rune, idx int, localFD *FontDesc, lineNum uint, rawContent *string) int {
	// 跳过 \h（空格） \n \N（换行）
	if idx < len(runes)-1 && runes[idx] == '\\' && (runes[idx+1] == 'h' || runes[idx+1] == 'n' || runes[idx+1] == 'N') {
		return idx + 2
	}
	// 样式覆盖段 {...}
	// Dialogue: 0,0:20:01.88,0:20:06.05,mianze,NTP,0,0,0,,{\fade(500,500)}本字幕由动漫国字幕组制作(dmguo.org)\N仅供试看,请支持购买正版音像制品
	if runes[idx] == '{' {
		endIdx := idx + 1
		for endIdx < len(runes) && runes[endIdx] != '}' {
			endIdx++
		}
		if endIdx >= len(runes) { // 没有找到 '}'，直接加入当前字符

			if localFD.FontName != "" {
				if _, ok := ap.FontSets[*localFD]; !ok {
					ap.FontSets[*localFD] = make(CodepointSet)
				}
				ap.FontSets[*localFD][runes[idx]] = struct{}{}
			}
			return idx + 1
		} else { // 处理样式覆盖
			override := string(runes[idx+1 : endIdx]) // \fad(500,0)\fnB3CJROEU\fs22\frz19.65\c&H6C6D6F&\pos(468,349)
			ap.styleOverride(override, localFD, lineNum, rawContent)
			return endIdx + 1
		}
	}
	// 普通字符
	if localFD.FontName != "" {
		if _, ok := ap.FontSets[*localFD]; !ok {
			ap.FontSets[*localFD] = make(CodepointSet)
		}
		ap.FontSets[*localFD][runes[idx]] = struct{}{}
	}
	return idx + 1
}

// 样式覆盖处理
func (ap *ASSParser) styleOverride(code string, localFD *FontDesc, lineNum uint, rawContent *string) {
	// code \fad(500,0)\fnB3CJROEU\fs22\frz19.65\c&H6C6D6F&\pos(468,349)

	fontPos := ap.changeFontname(code, localFD, lineNum, rawContent)
	boldPos := ap.changeBold(code, localFD)
	italicPos := ap.changeItalic(code, localFD)
	ap.changeStyle(code, localFD, lineNum, fontPos, boldPos, italicPos)
}

// 查找并处理 \fn 字体名覆盖
// 返回最后处理的位置
func (ap *ASSParser) changeFontname(code string, fd *FontDesc, lineNum uint, rawContent *string) int {
	pos := 0
	lastPos := 0

	for {
		lastPos = pos
		idx := strings.Index(code[pos:], `\fn`)
		if idx == -1 {
			break
		}
		pos += idx + 3 // 跳过 \fn
		// 提取字体名，直到下一个反斜杠或结尾
		rest := code[pos:]
		endIdx := strings.Index(rest, `\`)
		var fontView string
		if endIdx == -1 {
			fontView = rest
			pos = len(code)
		} else {
			fontView = rest[:endIdx]
			pos += endIdx
		}
		fontView = strings.TrimSpace(fontView)
		fontView = strings.TrimPrefix(fontView, "@") // 去掉前缀 @（如果有的话）
		if fontView == "" {
			continue // 如果没有指定字体名，使用原字体名
		}
		fd.FontName = fontView

		// 记录 RenameInfo
		beg := strings.Index(*rawContent, fontView)
		var end int
		if beg != -1 {
			end = beg + len(fontView)
		}
		renameInfo := RenameInfo{
			LineNum:  lineNum,
			Begin:    uint(beg),
			End:      uint(end),
			FontName: fontView,
		}
		ap.RenameInfos = append(ap.RenameInfos, renameInfo)
	}

	return lastPos
}

// 处理 \b 粗体覆盖
func (ap *ASSParser) changeBold(code string, fd *FontDesc) int {
	pos := 0
	lastPos := 0
	for {
		lastPos = pos
		idx := strings.Index(code[pos:], `\b`)
		if idx == -1 {
			break
		}
		pos += idx + 2 // 跳过 \b

		// 检查下一个字符是否为数字、- 或空格
		if pos >= len(code) || !(unicode.IsDigit(rune(code[pos])) || code[pos] == '-' || code[pos] == ' ') {
			continue
		}

		// 提取数字直到下一个反斜杠或结尾
		endIdx := pos
		for endIdx < len(code) && code[endIdx] != '\\' {
			endIdx++
		}
		boldStr := strings.TrimSpace(code[pos:endIdx])
		if boldStr != "" {
			if bold, err := calculateBold(boldStr); err == nil {
				fd.Bold = bold
			}
		}
		pos = endIdx
	}
	return lastPos
}

// 处理 \i 斜体覆盖
func (ap *ASSParser) changeItalic(code string, fd *FontDesc) int {
	pos := 0
	lastPos := 0
	for {
		lastPos = pos
		idx := strings.Index(code[pos:], `\i`)
		if idx == -1 {
			break
		}
		pos += idx + 2 // 跳过 \i

		// 检查下一个字符是否为数字、- 或空格
		if pos >= len(code) || !(unicode.IsDigit(rune(code[pos])) || code[pos] == '-' || code[pos] == ' ') {
			continue
		}

		// 提取数字直到下一个反斜杠或结尾
		endIdx := pos
		for endIdx < len(code) && code[endIdx] != '\\' {
			endIdx++
		}
		italicStr := strings.TrimSpace(code[pos:endIdx])
		if italicStr != "" {
			if italic, err := calculateItalic(italicStr); err == nil {
				fd.Italic = italic
			}
		}
		pos = endIdx
	}
	return lastPos
}

// 处理 \r 样式还原
func (ap *ASSParser) changeStyle(code string, localFD *FontDesc, lineNum uint, fontPos int, boldPos int, italicPos int) {
	pos := 0
	lastPos := 0
	updateFD := *localFD

	for {
		lastPos = pos
		idx := strings.Index(code[pos:], `\r`)
		if idx == -1 {
			break
		}
		pos += idx + 2 // 跳过 \r

		// 跳过 \rnd
		if pos+2 <= len(code) && code[pos:pos+2] == "nd" {
			continue
		}

		// 提取样式名直到下一个反斜杠或结尾
		endIdx := pos
		for endIdx < len(code) && code[endIdx] != '\\' {
			endIdx++
		}
		styleName := strings.TrimSpace(code[pos:endIdx])
		if styleName == "" {
			pos = endIdx
			continue
		}
		desc, ok := ap.StyleNameFontDesc[styleName]
		if !ok {
			fmt.Printf("Style \"%s\" not found. (Line %d)\n", styleName, lineNum)
		} else {
			updateFD = *desc // 复制字体描述，避免修改原始数据
		}
		pos = endIdx
	}

	// 根据 lastPos 和 fontPos、boldPos、italicPos 判断应该哪个字段被更新
	if lastPos > fontPos {
		localFD.FontName = updateFD.FontName
	}
	if lastPos > boldPos {
		localFD.Bold = updateFD.Bold
	}
	if lastPos > italicPos {
		localFD.Italic = updateFD.Italic
	}
}
func (ap *ASSParser) WriteWithEmbeddedFonts(fontDatas map[string][]byte, writer io.Writer) error {
	insertedFonts := false
	var err error

	for _, text := range ap.Texts {
		content := []byte(text.Text)
		trimmed := bytes.TrimSpace(bytes.ToLower(content))
		if !insertedFonts && bytes.Equal(trimmed, []byte("[events]")) {
			if _, err = writer.Write([]byte("[Fonts]")); err != nil {
				goto fail
			}
			for fontName, fontData := range fontDatas {
				if _, err = writer.Write([]byte("\nfontname:" + fontName + "\n")); err != nil {
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
		if _, err = writer.Write(content); err != nil {
			goto fail
		}
		if _, err = writer.Write([]byte("\n")); err != nil {
			goto fail
		}
	}
	return nil

fail:
	return fmt.Errorf("embed ass error when write to writer: %w", err)
}
