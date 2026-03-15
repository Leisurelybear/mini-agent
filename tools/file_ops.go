package tools

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FileOperationTool 文件操作工具
// FileOperatorTool 文件操作工具（读/写/追加）
type FileOperatorTool struct{}

func NewFileOperatorTool() *FileOperatorTool {
	return &FileOperatorTool{}
}

func (t *FileOperatorTool) Name() string {
	return "file_operator"
}

func (t *FileOperatorTool) Description() string {
	return `文件读写操作工具。
输入格式：操作类型|文件路径|内容（内容仅写/追加操作需要）
支持的操作类型：
- read: 读取文件内容
- write: 写入文件（覆盖原有内容）
- append: 追加内容到文件末尾
使用说明：
- 文件路径支持绝对路径和相对路径
- 路径中的特殊字符需要正确转义
示例：
输入: read|./test.txt|
输出: test.txt文件的内容
输入: write|./test.txt|Hello File!
输出: 文件写入成功
输入: append|./test.txt|New Line
输出: 文件追加成功`
}

func (t *FileOperatorTool) Execute(input string) (string, error) {
	parts := strings.SplitN(input, "|", 3)
	if len(parts) < 2 {
		return "", errors.New("输入格式错误，正确格式：操作类型|文件路径|内容")
	}

	opType := strings.TrimSpace(parts[0])
	filePath := strings.TrimSpace(parts[1])
	content := ""
	if len(parts) == 3 {
		content = parts[2]
	}

	filePath = filepath.Clean(filePath)

	switch opType {
	case "read":
		data, err := os.ReadFile(filePath)
		if err != nil {
			return "", fmt.Errorf("读取文件失败: %w", err)
		}
		return string(data), nil

	case "write":
		err := os.WriteFile(filePath, []byte(content), 0644)
		if err != nil {
			return "", fmt.Errorf("写入文件失败: %w", err)
		}
		return fmt.Sprintf("文件 %s 写入成功", filePath), nil

	case "append":
		f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return "", fmt.Errorf("打开文件失败: %w", err)
		}
		defer f.Close()

		_, err = f.WriteString(content)
		if err != nil {
			return "", fmt.Errorf("追加内容失败: %w", err)
		}
		return fmt.Sprintf("内容已追加到文件 %s", filePath), nil

	default:
		return "", fmt.Errorf("不支持的操作类型: %s，支持的类型：read/write/append", opType)
	}
}
