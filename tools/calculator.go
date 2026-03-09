package tools

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

// CalculatorTool 计算器工具
type CalculatorTool struct{}

func NewCalculatorTool() *CalculatorTool {
	return &CalculatorTool{}
}

func (t *CalculatorTool) Name() string {
	return "calculator"
}

func (t *CalculatorTool) Description() string {
	return `数学计算器。支持基本算术运算和数学函数。
输入格式：数学表达式（使用Python语法）
示例：
2 + 3 * 4
sqrt(16)
pow(2, 10)
sin(3.14159/2)`
}

func (t *CalculatorTool) Execute(input string) (string, error) {
	// 为了安全和简单，使用Python来执行数学表达式
	pythonCode := fmt.Sprintf(`
from math import *
try:
    result = %s
    print(result)
except Exception as e:
    print(f"计算错误: {e}")
`, input)

	pythonTool := NewPythonTool(5 * time.Second)
	return pythonTool.Execute(pythonCode)
}

// SimpleMathTool 简单数学工具（无需Python）
type SimpleMathTool struct{}

func NewSimpleMathTool() *SimpleMathTool {
	return &SimpleMathTool{}
}

func (t *SimpleMathTool) Name() string {
	return "simple_math"
}

func (t *SimpleMathTool) Description() string {
	return `简单计算器。仅支持基本算术：加减乘除。
输入格式：数字 运算符 数字
示例：
10 + 5
20 * 3
100 / 4`
}

func (t *SimpleMathTool) Execute(input string) (string, error) {
	parts := strings.Fields(input)
	if len(parts) != 3 {
		return "", fmt.Errorf("输入格式错误。期望：数字 运算符 数字")
	}

	a, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return "", fmt.Errorf("无效的第一个数字: %v", err)
	}

	b, err := strconv.ParseFloat(parts[2], 64)
	if err != nil {
		return "", fmt.Errorf("无效的第二个数字: %v", err)
	}

	var result float64
	switch parts[1] {
	case "+":
		result = a + b
	case "-":
		result = a - b
	case "*":
		result = a * b
	case "/":
		if b == 0 {
			return "", fmt.Errorf("除数不能为零")
		}
		result = a / b
	case "^", "**":
		result = math.Pow(a, b)
	default:
		return "", fmt.Errorf("不支持的运算符: %s", parts[1])
	}

	return fmt.Sprintf("%.6f", result), nil
}
