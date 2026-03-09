package tools

import (
	"time"
)

// StockTool 股票查询工具
type StockTool struct {
	timeout time.Duration // 请求超时时间
}

// NewStockTool 创建股票查询工具实例
func NewStockTool(timeout time.Duration) *StockTool {
	return &StockTool{timeout: timeout}
}

// Name 工具名称
func (t *StockTool) Name() string {
	return "stock"
}

// Description 工具描述（包含使用说明和示例）
func (t *StockTool) Description() string {
	return `查询A股股票实时行情（仅支持股票代码，不支持股票名称）。
输入格式：股票代码，纯数字
示例：
600000
000001
300059`
}

// 新浪财经API返回的结构体（仅保留核心字段）
type stockResponse struct {
	Code string `json:"code"` // 状态码，0表示成功
	Data struct {
		Price     string `json:"price"`         // 当前价格
		Open      string `json:"open"`          // 今日开盘价
		Close     string `json:"close"`         // 昨日收盘价
		High      string `json:"high"`          // 今日最高价
		Low       string `json:"low"`           // 今日最低价
		Change    string `json:"change"`        // 涨跌额
		ChangePct string `json:"changepercent"` // 涨跌幅（%）
		Name      string `json:"name"`          // 股票名称
	} `json:"data"`
}

// Execute 执行股票查询
func (t *StockTool) Execute(input string) (string, error) {
	// TODO
	return "To be implemented", nil
}
