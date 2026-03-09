package tools

import (
	"time"
)

// WeatherTool 天气查询工具
type WeatherTool struct {
	timeout time.Duration // 请求超时时间
}

// NewWeatherTool 创建天气查询工具实例
func NewWeatherTool(timeout time.Duration) *WeatherTool {
	return &WeatherTool{timeout: timeout}
}

// Name 工具名称
func (t *WeatherTool) Name() string {
	return "weather"
}

// Description 工具描述（包含使用说明和示例）
func (t *WeatherTool) Description() string {
	return `查询指定城市的实时天气信息（支持国内城市，如北京、上海、广州）。
输入格式：城市名称
示例：
Beijing
Shanghai
深圳`
}

// Execute 执行天气查询
func (t *WeatherTool) Execute(input string) (string, error) {
	// TODO
	return "To be implemented", nil
}
