# ASS字体工具 (assfonts-go)
**一个用于处理 ASS 字幕文件字体的 Go 语言工具库，支持字体子集化和嵌入功能**

 assfonts-go 可以搜索 ASS 字幕文件所需的字体并进行子集化处理。它还可以将 UUEncode 编码的字体文件直接嵌入到字幕脚本中（这是 ASS 字幕格式的一个特性）。通过这种方式，一个字幕文件可以包含所有渲染所需的信息，并实现类似图形字幕（如 PGS）的效果。

## 功能特性
- 🎯 **ASS字幕解析**：完整解析 ASS 字幕文件格式，包括样式和对话内容
- 🔤 **字体子集化**：提取字幕中实际使用的字符，生成精简字体
- 💾 **字体嵌入**：将字体数据嵌入到 ASS 字幕文件中
- 🗄️ **字体数据库**：构建和管理系统字体数据库，支持持久化
- 🖥️ **跨平台支持**：支持 macOS、Linux 和 Windows 系统
- ⚡ **并发处理**：支持多线程并发处理，提高性能
- 📝 **SRT支持**：支持 SRT 字幕格式的解析和处理

## 使用方法
### 库函数
参考 [main.go](./cmd/assfont-go/main.go)

### 二进制程序使用
1. 安装依赖
   - pkg-config
   - freetype
   - hb-subset

2. 克隆项目
```shell
git clone https://github.com/AkimioJR/assfonts-go.git
cd assfonts-go
go mod download
```

3. 编译
```shell
go build cmd/assfont-go/main.go -o assfont-go
```

4. 查看完整使用方法
```shell
./assfont-go -h
```

5. 子集化并嵌入 ASS 字幕
```shell
./assfont-go -input <输入 ASS 字幕路径> -output  <输出 ASS 字幕路径>
```

## 鸣谢
- [wyzdwdz/assfonts](https://github.com/wyzdwdz/assfonts)
- [freetype](https://freetype.org/)
- [harfbuzz](https://github.com/harfbuzz/harfbuzz)