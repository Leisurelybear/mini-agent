package agent

import (
	"fmt"
	"mini_agent/types"
	"sync"
	"time"

	"mini_agent/llm"
)

// Agent AI代理
type Agent struct {
	config   types.AgentConfig
	llm      llm.LLM
	tools    map[string]types.Tool
	toolsMux sync.RWMutex
}

// NewAgent 创建新的Agent
func NewAgent(config types.AgentConfig, llmInstance llm.LLM) *Agent {
	if config.MaxIterations <= 0 {
		config.MaxIterations = 5
	}
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	return &Agent{
		config: config,
		llm:    llmInstance,
		tools:  make(map[string]types.Tool),
	}
}

func (a *Agent) GetLLMInstance() llm.LLM {
	return a.llm
}

// RegisterTool 注册工具
func (a *Agent) RegisterTool(tool types.Tool) {
	a.toolsMux.Lock()
	defer a.toolsMux.Unlock()
	a.tools[tool.Name()] = tool
}

// ListTools 列出所有可用工具
func (a *Agent) ListTools() []string {
	a.toolsMux.RLock()
	defer a.toolsMux.RUnlock()

	tools := make([]string, 0, len(a.tools))
	for name := range a.tools {
		tools = append(tools, name)
	}
	return tools
}

// GetToolDescriptions 获取所有工具的描述
func (a *Agent) GetToolDescriptions() string {
	a.toolsMux.RLock()
	defer a.toolsMux.RUnlock()

	var desc string
	desc = "可用工具列表：\n\n"
	for name, tool := range a.tools {
		desc += fmt.Sprintf("工具名: %s\n", name)
		desc += fmt.Sprintf("描述: %s\n\n", tool.Description())
	}
	return desc
}

// Execute 执行用户查询
func (a *Agent) Execute(query string) *types.ExecutionResult {
	startTime := time.Now()

	result := &types.ExecutionResult{
		Tasks:      make([]types.Task, 0),
		Success:    false,
		Iterations: 0,
	}

	if a.config.Debug {
		fmt.Printf("\n=== 开始执行任务 ===\n")
		fmt.Printf("查询: %s\n", query)
		fmt.Printf("最大迭代次数: %d\n\n", a.config.MaxIterations)
	}

	// 主循环
	for iteration := 0; iteration < a.config.MaxIterations; iteration++ {
		result.Iterations = iteration + 1

		if a.config.Debug {
			fmt.Printf("\n--- 迭代 %d/%d ---\n", iteration+1, a.config.MaxIterations)
		}

		// 1. 分析任务，决定下一步行动
		decision, err := a.analyzeAndDecide(query, result.Tasks)
		if err != nil {
			result.ErrorMessage = fmt.Sprintf("分析任务失败: %v", err)
			break
		}

		if a.config.Debug {
			fmt.Printf("决策: %s\n", decision.Reasoning)
			fmt.Printf("需要继续工作: %v\n", decision.NeedMoreWork)
		}

		// 2. 如果不需要更多工作，返回最终答案
		if !decision.NeedMoreWork {
			result.FinalAnswer = decision.FinalAnswer
			result.Success = true
			if a.config.Debug {
				fmt.Printf("\n最终答案: %s\n", result.FinalAnswer)
			}
			break
		}

		// 3. 执行任务
		if len(decision.Tasks) == 0 {
			result.ErrorMessage = "没有可执行的任务"
			break
		}

		tasksCompleted := a.executeTasks(decision.Tasks, &result.Tasks)

		if a.config.Debug {
			fmt.Printf("本轮完成任务数: %d\n", tasksCompleted)
		}

		// 4. 检查是否所有任务都失败
		allFailed := true
		for i := len(result.Tasks) - tasksCompleted; i < len(result.Tasks); i++ {
			if result.Tasks[i].Status == "completed" {
				allFailed = false
				break
			}
		}

		if allFailed && tasksCompleted > 0 {
			result.ErrorMessage = "所有任务执行失败"
			break
		}
	}

	// 如果达到最大迭代次数仍未完成
	if result.Iterations >= a.config.MaxIterations && !result.Success {
		result.ErrorMessage = fmt.Sprintf("达到最大迭代次数 (%d)", a.config.MaxIterations)
		result.FinalAnswer = a.aggregateResults(result.Tasks)
	}

	result.TotalTime = time.Since(startTime).Seconds()

	if a.config.Debug {
		fmt.Printf("\n=== 执行完成 ===\n")
		fmt.Printf("总耗时: %.2f秒\n", result.TotalTime)
		fmt.Printf("迭代次数: %d\n", result.Iterations)
		fmt.Printf("成功: %v\n", result.Success)
	}

	return result
}

// analyzeAndDecide 分析任务并做出决策
func (a *Agent) analyzeAndDecide(query string, taskHistory []types.Task) (*types.AgentDecision, error) {
	// 构建上下文
	context := query
	if len(taskHistory) > 0 {
		context += "\n\n已完成的任务:\n"
		for _, task := range taskHistory {
			context += fmt.Sprintf("- %s [%s]: %s\n",
				task.Tool, task.Status, task.Output)
		}
	}

	// 调用LLM进行分析
	decision, err := a.llm.AnalyzeTask(context, taskHistory)
	if err != nil {
		return nil, err
	}

	return decision, nil
}

// executeTasks 执行任务列表
func (a *Agent) executeTasks(toolCalls []types.ToolCall, taskList *[]types.Task) int {
	completed := 0

	for i, call := range toolCalls {
		task := types.Task{
			ID:          fmt.Sprintf("task_%d_%d", len(*taskList), i),
			Description: call.Reason,
			Status:      "pending",
			Tool:        call.ToolName,
			Input:       call.Input,
			CreatedAt:   time.Now(),
		}

		if a.config.Debug {
			fmt.Printf("\n执行任务: %s\n", task.ID)
			fmt.Printf("工具: %s\n", task.Tool)
			fmt.Printf("输入: %s\n", task.Input)
		}

		// 查找工具
		a.toolsMux.RLock()
		tool, exists := a.tools[call.ToolName]
		a.toolsMux.RUnlock()

		if !exists {
			task.Status = "failed"
			task.Error = fmt.Sprintf("工具 '%s' 不存在", call.ToolName)
			if a.config.Debug {
				fmt.Printf("错误: %s\n", task.Error)
			}
			*taskList = append(*taskList, task)
			continue
		}

		// 执行工具
		task.Status = "running"
		fmt.Printf("🔧调用工具：[%s], Param:[%s]...\n", tool.Name(), call.Input)
		output, err := tool.Execute(call.Input)
		fmt.Printf("🔧工具[%s]结果：[%s]\n", tool.Name(), output)

		if err != nil {
			task.Status = "failed"
			task.Error = err.Error()
			if a.config.Debug {
				fmt.Printf("执行失败: %s\n", task.Error)
			}
		} else {
			task.Status = "completed"
			task.Output = output
			completed++
			if a.config.Debug {
				fmt.Printf("执行成功\n输出: %s\n", truncate(output, 200))
			}
		}

		task.CompletedAt = time.Now()
		*taskList = append(*taskList, task)
	}

	return completed
}

// aggregateResults 聚合任务结果
func (a *Agent) aggregateResults(tasks []types.Task) string {
	if len(tasks) == 0 {
		return "没有执行任何任务。"
	}

	result := "任务执行总结:\n\n"

	completedCount := 0
	for _, task := range tasks {
		if task.Status == "completed" {
			completedCount++
			result += fmt.Sprintf("✓ [%s] %s\n结果: %s\n\n",
				task.Tool, task.Description, truncate(task.Output, 150))
		} else if task.Status == "failed" {
			result += fmt.Sprintf("✗ [%s] %s\n错误: %s\n\n",
				task.Tool, task.Description, task.Error)
		}
	}

	result += fmt.Sprintf("完成 %d/%d 个任务", completedCount, len(tasks))
	return result
}

// truncate 截断字符串
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
