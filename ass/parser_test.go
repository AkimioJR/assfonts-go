package ass_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/Akimio521/assfonts-go/ass"

	"github.com/stretchr/testify/require"
)

type StyleOverrideTestCase struct {
	name   string
	code   string
	origin ass.FontDesc
	expect ass.FontDesc
}

var styleOverrideTestCases = []StyleOverrideTestCase{
	{
		name: `修改字体+\b1修改字重`,
		code: `\fn华康手札体W7-A\bord0.5\3c&H40ECED&\1c&H40ECED&\b1\fscx80\fs78\fsp-2\pos(1024,915.17)\frz0.6129`,
		origin: ass.FontDesc{
			FontName: "汉仪正圆-65S",
			Bold:     400,
			Italic:   0,
		},
		expect: ass.FontDesc{
			FontName: "华康手札体W7-A",
			Bold:     700,
			Italic:   0,
		},
	},
	{
		name: `修改字体+\b0修改字重`,
		code: `\bord0\fn思源黑体 CN\b0`,
		origin: ass.FontDesc{
			FontName: "汉仪正圆-65S",
			Bold:     600,
			Italic:   0,
		},
		expect: ass.FontDesc{
			FontName: "思源黑体 CN",
			Bold:     400,
			Italic:   0,
		},
	},
	{
		name: `修改字体+\b1修改字重`,
		code: `\fn华康手札体W7-A\b1\fax-0.1\bord3\1c&H9EE014&\3c&H9EE014&\blur8\frz335.9\pos(1632.538,805.205)\frx10\fry2\fad(0,0)`,
		origin: ass.FontDesc{
			FontName: "汉仪正圆-65S",
			Bold:     400,
			Italic:   0,
		},
		expect: ass.FontDesc{
			FontName: "华康手札体W7-A",
			Bold:     700,
			Italic:   0,
		},
	},
	{
		name: `修改字体+\b500修改字重+\i0取消斜体`,
		code: `\fn方正粗雅宋_GBK\fs180\1c&H000000&\b500\fsp8\an8\pos(970,140)\i0`,
		origin: ass.FontDesc{
			FontName: "汉仪正圆-65S",
			Bold:     400,
			Italic:   110,
		},
		expect: ass.FontDesc{
			FontName: "方正粗雅宋_GBK",
			Bold:     500,
			Italic:   0,
		},
	},
	{
		name: `修改字体+\b1+\i1`,
		code: `\fn宋体\b1\i1`,
		origin: ass.FontDesc{
			FontName: "汉仪正圆-65S",
			Bold:     400,
			Italic:   0,
		},
		expect: ass.FontDesc{
			FontName: "宋体",
			Bold:     700,
			Italic:   100,
		},
	},
	{
		name: `修改样式+\b500`,
		code: `\rSongTi\fs180\1c&H000000&\b500\fsp8\an8\pos(970,140)`,
		origin: ass.FontDesc{
			FontName: "汉仪正圆-65S",
			Bold:     400,
			Italic:   0,
		},
		expect: ass.FontDesc{
			FontName: "宋体",
			Bold:     500,
			Italic:   50,
		},
	},
	{
		name: `b500+修改样式`,
		code: `\fs180\1c&H000000&\b500\fsp8\rSongTi\an8\r\i70\pos(970,140)`,
		origin: ass.FontDesc{
			FontName: "汉仪正圆-65S",
			Bold:     400,
			Italic:   0,
		},
		expect: ass.FontDesc{
			FontName: "汉仪正圆-65S",
			Bold:     400,
			Italic:   70,
		},
	},
}

type ParseDialogueTestCase struct {
	name   string
	d      ass.DialogueInfo
	fd     map[string]ass.FontDesc
	expect map[ass.FontDesc]ass.CodepointSet
}

var parseDialogueTestCases = []ParseDialogueTestCase{
	{
		name: "基础文本",
		d: ass.DialogueInfo{
			Fields: map[string]string{
				"Layer":   "0",
				"Start":   "0:00:00.00",
				"End":     "0:00:05.00",
				"Style":   "Default",
				"Name":    "",
				"MarginL": "0",
				"MarginR": "0",
				"MarginV": "0",
				"Effect":  "",
				"Text":    "简单文本",
			},
		},
		fd: map[string]ass.FontDesc{
			"Default": {FontName: "楷体", Bold: 400, Italic: 0},
		},
		expect: map[ass.FontDesc]ass.CodepointSet{
			{FontName: "楷体", Bold: 400, Italic: 0}: {
				'简': {},
				'单': {},
				'文': {},
				'本': {},
			},
		},
	},
	{
		name: "样式重置到初始",
		d: ass.DialogueInfo{
			Fields: map[string]string{
				"Layer":   "0",
				"Start":   "0:00:00.00",
				"End":     "0:00:05.00",
				"Style":   "Default",
				"Name":    "",
				"MarginL": "0",
				"MarginR": "0",
				"MarginV": "0",
				"Effect":  "",
				"Text":    `{\fn宋体\b1}重置前{\r}重置后`,
			},
		},
		fd: map[string]ass.FontDesc{
			"Default": {FontName: "楷体", Bold: 400, Italic: 0},
		},
		expect: map[ass.FontDesc]ass.CodepointSet{
			{FontName: "宋体", Bold: 700, Italic: 0}: {
				'重': {},
				'置': {},
				'前': {},
			},
			{FontName: "楷体", Bold: 400, Italic: 0}: {
				'重': {},
				'置': {},
				'后': {},
			},
		},
	},
	{
		name: "样式重置到指定",
		d: ass.DialogueInfo{
			Fields: map[string]string{
				"Layer":   "0",
				"Start":   "0:00:00.00",
				"End":     "0:00:05.00",
				"Style":   "style1",
				"Name":    "",
				"MarginL": "0",
				"MarginR": "0",
				"MarginV": "0",
				"Effect":  "",
				"Text":    `{\rstyle2}指定样式`,
			},
		},
		fd: map[string]ass.FontDesc{
			"style1": {FontName: "楷体", Bold: 400, Italic: 0},
			"style2": {FontName: "宋体", Bold: 400, Italic: 0},
		},
		expect: map[ass.FontDesc]ass.CodepointSet{
			{FontName: "楷体", Bold: 400, Italic: 0}: {},
			{FontName: "宋体", Bold: 400, Italic: 0}: {
				'指': {},
				'定': {},
				'样': {},
				'式': {},
			},
		},
	},
	{
		name: "转义字符",
		d: ass.DialogueInfo{
			Fields: map[string]string{
				"Layer":   "0",
				"Start":   "0:00:00.00",
				"End":     "0:00:05.00",
				"Style":   "Default",
				"Name":    "",
				"MarginL": "0",
				"MarginR": "0",
				"MarginV": "0",
				"Effect":  "",
				"Text":    `转义\{字符\}测试`,
			},
		},
		fd: map[string]ass.FontDesc{
			"Default": {FontName: "楷体", Bold: 400, Italic: 0},
		},
		expect: map[ass.FontDesc]ass.CodepointSet{
			{FontName: "楷体", Bold: 400, Italic: 0}: {
				'转': {},
				'义': {},
				'{': {},
				'字': {},
				'符': {},
				'}': {},
				'测': {},
				'试': {},
			},
		},
	},
	{
		name: "重置后转义",
		d: ass.DialogueInfo{
			Fields: map[string]string{
				"Layer":   "0",
				"Start":   "0:00:00.00",
				"End":     "0:00:05.00",
				"Style":   "Default",
				"Name":    "",
				"MarginL": "0",
				"MarginR": "0",
				"MarginV": "0",
				"Effect":  "",
				"Text":    `{\fn宋体}重置前{\r}\{重置后\}`,
			},
		},
		fd: map[string]ass.FontDesc{
			"Default": {FontName: "楷体", Bold: 400, Italic: 0},
		},
		expect: map[ass.FontDesc]ass.CodepointSet{
			{FontName: "宋体", Bold: 400, Italic: 0}: {
				'重': {},
				'置': {},
				'前': {},
			},
			{FontName: "楷体", Bold: 400, Italic: 0}: {
				'{': {},
				'重': {},
				'置': {},
				'后': {},
				'}': {},
			},
		},
	},
	{
		name: "混合样式标签",
		d: ass.DialogueInfo{
			Fields: map[string]string{
				"Layer":   "0",
				"Start":   "0:00:00.00",
				"End":     "0:00:05.00",
				"Style":   "style1",
				"Name":    "",
				"MarginL": "0",
				"MarginR": "0",
				"MarginV": "0",
				"Effect":  "",
				"Text":    `{\fnA\b1}粗体A{\rstyle2\i1}斜体B{\r}普通`,
			},
		},
		fd: map[string]ass.FontDesc{
			"style1": {FontName: "楷体", Bold: 400, Italic: 0},
			"style2": {FontName: "宋体", Bold: 400, Italic: 0},
		},
		expect: map[ass.FontDesc]ass.CodepointSet{
			{FontName: "A", Bold: 700, Italic: 0}: {
				'粗': {},
				'体': {},
				'A': {},
			},
			{FontName: "宋体", Bold: 400, Italic: 100}: {
				'斜': {},
				'体': {},
				'B': {},
			},
			{FontName: "楷体", Bold: 400, Italic: 0}: {
				'普': {},
				'通': {},
			},
		},
	},
	{
		name: "特殊字符",
		d: ass.DialogueInfo{
			Fields: map[string]string{
				"Layer":   "0",
				"Start":   "0:00:00.00",
				"End":     "0:00:05.00",
				"Style":   "style1",
				"Name":    "",
				"MarginL": "0",
				"MarginR": "0",
				"MarginV": "0",
				"Effect":  "",
				"Text":    `特殊字符: \n \h \{ \} \\`,
			},
		},
		fd: map[string]ass.FontDesc{
			"style1": {FontName: "楷体", Bold: 400, Italic: 0},
		},
		expect: map[ass.FontDesc]ass.CodepointSet{
			{FontName: "楷体", Bold: 400, Italic: 0}: {
				'特':  {},
				'殊':  {},
				'字':  {},
				'符':  {},
				':':  {},
				' ':  {},
				'{':  {},
				'}':  {},
				'\\': {},
			},
		},
	},
	{
		name: "复杂嵌套样式",
		d: ass.DialogueInfo{
			Fields: map[string]string{
				"Layer":   "0",
				"Start":   "0:00:00.00",
				"End":     "0:00:10.00",
				"Style":   "style1",
				"Name":    "",
				"MarginL": "0",
				"MarginR": "0",
				"MarginV": "0",
				"Effect":  "",
				"Text":    `开始{\fnA}样式A{\fnB\b1}样式B{\r}重置{\rstyle2}样式2{\r}结束`,
			},
		},
		fd: map[string]ass.FontDesc{
			"style1": {FontName: "楷体", Bold: 400, Italic: 0},
			"style2": {FontName: "宋体", Bold: 400, Italic: 0},
		},
		expect: map[ass.FontDesc]ass.CodepointSet{
			{FontName: "楷体", Bold: 400, Italic: 0}: {
				'开': {},
				'始': {},
				'重': {},
				'置': {},
				'结': {},
				'束': {},
			},
			{FontName: "A", Bold: 400, Italic: 0}: {
				'样': {},
				'式': {},
				'A': {},
			},
			{FontName: "B", Bold: 700, Italic: 0}: {
				'样': {},
				'式': {},
				'B': {},
			},
			{FontName: "宋体", Bold: 400, Italic: 0}: {
				'样': {},
				'式': {},
				'2': {},
			},
		},
	},
	{
		name: "空样式块",
		d: ass.DialogueInfo{
			Fields: map[string]string{
				"Layer":   "0",
				"Start":   "0:00:00.00",
				"End":     "0:00:05.00",
				"Style":   "style1",
				"Name":    "",
				"MarginL": "0",
				"MarginR": "0",
				"MarginV": "0",
				"Effect":  "",
				"Text":    `开始{}结束`,
			},
		},
		fd: map[string]ass.FontDesc{
			"style1": {FontName: "楷体", Bold: 400, Italic: 0},
		},
		expect: map[ass.FontDesc]ass.CodepointSet{
			{FontName: "楷体", Bold: 400, Italic: 0}: {
				'开': {},
				'始': {},
				'结': {},
				'束': {},
			},
		},
	},
	{
		name: "多语言字符",
		d: ass.DialogueInfo{
			Fields: map[string]string{
				"Layer":   "0",
				"Start":   "0:00:00.00",
				"End":     "0:00:05.00",
				"Style":   "style1",
				"Name":    "",
				"MarginL": "0",
				"MarginR": "0",
				"MarginV": "0",
				"Effect":  "",
				"Text":    `中文Chinese 日本語한국어`,
			},
		},
		fd: map[string]ass.FontDesc{
			"style1": {FontName: "楷体", Bold: 400, Italic: 0},
		},
		expect: map[ass.FontDesc]ass.CodepointSet{
			{FontName: "楷体", Bold: 400, Italic: 0}: {
				'中': {},
				'文': {},
				'C': {},
				'h': {},
				'i': {},
				'n': {},
				'e': {},
				's': {},
				' ': {},
				'日': {},
				'本': {},
				'語': {},
				'한': {},
				'국': {},
				'어': {},
			},
		},
	},
	{
		name: "混合测试",
		d: ass.DialogueInfo{
			Fields: map[string]string{
				"Layer":   "0",
				"Start":   "0:00:00.88",
				"End":     "0:59:59.82",
				"Style":   "style1",
				"Name":    "",
				"MarginL": "0",
				"MarginR": "0",
				"MarginV": "0",
				"Effect":  "",
				"Text":    `我{你甚至可以在这里写注释\CODE_HERE\前面可以是一段代码，但无需关注}能{这里是\rndx10}吞下{\fn宋体\b1\i1}玻璃而{\pos(400,400)}不{\r}伤身{\rstyle2}体\{这是转义的\n括号\}`,
			},
		},
		fd: map[string]ass.FontDesc{
			"Default": {
				FontName: "黑体",
				Bold:     400,
				Italic:   0,
			},
			"style1": {
				FontName: "楷体",
				Bold:     400,
				Italic:   0,
			},
			"style2": {
				FontName: "宋体",
				Bold:     400,
				Italic:   0,
			},
		},
		expect: map[ass.FontDesc]ass.CodepointSet{
			{ // style1
				FontName: "楷体",
				Bold:     400,
				Italic:   0,
			}: {
				'我': {}, //25105
				'能': {}, //33021
				'吞': {}, // 21534
				'下': {}, //19979
				'伤': {}, // 20260
				'身': {}, //36523
			},
			{ // 宋体粗斜体
				FontName: "宋体",
				Bold:     700,
				Italic:   100,
			}: {
				'玻': {}, //29627
				'璃': {}, //29827
				'而': {}, //32780
				'不': {}, //19981

			},
			{ // 宋体常规 (style2)
				FontName: "宋体",
				Bold:     400,
				Italic:   0,
			}: {
				'体': {}, //20307
				'{': {}, //123
				'这': {}, //36825
				'是': {}, //26159
				'转': {}, //36716
				'义': {}, //20041
				'的': {}, //30340
				'括': {}, //25324
				'号': {}, //21495
				'}': {}, //125
			},
		},
	},
}

func TestStyleOverride(t *testing.T) {
	for _, c := range styleOverrideTestCases {
		t.Run(
			c.name,
			func(t *testing.T) {
				ap := ass.ASSParser{
					StyleTable: ass.NewStyleTable(map[string]ass.FontDesc{
						"SongTi": {
							FontName: "宋体",
							Bold:     700,
							Italic:   50,
						},
					}),
				}
				currentFD := c.origin
				ap.StyleOverride([]rune(c.code), &currentFD, &c.origin, nil)
				require.Equal(t, c.expect, currentFD)
			},
		)

	}

}
func TestParseDialogue(t *testing.T) {
	for _, c := range parseDialogueTestCases {
		t.Run(
			c.name,
			func(t *testing.T) {
				ap := ass.ASSParser{
					FontSets:   make(map[ass.FontDesc]ass.CodepointSet),
					StyleTable: ass.NewStyleTable(c.fd),
				}
				err := ap.ParseDialogue(&c.d)
				require.NoError(t, err)
				require.Equal(t, c.expect, ap.FontSets)
			},
		)

	}

}

func generateASSContent(dialogueCount int) string {
	chineseWords := []string{
		"世界", "生活", "时间", "朋友", "工作", "学习", "梦想", "希望", "快乐", "美好",
		"阳光", "雨水", "花朵", "音乐", "故事", "旅行", "家庭", "爱情", "友谊", "勇气",
		"智慧", "创造", "发现", "探索", "成长", "改变", "坚持", "努力", "成功", "失败",
		"经验", "回忆", "未来", "现在", "过去", "机会", "挑战", "选择", "决定", "行动",
		"思考", "感受", "体验", "享受", "分享", "帮助", "支持", "理解", "包容", "宽恕",
		"春天", "夏日", "秋风", "冬雪", "山川", "河流", "大海", "星空", "月亮", "太阳",
		"森林", "草原", "城市", "乡村", "道路", "桥梁", "建筑", "风景", "自然", "环境",
		"学校", "老师", "学生", "课本", "知识", "文化", "艺术", "科学", "技术", "创新",
		"健康", "运动", "休息", "娱乐", "游戏", "电影", "书籍", "新闻", "信息", "网络",
		"父母", "孩子", "兄弟", "姐妹", "祖父", "祖母", "亲戚", "邻居", "同事", "伙伴",
	}

	japaneseWords := []string{
		"こんにちは", "ありがとう", "さようなら", "おはよう", "こんばんは", "すみません",
		"世界", "人生", "時間", "友達", "仕事", "勉強", "夢", "希望", "幸せ", "美しい",
		"桜", "雨", "花", "音楽", "物語", "旅行", "家族", "愛", "友情", "勇気",
		"知恵", "創造", "発見", "探索", "成長", "変化", "頑張る", "努力", "成功", "失敗",
		"経験", "思い出", "未来", "現在", "過去", "機会", "挑戦", "選択", "決定", "行動",
		"考える", "感じる", "体験", "楽しむ", "共有", "助ける", "支援", "理解", "包容", "許す",
		"春", "夏", "秋", "冬", "山", "川", "海", "空", "月", "太陽",
		"森", "草原", "都市", "田舎", "道", "橋", "建物", "風景", "自然", "環境",
	}

	englishWords := []string{
		"Hello", "Thank you", "Goodbye", "Good morning", "Good evening", "Excuse me",
		"World", "Life", "Time", "Friend", "Work", "Study", "Dream", "Hope", "Happy", "Beautiful",
		"Sunshine", "Rain", "Flower", "Music", "Story", "Travel", "Family", "Love", "Friendship", "Courage",
		"Wisdom", "Create", "Discover", "Explore", "Growth", "Change", "Persist", "Effort", "Success", "Failure",
		"Experience", "Memory", "Future", "Present", "Past", "Opportunity", "Challenge", "Choice", "Decision", "Action",
		"Think", "Feel", "Experience", "Enjoy", "Share", "Help", "Support", "Understand", "Accept", "Forgive",
		"Spring", "Summer", "Autumn", "Winter", "Mountain", "River", "Ocean", "Sky", "Moon", "Sun",
		"Forest", "Prairie", "City", "Village", "Road", "Bridge", "Building", "Scenery", "Nature", "Environment",
		"Amazing", "Wonderful", "Fantastic", "Incredible", "Brilliant", "Awesome", "Perfect", "Excellent", "Outstanding", "Remarkable",
		"Adventure", "Journey", "Discovery", "Innovation", "Creativity", "Inspiration", "Motivation", "Dedication", "Passion", "Achievement",
	}

	allWords := make([]string, 0, len(chineseWords)+len(japaneseWords)+len(englishWords))
	allWords = append(allWords, chineseWords...)
	allWords = append(allWords, japaneseWords...)
	allWords = append(allWords, englishWords...)

	styleEffects := []string{
		"",                             // 无效果
		"{\\b1}",                       // 粗体
		"{\\i1}",                       // 斜体
		"{\\b1\\i1}",                   // 粗斜体
		"{\\fn宋体}",                     // 字体变化
		"{\\fn黑体\\b1}",                 // 字体+粗体
		"{\\fn楷体\\i1}",                 // 字体+斜体
		"{\\fn微软雅黑\\b700}",             // 字体+字重
		"{\\c&H00FF00&}",               // 颜色变化
		"{\\c&HFF0000&\\b1}",           // 颜色+粗体
		"{\\pos(100,200)}",             // 位置
		"{\\fade(300,300)}",            // 淡入淡出
		"{\\blur2}",                    // 模糊
		"{\\bord2}",                    // 边框
		"{\\fscx120}",                  // 水平缩放
		"{\\fscy80}",                   // 垂直缩放
		"{\\frz15}",                    // 旋转
		"{\\fsp3}",                     // 字符间距
		"{\\rTitle}",                   // 样式重置
		"{\\fn思源黑体\\fs24\\c&HFF0000&}", // 复合样式
	}

	// 根据对话数量确定样式集合
	var styles []string
	var title string

	if dialogueCount <= 50 {
		title = "Small Test ASS"
		styles = []string{"Default", "Title"}
	} else if dialogueCount <= 500 {
		title = "Medium Test ASS"
		styles = []string{"Default", "Title", "Subtitle", "Comment"}
	} else {
		title = "Large Test ASS"
		styles = []string{"Default", "Title", "Subtitle", "Comment", "Narrator", "MainChar", "SideChar", "Song"}
	}

	// 构建头部内容
	content := fmt.Sprintf(`[Script Info]
Title: %s
ScriptType: v4.00+

[V4+ Styles]
Format: Name, Fontname, Fontsize, PrimaryColour, SecondaryColour, OutlineColour, BackColour, Bold, Italic, Underline, StrikeOut, ScaleX, ScaleY, Spacing, Angle, BorderStyle, Outline, Shadow, Alignment, MarginL, MarginR, MarginV, Encoding
`, title)

	// 基础样式定义
	content += "Style: Default,思源黑体,20,&H00FFFFFF,&H000000FF,&H00000000,&H80000000,0,0,0,0,100,100,0,0,1,2,0,2,10,10,10,1\n"
	content += "Style: Title,思源黑体,24,&H00FFFF00,&H000000FF,&H00000000,&H80000000,1,0,0,0,100,100,0,0,1,2,0,2,10,10,10,1\n"

	// 根据对话数量添加更多样式
	if dialogueCount > 50 {
		content += "Style: Subtitle,方正宋体,18,&H00CCCCCC,&H000000FF,&H00000000,&H80000000,0,1,0,0,100,100,0,0,1,2,0,2,10,10,10,1\n"
		content += "Style: Comment,华文楷体,16,&H0000FF00,&H000000FF,&H00000000,&H80000000,0,0,0,0,100,100,0,0,1,2,0,2,10,10,10,1\n"
	}

	if dialogueCount > 500 {
		content += "Style: Narrator,微软雅黑,22,&H00FFFFFF,&H000000FF,&H00000000,&H80000000,0,0,0,0,100,100,0,0,1,2,0,2,10,10,10,1\n"
		content += "Style: MainChar,华康手札体,20,&H0000FFFF,&H000000FF,&H00000000,&H80000000,1,0,0,0,100,100,0,0,1,2,0,2,10,10,10,1\n"
		content += "Style: SideChar,仿宋,18,&H00FF00FF,&H000000FF,&H00000000,&H80000000,0,0,0,0,100,100,0,0,1,2,0,2,10,10,10,1\n"
		content += "Style: Song,楷体,24,&H00FFFF00,&H000000FF,&H00000000,&H80000000,0,1,0,0,100,100,0,0,1,2,0,2,10,10,10,1\n"
	}

	content += "\n[Events]\nFormat: Layer, Start, End, Style, Name, MarginL, MarginR, MarginV, Effect, Text\n"

	// 生成随机对话
	for i := range dialogueCount {
		// 随机选择样式
		style := styles[i%len(styles)]

		// 生成随机对话内容
		wordCount := 3 + (i % 8) // 3-10个词
		var dialogueWords []string
		for j := 0; j < wordCount; j++ {
			dialogueWords = append(dialogueWords, allWords[(i*7+j)%len(allWords)])
		}

		// 随机添加样式效果 (30%概率)
		baseText := strings.Join(dialogueWords, "")
		var dialogue string
		if i%10 < 3 { // 30%概率添加样式效果
			effect := styleEffects[(i*3)%len(styleEffects)]
			if effect == "" {
				dialogue = baseText
			} else {
				// 随机决定效果位置
				if i%2 == 0 {
					dialogue = effect + baseText
				} else {
					// 在中间插入效果
					mid := len(dialogueWords) / 2
					firstPart := strings.Join(dialogueWords[:mid], "")
					secondPart := strings.Join(dialogueWords[mid:], "")
					dialogue = firstPart + effect + secondPart
				}
			}
		} else {
			dialogue = baseText
		}

		// 计算时间
		var startTime, endTime string
		if dialogueCount > 1000 {
			// 大规模使用小时格式
			hours := i / 3600
			minutes := (i % 3600) / 60
			seconds := i % 60
			endHours := hours
			endMinutes := minutes
			endSeconds := seconds + 3
			if endSeconds >= 60 {
				endMinutes++
				endSeconds -= 60
			}
			if endMinutes >= 60 {
				endHours++
				endMinutes -= 60
			}
			startTime = fmt.Sprintf("%d:%02d:%02d.00", hours, minutes, seconds)
			endTime = fmt.Sprintf("%d:%02d:%02d.00", endHours, endMinutes, endSeconds)
		} else {
			// 小中规模使用分钟格式
			minutes := i / 30
			seconds := (i % 30) * 2
			endMinutes := minutes
			endSeconds := seconds + 3
			if endSeconds >= 60 {
				endMinutes++
				endSeconds -= 60
			}
			startTime = fmt.Sprintf("0:%02d:%02d.00", minutes, seconds)
			endTime = fmt.Sprintf("0:%02d:%02d.00", endMinutes, endSeconds)
		}

		content += fmt.Sprintf("Dialogue: 0,%s,%s,%s,,0,0,0,,%s\n",
			startTime, endTime, style, dialogue)
	}

	return content
}

// BenchmarkNewASSParserSmall 测试小规模 ASS 文件解析性能 (50行对话)
func BenchmarkNewASSParserSmall(b *testing.B) {
	assContent := generateASSContent(50)

	b.ResetTimer()
	for b.Loop() {
		reader := strings.NewReader(assContent)
		parser, err := ass.NewASSParser(reader)
		if err != nil {
			b.Fatalf("NewASSParser failed: %v", err)
		}
		if len(parser.Contents) == 0 {
			b.Fatal("Parser should have contents")
		}
	}
}

// BenchmarkNewASSParserMedium 测试中规模 ASS 文件解析性能 (500行对话)
func BenchmarkNewASSParserMedium(b *testing.B) {
	assContent := generateASSContent(500)

	b.ResetTimer()
	for b.Loop() {
		reader := strings.NewReader(assContent)
		parser, err := ass.NewASSParser(reader)
		if err != nil {
			b.Fatalf("NewASSParser failed: %v", err)
		}
		if len(parser.Contents) == 0 {
			b.Fatal("Parser should have contents")
		}
	}
}

// BenchmarkNewASSParserLarge 测试大规模 ASS 文件解析性能 (5000行对话)
func BenchmarkNewASSParserLarge(b *testing.B) {
	assContent := generateASSContent(5000)

	b.ResetTimer()
	for b.Loop() {
		reader := strings.NewReader(assContent)
		parser, err := ass.NewASSParser(reader)
		if err != nil {
			b.Fatalf("NewASSParser failed: %v", err)
		}
		if len(parser.Contents) == 0 {
			b.Fatal("Parser should have contents")
		}
	}
}
