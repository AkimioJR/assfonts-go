package ass_test

import (
	"testing"

	"github.com/AkimioJR/assfonts-go/ass"
	"github.com/stretchr/testify/require"
)

func TestParseDataLineWithCommas(t *testing.T) {
	// 测试 Text 字段包含逗号的情况
	format := &ass.FormatInfo{
		Fields: []string{"Layer", "Start", "End", "Style", "Name", "MarginL", "MarginR", "MarginV", "Effect", "Text"},
	}

	testCases := []struct {
		name     string
		line     string
		expected map[string]string
	}{
		{
			name: "Text字段包含逗号",
			line: "Dialogue: 1,0:56:02.80,0:56:08.34,OP-JP,,0,0,10,,这是包含,逗号的文本内容",
			expected: map[string]string{
				"Layer":   "1",
				"Start":   "0:56:02.80",
				"End":     "0:56:08.34",
				"Style":   "OP-JP",
				"Name":    "",
				"MarginL": "0",
				"MarginR": "0",
				"MarginV": "10",
				"Effect":  "",
				"Text":    "这是包含,逗号的文本内容",
			},
		},
		{
			name: "Comment行包含多个逗号",
			line: "Comment: 0,0:00:03.00,0:00:06.00,Upper,,0,0,0,,{\\fad(300,300)}翻译：こばね sloka 杉树 SinceL  校对：错党  时轴：太刀 鮟鱇 SinceL  后期：太刀  繁化：错党",
			expected: map[string]string{
				"Layer":   "0",
				"Start":   "0:00:03.00",
				"End":     "0:00:06.00",
				"Style":   "Upper",
				"Name":    "",
				"MarginL": "0",
				"MarginR": "0",
				"MarginV": "0",
				"Effect":  "",
				"Text":    "{\\fad(300,300)}翻译：こばね sloka 杉树 SinceL  校对：错党  时轴：太刀 鮟鱇 SinceL  后期：太刀  繁化：错党",
			},
		},
		{
			name: "Text字段包含样式标签和逗号",
			line: "Dialogue: 1,0:56:02.80,0:56:08.34,OP-JP,,0,0,10,,{\\an2\\c&HFFFFFF&}翻译：abc, def, ghi",
			expected: map[string]string{
				"Layer":   "1",
				"Start":   "0:56:02.80",
				"End":     "0:56:08.34",
				"Style":   "OP-JP",
				"Name":    "",
				"MarginL": "0",
				"MarginR": "0",
				"MarginV": "10",
				"Effect":  "",
				"Text":    "{\\an2\\c&HFFFFFF&}翻译：abc, def, ghi",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ass.ParseDataLine(tc.line, format)
			require.NoErrorf(t, err, "parseDataLine 失败: %v", err)

			// 检查每个字段
			for field, expectedValue := range tc.expected {
				require.Contains(t, result, field, "字段 %s 不存在", field)
				require.Equal(t, expectedValue, result[field], "字段 %s 值不匹配: 期望 '%s', 实际 '%s'", field, expectedValue, result[field])
			}
			// 特别检查 Text 字段
			if text, exists := result["Text"]; exists {
				t.Logf("Text 字段解析结果: '%s'", text)
			}
		})
	}
}

func TestCleanEffects(t *testing.T) {

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "简单特效标记",
			input:    "{\\fade(500,500)}本字幕由动漫国字幕组制作",
			expected: "本字幕由动漫国字幕组制作",
		},
		{
			name:     "多重特效标记",
			input:    "{\\fad(500,0)\\fnB3CJROEU\\fs22\\frz19.65\\c&H6C6D6F&\\pos(468,349)}测试文本",
			expected: "测试文本",
		},
		{
			name:     "换行符处理",
			input:    "第一行\\N第二行\\n第三行",
			expected: "第一行\n第二行\n第三行",
		},
		{
			name:     "硬空格处理",
			input:    "文本\\h空格\\h处理",
			expected: "文本 空格 处理",
		},
		{
			name:     "转义花括号",
			input:    "普通文本\\{保留的花括号\\}",
			expected: "普通文本{保留的花括号}",
		},
		{
			name:     "混合内容",
			input:    "{\\fade(500,500)}本字幕由动漫国字幕组制作(dmguo.org)\\N仅供试看,请支持购买正版音像制品",
			expected: "本字幕由动漫国字幕组制作(dmguo.org)\n仅供试看,请支持购买正版音像制品",
		},
		{
			name:     "嵌套花括号",
			input:    "{\\fade(500,500){\\fn字体}}测试文本",
			expected: "测试文本",
		},
		{
			name:     "空文本",
			input:    "",
			expected: "",
		},
		{
			name:     "纯文本",
			input:    "没有特效的纯文本",
			expected: "没有特效的纯文本",
		},
		{
			name:     "不匹配的花括号",
			input:    "{\\fade(500,500没有结束的特效标记",
			expected: "没有结束的特效标记",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ass.CleanEffects(tt.input)
			require.Equal(t, tt.expected, result)
		})
	}
}
