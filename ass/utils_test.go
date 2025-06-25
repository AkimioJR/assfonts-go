package ass_test

import (
	"github/Akimio521/assfonts-go/ass"
	"testing"
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
			if err != nil {
				t.Fatalf("parseDataLine 失败: %v", err)
			}

			// 检查字段数量
			if len(result) != len(tc.expected) {
				t.Errorf("字段数量不匹配: 期望 %d, 实际 %d", len(tc.expected), len(result))
			}

			// 检查每个字段
			for field, expectedValue := range tc.expected {
				if actualValue, exists := result[field]; !exists {
					t.Errorf("字段 %s 不存在", field)
				} else if actualValue != expectedValue {
					t.Errorf("字段 %s 值不匹配: 期望 '%s', 实际 '%s'", field, expectedValue, actualValue)
				}
			}

			// 特别检查 Text 字段
			if text, exists := result["Text"]; exists {
				t.Logf("Text 字段解析结果: '%s'", text)
			}
		})
	}
}
