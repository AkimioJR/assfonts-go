package ass_test

import (
	"github/Akimio521/assfonts-go/ass"
	"testing"

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
			Dialogue: []string{"{Dialogue", "0", "0:00:00.00", "0:00:05.00", "Default", "", "0", "0", "0", `简单文本`},
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
			Dialogue: []string{"{Dialogue", "0", "0:00:00.00", "0:00:05.00", "Default", "", "0", "0", "0", `{\fn宋体\b1}重置前{\r}重置后`},
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
			Dialogue: []string{"{Dialogue", "0", "0:00:00.00", "0:00:05.00", "style1", "", "0", "0", "0", `{\rstyle2}指定样式`},
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
			Dialogue: []string{"{Dialogue", "0", "0:00:00.00", "0:00:05.00", "Default", "", "0", "0", "0", `转义\{字符\}测试`},
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
			Dialogue: []string{"{Dialogue", "0", "0:00:00.00", "0:00:05.00", "Default", "", "0", "0", "0", `{\fn宋体}重置前{\r}\{重置后\}`},
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
			Dialogue: []string{"{Dialogue", "0", "0:00:00.00", "0:00:05.00", "style1", "", "0", "0", "0", `{\fnA\b1}粗体A{\rstyle2\i1}斜体B{\r}普通`},
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
			Dialogue: []string{"{Dialogue", "0", "0:00:00.00", "0:00:05.00", "style1", "", "0", "0", "0", `特殊字符: \n \h \{ \} \\`},
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
			Dialogue: []string{"{Dialogue", "0", "0:00:00.00", "0:00:10.00", "style1", "", "0", "0", "0", `开始{\fnA}样式A{\fnB\b1}样式B{\r}重置{\rstyle2}样式2{\r}结束`},
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
			Dialogue: []string{"{Dialogue", "0", "0:00:00.00", "0:00:05.00", "style1", "", "0", "0", "0", `开始{}结束`},
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
			Dialogue: []string{"{Dialogue", "0", "0:00:00.00", "0:00:05.00", "style1", "", "0", "0", "0", `中文Chinese 日本語한국어`},
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
			Content:  nil,
			Dialogue: []string{"{ Dialogue", "0", "0:00:00.88", "0:59:59.82", "style1", "", "0", "0", "0", `我{你甚至可以在这里写注释\CODE_HERE\前面可以是一段代码，但无需关注}能{这里是\rndx10}吞下{\fn宋体\b1\i1}玻璃而{\pos(400,400)}不{\r}伤身{\rstyle2}体\{这是转义的\n括号\}`},
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
					StyleNameFontDesc: map[string]ass.FontDesc{
						"SongTi": {
							FontName: "宋体",
							Bold:     700,
							Italic:   50,
						},
					},
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
					FontSets:          make(map[ass.FontDesc]ass.CodepointSet),
					StyleNameFontDesc: c.fd,
				}
				err := ap.ParseDialogue(&c.d)
				require.NoError(t, err)
				require.Equal(t, c.expect, ap.FontSets)
			},
		)

	}

}
