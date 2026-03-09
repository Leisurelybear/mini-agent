package tools

import (
	"mini_agent/llm"
	"mini_agent/types"
)

// AISearchTool 网络搜索工具
type AISearchTool struct {
	agent llm.LLM
}

func NewAISearchTool(agent llm.LLM) *AISearchTool {
	return &AISearchTool{
		agent: agent,
	}
}

func (t *AISearchTool) Name() string {
	return "ai-search"
}

func (t *AISearchTool) Description() string {
	return `AI搜索工具。只能用来查找各个领域常识定理等知识，不能用于查询实时内容！。
输入格式：搜索查询字符串
示例：
Python 用法
怎么处理windows文件`
}

func (t *AISearchTool) Execute(input string) (string, error) {
	resp, err := t.agent.Chat([]types.Message{
		{
			Role:    "User",
			Content: "Question:" + input,
		},
	})

	return resp, err
}
