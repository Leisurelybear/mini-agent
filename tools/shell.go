package tools

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// CommandExecutorTool 执行系统命令工具
type CommandExecutorTool struct {
	Timeout time.Duration // 命令执行超时时间
}

// NewCommandExecutorTool 创建命令执行工具，默认超时10秒
func NewCommandExecutorTool(timeout ...time.Duration) *CommandExecutorTool {
	tool := &CommandExecutorTool{
		Timeout: 10 * time.Second, // 默认超时
	}
	if len(timeout) > 0 && timeout[0] > 0 {
		tool.Timeout = timeout[0]
	}
	return tool
}

func (t *CommandExecutorTool) Name() string {
	return "command_executor"
}

func (t *CommandExecutorTool) Description() string {
	return `执行系统命令（跨平台兼容）。
输入格式：需要执行的完整命令
使用说明：
- Windows系统会自动使用cmd.exe执行，Linux/macOS使用bash
- 避免执行危险命令（如rm -rf、format等）
- 超时时间：` + t.Timeout.String() + `
示例：
输入: dir
输出: 列出当前目录文件（Windows）
输入: ls -l
输出: 列出当前目录文件（Linux/macOS）
输入: echo "hello agent"
输出: hello agent`
}

func (t *CommandExecutorTool) Execute(input string) (string, error) {
	if strings.TrimSpace(input) == "" {
		return "", errors.New("命令不能为空")
	}

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", input)
	} else {
		cmd = exec.Command("bash", "-c", input)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// 启动命令
	err := cmd.Start()
	if err != nil {
		return "", fmt.Errorf("启动命令失败: %w", err)
	}

	// 超时控制
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case err = <-done:
	case <-time.After(t.Timeout):
		// 超时杀死进程
		if errKill := cmd.Process.Kill(); errKill != nil {
			return "", fmt.Errorf("命令执行超时（%s），且杀死进程失败: %w", t.Timeout, errKill)
		}
		return "", fmt.Errorf("命令执行超时（%s）", t.Timeout)
	}

	output := stdout.String()
	if stderr.Len() > 0 {
		output += "\n错误信息: " + stderr.String()
	}

	if err != nil {
		return output, fmt.Errorf("命令执行失败: %w", err)
	}

	return output, nil
}
