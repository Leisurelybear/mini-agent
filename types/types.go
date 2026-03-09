package types

import "time"

// Message 表示对话消息
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Task 表示一个任务
type Task struct {
	ID          string    `json:"id"`
	Description string    `json:"description"`
	Status      string    `json:"status"` // pending, running, completed, failed
	Tool        string    `json:"tool,omitempty"`
	Input       string    `json:"input,omitempty"`
	Output      string    `json:"output,omitempty"`
	Error       string    `json:"error,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	CompletedAt time.Time `json:"completed_at,omitempty"`
}

// Tool 工具接口
type Tool interface {
	Name() string
	Description() string
	Execute(input string) (string, error)
}

// AgentConfig Agent配置
type AgentConfig struct {
	MaxIterations int           // 最大循环次数
	Model         string        // LLM模型名称
	Temperature   float64       // 温度参数
	Timeout       time.Duration // 每次工具调用超时时间
	Debug         bool          // 调试模式
}

// ExecutionResult 执行结果
type ExecutionResult struct {
	FinalAnswer  string  `json:"final_answer"`
	Tasks        []Task  `json:"tasks"`
	Iterations   int     `json:"iterations"`
	TotalTime    float64 `json:"total_time_seconds"`
	Success      bool    `json:"success"`
	ErrorMessage string  `json:"error_message,omitempty"`
}

// ToolCall 工具调用
type ToolCall struct {
	ToolName string `json:"tool_name"`
	Input    string `json:"input"`
	Reason   string `json:"reason"`
}

// AgentDecision Agent决策
type AgentDecision struct {
	NeedMoreWork bool       `json:"need_more_work"`
	Tasks        []ToolCall `json:"tasks"`
	FinalAnswer  string     `json:"final_answer,omitempty"`
	Reasoning    string     `json:"reasoning"`
}
