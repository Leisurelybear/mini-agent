package tools

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// PythonRunnerTool 运行Python代码工具
type PythonRunnerTool struct {
	Timeout time.Duration // Python执行超时时间
}

// NewPythonRunnerTool 创建Python运行工具，默认超时15秒
func NewPythonRunnerTool(timeout ...time.Duration) *PythonRunnerTool {
	tool := &PythonRunnerTool{
		Timeout: 15 * time.Second, // 默认超时
	}
	if len(timeout) > 0 && timeout[0] > 0 {
		tool.Timeout = timeout[0]
	}
	return tool
}

func (t *PythonRunnerTool) Name() string {
	return "python_runner"
}

func (t *PythonRunnerTool) Description() string {
	return `运行Python代码片段。
输入格式：完整的Python代码
使用说明：
- 确保系统已安装Python并添加到环境变量
- 支持单行和多行代码（多行用换行分隔）
- 超时时间：` + t.Timeout.String() + `
示例：
输入: print("Hello Python!")
输出: Hello Python!
输入: 
a = 10
b = 20
print(a + b)
输出: 30`
}

func (t *PythonRunnerTool) Execute(input string) (string, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "", errors.New("Python代码不能为空")
	}

	// 创建临时文件
	tmpFile, err := os.CreateTemp("", "agent_python_*.py")
	if err != nil {
		return "", fmt.Errorf("创建临时文件失败: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(input)
	if err != nil {
		tmpFile.Close()
		return "", fmt.Errorf("写入Python代码失败: %w", err)
	}
	tmpFile.Close()

	// 执行Python
	cmd := exec.Command("python", tmpFile.Name())
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Start()
	if err != nil {
		return "", fmt.Errorf("启动Python失败: %w", err)
	}

	// 超时控制
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case err = <-done:
	case <-time.After(t.Timeout):
		if errKill := cmd.Process.Kill(); errKill != nil {
			return "", fmt.Errorf("Python执行超时（%s），且杀死进程失败: %w", t.Timeout, errKill)
		}
		return "", fmt.Errorf("Python执行超时（%s）", t.Timeout)
	}

	output := stdout.String()
	if stderr.Len() > 0 {
		output += "\n错误信息: " + stderr.String()
	}

	if err != nil {
		return output, fmt.Errorf("Python执行失败: %w", err)
	}

	return output, nil
}
