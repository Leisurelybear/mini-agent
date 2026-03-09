package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FileOperationTool 文件操作工具
type FileOperationTool struct {
	workDir string
}

func NewFileOperationTool(workDir string) *FileOperationTool {
	if workDir == "" {
		workDir = "/tmp/agent-workspace"
	}
	os.MkdirAll(workDir, 0755)
	return &FileOperationTool{workDir: workDir}
}

func (t *FileOperationTool) Name() string {
	return "file_ops"
}

func (t *FileOperationTool) Description() string {
	return `文件操作工具。支持读取、写入、列出文件。
输入格式：命令|参数
命令：
- read|filepath: 读取文件内容
- write|filepath|content: 写入文件内容
- list|dirpath: 列出目录内容
- delete|filepath: 删除文件
示例：
read|test.txt
write|output.txt|Hello World
list|.
delete|temp.txt`
}

func (t *FileOperationTool) Execute(input string) (string, error) {
	parts := strings.SplitN(input, "|", 3)
	if len(parts) < 2 {
		return "", fmt.Errorf("输入格式错误。期望：命令|参数")
	}

	command := strings.TrimSpace(parts[0])

	switch command {
	case "read":
		return t.readFile(parts[1])
	case "write":
		if len(parts) < 3 {
			return "", fmt.Errorf("write命令需要：write|filepath|content")
		}
		return t.writeFile(parts[1], parts[2])
	case "list":
		return t.listDir(parts[1])
	case "delete":
		return t.deleteFile(parts[1])
	default:
		return "", fmt.Errorf("未知命令: %s", command)
	}
}

func (t *FileOperationTool) readFile(path string) (string, error) {
	fullPath := t.resolvePath(path)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return "", fmt.Errorf("读取文件失败: %v", err)
	}
	return string(content), nil
}

func (t *FileOperationTool) writeFile(path, content string) (string, error) {
	fullPath := t.resolvePath(path)
	err := os.WriteFile(fullPath, []byte(content), 0644)
	if err != nil {
		return "", fmt.Errorf("写入文件失败: %v", err)
	}
	return fmt.Sprintf("成功写入 %d 字节到 %s", len(content), fullPath), nil
}

func (t *FileOperationTool) listDir(path string) (string, error) {
	fullPath := t.resolvePath(path)
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return "", fmt.Errorf("列出目录失败: %v", err)
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("目录 %s 的内容:\n", fullPath))
	for _, entry := range entries {
		info, _ := entry.Info()
		fileType := "文件"
		if entry.IsDir() {
			fileType = "目录"
		}
		result.WriteString(fmt.Sprintf("- %s (%s, %d 字节)\n",
			entry.Name(), fileType, info.Size()))
	}
	return result.String(), nil
}

func (t *FileOperationTool) deleteFile(path string) (string, error) {
	fullPath := t.resolvePath(path)
	err := os.Remove(fullPath)
	if err != nil {
		return "", fmt.Errorf("删除文件失败: %v", err)
	}
	return fmt.Sprintf("成功删除文件: %s", fullPath), nil
}

func (t *FileOperationTool) resolvePath(path string) string {
	path = strings.TrimSpace(path)
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(t.workDir, path)
}
