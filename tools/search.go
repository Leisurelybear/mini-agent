package tools

import (
	"time"
)

// SearchTool 网络搜索工具
type SearchTool struct {
	timeout time.Duration
	apiKey  string // 如果需要API密钥
}

func NewSearchTool(timeout time.Duration, apiKey string) *SearchTool {
	return &SearchTool{
		timeout: timeout,
		apiKey:  apiKey,
	}
}

func (t *SearchTool) Name() string {
	return "search"
}

func (t *SearchTool) Description() string {
	return `网络搜索工具。用于查找最新信息、事实核查等。
输入格式：搜索查询字符串
示例：
Go语言最新版本
2024年人工智能发展趋势`
}

func (t *SearchTool) Execute(input string) (string, error) {
	// TODO

	return "To be implemented", nil
}
