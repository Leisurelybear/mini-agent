package tools

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// ShellTool 执行Shell命令的工具
type ShellTool struct {
	timeout time.Duration
}

func NewShellTool(timeout time.Duration) *ShellTool {
	return &ShellTool{timeout: timeout}
}

func (t *ShellTool) Name() string {
	return "shell"
}

func (t *ShellTool) Description() string {
	return `执行Shell命令。可以用于文件操作、系统信息查询等。
输入格式：Shell命令，注意本机系统，可能需要使用不同的命令
示例：
dir
ls -la
cat /etc/os-release
echo "Hello World"`
}

func (t *ShellTool) Execute(input string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), t.timeout)
	defer cancel()

	// 使用 bash -c 执行命令
	cmd := exec.CommandContext(ctx, "bash", "-c", input)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	if ctx.Err() == context.DeadlineExceeded {
		return "", fmt.Errorf("命令执行超时（%v）", t.timeout)
	}

	if err != nil {
		return "", fmt.Errorf("执行错误: %v\nStderr: %s", err, stderr.String())
	}

	output := stdout.String()
	if stderr.Len() > 0 {
		output += "\n[Stderr]: " + stderr.String()
	}

	return strings.TrimSpace(output), nil
}
