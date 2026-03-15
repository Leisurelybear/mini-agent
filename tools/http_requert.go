package tools

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// HTTPRequestTool HTTP请求工具
type HTTPRequestTool struct {
	Timeout time.Duration // HTTP请求超时时间
}

// NewHTTPRequestTool 创建HTTP请求工具，默认超时30秒
func NewHTTPRequestTool(timeout ...time.Duration) *HTTPRequestTool {
	tool := &HTTPRequestTool{
		Timeout: 30 * time.Second, // 默认超时
	}
	if len(timeout) > 0 && timeout[0] > 0 {
		tool.Timeout = timeout[0]
	}
	return tool
}

func (t *HTTPRequestTool) Name() string {
	return "http_request"
}

func (t *HTTPRequestTool) Description() string {
	return `发送HTTP请求工具。
输入格式：请求方法|URL|请求体（GET请求不需要请求体）
支持的请求方法：GET/POST/PUT/DELETE
使用说明：
- 请求体为JSON格式时请确保格式正确
- 超时时间：` + t.Timeout.String() + `
示例：
输入: GET|https://httpbin.org/get|
输出: httpbin返回的GET请求响应
输入: POST|https://httpbin.org/post|{"name":"agent","age":18}
输出: httpbin返回的POST请求响应`
}

func (t *HTTPRequestTool) Execute(input string) (string, error) {
	parts := strings.SplitN(input, "|", 3)
	if len(parts) < 2 {
		return "", errors.New("输入格式错误，正确格式：请求方法|URL|请求体")
	}

	method := strings.TrimSpace(strings.ToUpper(parts[0]))
	url := strings.TrimSpace(parts[1])
	body := ""
	if len(parts) == 3 {
		body = parts[2]
	}

	// 创建请求
	var req *http.Request
	var err error
	if method == "GET" || body == "" {
		req, err = http.NewRequest(method, url, nil)
	} else {
		req, err = http.NewRequest(method, url, strings.NewReader(body))
	}
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置Header
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "AgentTool/1.0")

	// 带超时的客户端
	client := &http.Client{Timeout: t.Timeout}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	result := fmt.Sprintf("状态码: %d\n响应内容:\n%s", resp.StatusCode, string(respBody))
	return result, nil
}
