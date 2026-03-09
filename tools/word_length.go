package tools

import "strconv"

// WordLengthTool 字符长度工具
type WordLengthTool struct {
}

// NewWordLengthTool ...
func NewWordLengthTool() *WordLengthTool {
	return &WordLengthTool{}
}

// Name 工具名称
func (t *WordLengthTool) Name() string {
	return "word_length"
}

// Description 工具描述（包含使用说明和示例）
func (t *WordLengthTool) Description() string {
	return `查询字符输入的长度。
输入格式：城市名称
示例：
我是一个工具
Hello world!
12345678`
}

// Execute 执行天气查询
func (t *WordLengthTool) Execute(input string) (string, error) {
	return strconv.Itoa(len(input)), nil
}
