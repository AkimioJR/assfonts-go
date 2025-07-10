package ass

import (
	"errors"
	"fmt"
)

type CodepointSet map[rune]struct{}

type ContentInfo struct {
	LineNum    uint   // 行号
	RawContent string // 文本内容
}

type FormatInfo struct {
	Fields []string // 字段名称列表
}

type StyleInfo struct {
	content    *ContentInfo      // 原始内容
	formatInfo *FormatInfo       // 格式定义
	Fields     map[string]string // 字段名->值的映射
}

type DialogueInfo struct {
	content    *ContentInfo      // 原始内容
	formatInfo *FormatInfo       // 格式定义
	Fields     map[string]string // 字段名->值的映射
}

type FontDesc struct {
	FontName string // 字体名称
	Bold     uint   // 字粗
	Italic   uint   // 是否启用斜体，0->不启用
}

// String 返回 FontDesc 的字符串表示，用于排序
func (fd *FontDesc) String() string {
	return fmt.Sprintf("%s_%d_%d", fd.FontName, fd.Bold, fd.Italic)
}

type parseState struct {
	inStyleSection bool // 是否在 [V4 Styles] 模块中
	inEventSection bool // 是否在 [Events] 模块中
	hasStyle       bool // 是否已找到 [V4 Styles] 模块
	hasEvent       bool // 是否已找到 [Events] 模块
}

// 样式表结构体
type StyleTable struct {
	Format *FormatInfo // 表头格式定义
	Rows   []StyleInfo // 数据行
}

// 根据样式名称获取样式信息
func (st *StyleTable) GetStyleByName(name string) (*StyleInfo, bool) {
	for i := range st.Rows {
		if styleName, ok := st.Rows[i].Fields["Name"]; ok && styleName == name {
			return &st.Rows[i], true
		}
	}
	return nil, false
}

// 对话事件表结构体
type EventTable struct {
	Format *FormatInfo    // 表头格式定义
	Rows   []DialogueInfo // 数据行
}

const (
	defaultFontName     = "Default" // 默认字体名称
	defaultFontSize     = 400       // 默认字体大小
	defaultBoldFontSize = 700       // 默认粗细大小
	defaultItalic       = 0         // 默认不斜体
	defaultItalicSlant  = 100       // 默认斜体倾斜度
)

var (
	ErrStyleParseFailed   = errors.New("failed to parse style") // 未找到 [V4 Styles] 等模块
	ErrInvalidStyleFormat = errors.New("invalid style format")  // Styles 格式解析失败
	ErrEventParseFailed   = errors.New("failed to parse event") // 未找到 [Events] 等模块
	ErrInvalidEventFormat = errors.New("invalid event format")  // Events 格式解析失败
	ErrInvalidBoldValue   = errors.New("invalid bold value")    // 不合法字重值
	ErrInvalidItalicValue = errors.New("invalid italic value")  // 不合法斜体值
	ErrMissingFormat      = errors.New("missing format line")   // 缺少格式定义行
)
