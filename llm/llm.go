package llm

import (
	"encoding/json"
	"fmt"
	"strings"

	"mini_agent/types"
)

// LLM 大语言模型接口
type LLM interface {
	BindToolsDesc(toolsDesc string) // 所有LLM必须实现该方法
	Chat(messages []types.Message) (string, error)
	AnalyzeTask(query string, taskHistory []types.Task) (*types.AgentDecision, error)
}

// MockLLM 模拟LLM（用于测试）
type MockLLM struct {
	model string
}

func NewMockLLM(model string) *MockLLM {
	return &MockLLM{model: model}
}

func (l *MockLLM) BindToolsDesc(toolsDesc string) {

}

func (l *MockLLM) Chat(messages []types.Message) (string, error) {
	// 模拟简单的响应逻辑
	if len(messages) == 0 {
		return "你好！我是AI助手，有什么可以帮你的吗？", nil
	}

	lastMsg := messages[len(messages)-1]
	content := strings.ToLower(lastMsg.Content)

	if strings.Contains(content, "计算") || strings.Contains(content, "算") {
		return "这需要进行数学计算。", nil
	} else if strings.Contains(content, "搜索") || strings.Contains(content, "查询") {
		return "这需要进行网络搜索。", nil
	} else if strings.Contains(content, "python") || strings.Contains(content, "代码") {
		return "这需要执行Python代码。", nil
	}

	return "我理解了你的问题。让我来帮你处理。", nil
}

func (l *MockLLM) AnalyzeTask(query string, taskHistory []types.Task) (*types.AgentDecision, error) {
	// 模拟任务分析逻辑
	query = strings.ToLower(query)

	decision := &types.AgentDecision{
		Tasks:     []types.ToolCall{},
		Reasoning: "分析用户查询并决定下一步行动",
	}

	// 简单的规则匹配
	if strings.Contains(query, "计算") || strings.Contains(query, "多少") {
		if strings.Contains(query, "+") || strings.Contains(query, "-") ||
			strings.Contains(query, "*") || strings.Contains(query, "/") {
			decision.NeedMoreWork = true
			decision.Tasks = append(decision.Tasks, types.ToolCall{
				ToolName: "calculator",
				Input:    extractMathExpression(query),
				Reason:   "执行数学计算",
			})
		}
	}

	if strings.Contains(query, "搜索") || strings.Contains(query, "查找") ||
		strings.Contains(query, "最新") {
		decision.NeedMoreWork = true
		decision.Tasks = append(decision.Tasks, types.ToolCall{
			ToolName: "search",
			Input:    query,
			Reason:   "搜索相关信息",
		})
	}

	if strings.Contains(query, "python") || strings.Contains(query, "运行代码") {
		decision.NeedMoreWork = true
		decision.Tasks = append(decision.Tasks, types.ToolCall{
			ToolName: "python",
			Input:    extractPythonCode(query),
			Reason:   "执行Python代码",
		})
	}

	if strings.Contains(query, "文件") && (strings.Contains(query, "读取") ||
		strings.Contains(query, "写入") || strings.Contains(query, "创建")) {
		decision.NeedMoreWork = true
		decision.Tasks = append(decision.Tasks, types.ToolCall{
			ToolName: "file_ops",
			Input:    query,
			Reason:   "文件操作",
		})
	}

	// 如果有任务历史，检查是否需要继续
	if len(taskHistory) > 0 {
		// 检查最后一个任务
		lastTask := taskHistory[len(taskHistory)-1]
		if lastTask.Status == "completed" && decision.NeedMoreWork == false {
			decision.FinalAnswer = fmt.Sprintf("任务完成。结果：%s", lastTask.Output)
			decision.NeedMoreWork = false
		}
	}

	// 如果没有识别出需要的工具，直接回答
	if len(decision.Tasks) == 0 && len(taskHistory) == 0 {
		decision.FinalAnswer = "我已经理解了你的问题。" + l.generateDirectAnswer(query)
		decision.NeedMoreWork = false
	}

	return decision, nil
}

func extractMathExpression(query string) string {
	// 简单提取数学表达式
	for _, op := range []string{"+", "-", "*", "/"} {
		if idx := strings.Index(query, op); idx != -1 {
			// 尝试提取表达式
			parts := strings.Fields(query)
			for i, part := range parts {
				if strings.Contains(part, op) {
					if i > 0 && i < len(parts)-1 {
						return parts[i-1] + " " + op + " " + parts[i+1]
					}
				}
			}
		}
	}
	return query
}

func extractPythonCode(query string) string {
	// 提取代码块
	if idx := strings.Index(query, "```python"); idx != -1 {
		end := strings.Index(query[idx+9:], "```")
		if end != -1 {
			return strings.TrimSpace(query[idx+9 : idx+9+end])
		}
	}
	return query
}

func (l *MockLLM) generateDirectAnswer(query string) string {
	if strings.Contains(query, "你好") || strings.Contains(query, "hello") {
		return "你好！我是AI助手，可以帮你执行计算、搜索、运行代码等任务。"
	}
	if strings.Contains(query, "帮助") || strings.Contains(query, "help") {
		return `我可以帮你：
1. 数学计算（使用calculator工具）
2. 网络搜索（使用search工具）
3. 执行Python代码（使用python工具）
4. 执行Shell命令（使用shell工具）
5. 文件操作（使用file_ops工具）`
	}
	return "如果需要具体帮助，请告诉我更多细节。"
}

// 辅助函数：将决策转换为JSON字符串（用于调试）
func DecisionToJSON(decision *types.AgentDecision) string {
	data, _ := json.MarshalIndent(decision, "", "  ")
	return string(data)
}
