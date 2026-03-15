package tools

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// DateTimeTool 日期时间处理工具
type DateTimeTool struct{}

func NewDateTimeTool() *DateTimeTool {
	return &DateTimeTool{}
}

func (t *DateTimeTool) Name() string {
	return "datetime_tool"
}

func (t *DateTimeTool) Description() string {
	return `日期时间处理工具。
输入格式：操作类型|参数（可选）
支持的操作类型：
- now: 获取当前时间（参数为格式字符串，默认：2006-01-02 15:04:05）
- parse: 解析时间字符串（参数：时间字符串|格式字符串）
- calc: 时间计算（参数：基准时间|格式|偏移量，偏移量格式：+1h/-2d/30m）
示例：
输入: now|
输出: 2024-05-20 14:30:00
输入: now|2006/01/02
输出: 2024/05/20
输入: parse|2024-05-20|2006-01-02
输出: 2024-05-20 00:00:00
输入: calc|2024-05-20 12:00:00|2006-01-02 15:04:05|+1h30m
输出: 2024-05-20 13:30:00`
}

func (t *DateTimeTool) Execute(input string) (string, error) {
	parts := strings.SplitN(input, "|", 3)
	if len(parts) < 1 {
		return "", errors.New("输入格式错误，正确格式：操作类型|参数")
	}

	opType := strings.TrimSpace(parts[0])
	param := ""
	if len(parts) >= 2 {
		param = strings.TrimSpace(parts[1])
	}

	switch opType {
	case "now":
		format := "2006-01-02 15:04:05"
		if param != "" {
			format = param
		}
		return time.Now().Format(format), nil

	case "parse":
		parseParts := strings.SplitN(param, "|", 2)
		if len(parseParts) < 2 {
			return "", errors.New("parse操作参数错误，格式：时间字符串|格式字符串")
		}
		timeStr := strings.TrimSpace(parseParts[0])
		format := strings.TrimSpace(parseParts[1])
		parsedTime, err := time.Parse(format, timeStr)
		if err != nil {
			return "", fmt.Errorf("解析时间失败: %w", err)
		}
		return parsedTime.Format("2006-01-02 15:04:05"), nil

	case "calc":
		calcParts := strings.SplitN(param, "|", 3)
		if len(calcParts) < 3 {
			return "", errors.New("calc操作参数错误，格式：基准时间|格式|偏移量")
		}
		baseTimeStr := strings.TrimSpace(calcParts[0])
		format := strings.TrimSpace(calcParts[1])
		offset := strings.TrimSpace(calcParts[2])

		baseTime, err := time.Parse(format, baseTimeStr)
		if err != nil {
			return "", fmt.Errorf("解析基准时间失败: %w", err)
		}

		duration, err := parseDuration(offset)
		if err != nil {
			return "", fmt.Errorf("解析偏移量失败: %w", err)
		}

		resultTime := baseTime.Add(duration)
		return resultTime.Format(format), nil

	default:
		return "", fmt.Errorf("不支持的操作类型: %s，支持的类型：now/parse/calc", opType)
	}
}

// parseDuration 解析自定义偏移量（+1h/-2d/30m）
func parseDuration(s string) (time.Duration, error) {
	s = strings.ReplaceAll(s, "d", "h*24")
	s = strings.ReplaceAll(s, "m", "m")
	s = strings.ReplaceAll(s, "h", "h")
	return time.ParseDuration(s)
}
