package tools

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// JSONTool JSON解析和格式化工具
type JSONTool struct{}

func NewJSONTool() *JSONTool {
	return &JSONTool{}
}

func (t *JSONTool) Name() string {
	return "json_tool"
}

func (t *JSONTool) Description() string {
	return `JSON格式化和解析工具。
输入格式：操作类型|JSON字符串
支持的操作类型：
- format: 格式化JSON字符串（便于阅读）
- extract: 提取JSON中的字段（格式：extract|JSON字符串|字段路径，字段路径用.分隔）
示例：
输入: format|{"name":"agent","age":18}
输出: 
{
  "name": "agent",
  "age": 18
}
输入: extract|{"name":"agent","age":18}|name
输出: agent`
}

func (t *JSONTool) Execute(input string) (string, error) {
	parts := strings.SplitN(input, "|", 3)
	if len(parts) < 2 {
		return "", errors.New("输入格式错误，正确格式：操作类型|JSON字符串|字段路径（extract需要）")
	}

	opType := strings.TrimSpace(parts[0])
	jsonStr := strings.TrimSpace(parts[1])

	if jsonStr == "" {
		return "", errors.New("JSON字符串不能为空")
	}

	var data interface{}
	err := json.Unmarshal([]byte(jsonStr), &data)
	if err != nil {
		return "", fmt.Errorf("JSON解析失败: %w", err)
	}

	switch opType {
	case "format":
		formatted, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return "", fmt.Errorf("JSON格式化失败: %w", err)
		}
		return string(formatted), nil

	case "extract":
		if len(parts) < 3 {
			return "", errors.New("extract操作需要指定字段路径，格式：extract|JSON字符串|字段路径")
		}
		fieldPath := strings.TrimSpace(parts[2])
		if fieldPath == "" {
			return "", errors.New("字段路径不能为空")
		}

		value, err := extractJSONField(data, fieldPath)
		if err != nil {
			return "", fmt.Errorf("提取字段失败: %w", err)
		}
		result, err := json.Marshal(value)
		if err != nil {
			return fmt.Sprintf("%v", value), nil
		}
		return string(result), nil

	default:
		return "", fmt.Errorf("不支持的操作类型: %s，支持的类型：format/extract", opType)
	}
}

// extractJSONField 递归提取JSON字段
func extractJSONField(data interface{}, fieldPath string) (interface{}, error) {
	fields := strings.Split(fieldPath, ".")
	current := data

	for _, field := range fields {
		switch v := current.(type) {
		case map[string]interface{}:
			var ok bool
			current, ok = v[field]
			if !ok {
				return nil, fmt.Errorf("字段 %s 不存在", field)
			}
		case []interface{}:
			index, err := strconv.Atoi(field)
			if err != nil {
				return nil, fmt.Errorf("数组索引必须是数字: %s", field)
			}
			if index < 0 || index >= len(v) {
				return nil, fmt.Errorf("数组索引 %d 越界", index)
			}
			current = v[index]
		default:
			return nil, fmt.Errorf("字段 %s 不是对象或数组", field)
		}
	}

	return current, nil
}
