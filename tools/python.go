package tools

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// PythonTool 执行Python代码的工具
type PythonTool struct {
	timeout time.Duration
}

func NewPythonTool(timeout time.Duration) *PythonTool {
	return &PythonTool{timeout: timeout}
}

func (t *PythonTool) Name() string {
	return "python"
}

func (t *PythonTool) Description() string {
	return `执行Python代码。可以用于数据处理、计算、文件操作等。
输入格式：直接提供Python代码
示例：
print(sum([1, 2, 3, 4, 5]))
import json
data = {"name": "test"}
print(json.dumps(data))`
}

func (t *PythonTool) Execute(input string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), t.timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "python3", "-c", input)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	if ctx.Err() == context.DeadlineExceeded {
		return "", fmt.Errorf("Python执行超时（%v）", t.timeout)
	}

	if err != nil {
		return "", fmt.Errorf("执行错误: %v\nStderr: %s", err, stderr.String())
	}

	output := stdout.String()
	if stderr.Len() > 0 {
		output += "\n[警告]: " + stderr.String()
	}

	return strings.TrimSpace(output), nil
}
