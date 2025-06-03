package ass

import "errors"

// RenameInfo 表示字体重命名信息
type RenameInfo struct {
	LineNum  uint64 // 样式行号
	Begin    uint   // 字体名称在原始样式行中的起始的偏移量（字节偏移量）
	End      uint   // 字体名称在原始样式行中结束位置的偏移量（字节偏移量）
	FontName string
	NewName  string
}

// TextInfo 表示文本行
type TextInfo struct {
	LineNum uint   // 行号
	Text    string // 文本内容
}

// StyleInfo 对应 ASS 样式信息
type StyleInfo struct {
	LineNum    uint64   // 行号
	RawContent string   // 原始内容
	Style      []string // 切分的字段
}

// DialogueInfo 对应 ASS 对话信息
type DialogueInfo struct {
	LineNum    uint     // 该 Dialogue 在原文件的行号
	RawContent string   // Dialogue 原始内容
	Dialogue   []string // Dialogue 切分后的字段
}

// FontDesc 表示字体描述
type FontDesc struct {
	FontName string // 字体名称
	Bold     uint   // 字粗
	Italic   bool   // 是否启用斜体
}

const (
	defaultFontName     = "Default" // 默认字体名称
	defaultFontSize     = 400       // 默认字体大小
	defaultBoldFontSize = 700       // 默认粗细大小
	defaultItalic       = false     // 默认不斜体
)

var (
	ErrStyleParseFailed   = errors.New("failed to parse style") // 未找到 [V4 Styles] 等模块
	ErrInvalidStyleFormat = errors.New("invalid style format")  // Styles 格式解析失败
	ErrEventParseFailed   = errors.New("failed to parse event") // 未找到 [Events] 等模块
	ErrInvalidEventFormat      = errors.New("invalid event format")  // Events 格式解析失败
	ErrInvalidBoldValue   = errors.New("invalid bold value")
)
