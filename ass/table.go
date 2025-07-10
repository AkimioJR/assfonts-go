package ass

import "strings"

// 样式表结构体
type StyleTable struct {
	Format            *FormatInfo         // 表头格式定义
	rows              []StyleInfo         // 数据行
	styleNameFontDesc map[string]FontDesc // 样式名->字体信息
}

func NewStyleTable(styleNameFontDesc map[string]FontDesc) *StyleTable {
	s := StyleTable{
		Format:            nil,
		rows:              make([]StyleInfo, 0),
		styleNameFontDesc: styleNameFontDesc,
	}
	return &s
}

// Append 添加样式到样式表
func (st *StyleTable) Append(style StyleInfo) {
	// Format: Name, Fontname, Fontsize, PrimaryColour, SecondaryColour, OutlineColour, BackColour, Bold, Italic, Underline, StrikeOut, ScaleX, ScaleY, Spacing, Angle, BorderStyle, Outline, Shadow, Alignment, MarginL, MarginR, MarginV, Encoding
	// Style: Default,方正准圆_GBK,48,&H00FFFFFF,&HF0000000,&H00665806,&H0058281B,0,0,0,0,100,100,1,0,1,2,0,2,30,30,10,1

	styleName, ok := style.Fields["Name"]
	if !ok || styleName == "" {
		styleName = defaultFontName
	}

	fontname, ok := style.Fields["Fontname"]
	if !ok {
		fontname = ""
	}
	fontname = strings.TrimPrefix(fontname, "@") // 去掉前缀 @（如果有的话）

	fd := FontDesc{
		FontName: fontname,
		Bold:     defaultFontSize, // 默认粗细大小
		Italic:   defaultItalic,   // 默认不斜体
	}

	if boldStr, ok := style.Fields["Bold"]; ok {
		if bold, err := calculateBold(boldStr); err == nil || err == ErrInvalidBoldValue {
			fd.Bold = bold // 计算粗体大小
		}
	}

	if italicStr, ok := style.Fields["Italic"]; ok {
		if italic, err := calculateItalic(italicStr); err == nil || err == ErrInvalidItalicValue {
			fd.Italic = italic // 是否启用斜体
		}
	}

	st.rows = append(st.rows, style)     // 添加样式到样式表
	st.styleNameFontDesc[styleName] = fd // 保存样式名称对应的字体描述
}

// 根据样式名称获取样式信息
func (st *StyleTable) GetFontDescByName(name string) *FontDesc {
	for styleName, fd := range st.styleNameFontDesc {
		if styleName == name {
			return &fd
		}
	}
	return nil
}

// 对话事件表结构体
type EventTable struct {
	Format *FormatInfo    // 表头格式定义
	Rows   []DialogueInfo // 数据行
}
