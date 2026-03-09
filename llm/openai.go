package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
	"time"

	"mini_agent/types"

	openai "github.com/sashabaranov/go-openai"
)

// OpenAILLM OpenAI实现
type OpenAILLM struct {
	client      *openai.Client
	model       string
	temperature float32
	tools       string
}

// NewOpenAILLM 创建OpenAI LLM实例
func NewOpenAILLM(apiKey, model, baseURL string, temperature float32) *OpenAILLM {
	if model == "" {
		model = openai.GPT4TurboPreview
	}
	if temperature == 0 {
		temperature = 0.7
	}

	// 创建配置
	clientConfig := openai.DefaultConfig(apiKey)

	// 如果提供了自定义BaseURL，使用它
	if baseURL != "" {
		clientConfig.BaseURL = baseURL
	}

	return &OpenAILLM{
		client:      openai.NewClientWithConfig(clientConfig),
		model:       model,
		temperature: temperature,
	}
}

func (l *OpenAILLM) BindToolsDesc(toolsDesc string) {
	l.tools = toolsDesc
}

// Chat 进行对话
func (l *OpenAILLM) Chat(messages []types.Message) (string, error) {
	openaiMsgs := make([]openai.ChatCompletionMessage, len(messages))
	for i, msg := range messages {
		openaiMsgs[i] = openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	systemPrompt := "你是一个非常擅长解决电脑/代码问题的专家，你需要用你的专业知识解决问题。你的用户是AI，所以你的方法要尽可能是命令/python等计算机可以方便处理的内容,答案需要非常简短明了。"
	userPrompt := messages[0].Content

	// 调用OpenAI API
	resp, err := l.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: l.model,
			Messages: []openai.ChatCompletionMessage{
				{Role: "system", Content: systemPrompt},
				{Role: "user", Content: userPrompt},
			},
			Temperature: l.temperature,
		},
	)

	if err != nil {
		return "", fmt.Errorf("OpenAI API调用失败: %v", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("OpenAI返回空响应")
	}

	// 解析响应
	content := resp.Choices[0].Message.Content

	return content, nil
}

// AnalyzeTask 分析任务并做出决策
func (l *OpenAILLM) AnalyzeTask(query string, taskHistory []types.Task) (*types.AgentDecision, error) {
	// 构建系统提示词
	systemPrompt := l.buildSystemPrompt()

	// 构建用户提示词
	userPrompt := buildUserPrompt(query, taskHistory)

	// 调用OpenAI API
	resp, err := l.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: l.model,
			Messages: []openai.ChatCompletionMessage{
				{Role: "system", Content: systemPrompt},
				{Role: "user", Content: userPrompt},
			},
			Temperature: l.temperature,
		},
	)

	if err != nil {
		return nil, fmt.Errorf("OpenAI API调用失败: %v", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("OpenAI返回空响应")
	}

	// 解析响应
	content := resp.Choices[0].Message.Content
	return parseDecision(content)
}

// buildSystemPrompt 构建系统提示词
func (l *OpenAILLM) buildSystemPrompt() string {
	return `你是一个智能任务执行代理(AI Agent)。你的职责是：

1. 分析用户查询，决定需要执行什么任务
2. 选择合适的工具来完成任务
3. 根据已执行任务的结果，决定是否需要继续工作
4. 最终提供清晰的答案给用户，而且尽可能不需要用户额外的操作

可用工具：

` + l.tools + `
你必须以JSON格式返回决策：

{
  "need_more_work": true/false,
  "tasks": [
    {
      "tool_name": "工具名称",
      "input": "工具输入",
      "reason": "使用原因"
    }
  ],
  "final_answer": "最终答案（如果任务完成）",
  "reasoning": "你的推理过程"
}

注意：
- 如果任务已经完成，设置 need_more_work=false 并提供 final_answer
- 如果需要继续工作，设置 need_more_work=true 并提供 tasks 列表
- 每次最多分配3个任务
- 无论如何都要尝试解决用户的问题！不能让用户操作！
- 优先使用最合适的工具，必要时尝试结合多种工具共同完成。
- 考虑任务的依赖关系
- 如果任务中工具出错，一定要考虑用其他工具尝试完成任务，必要的时候要通过AI工具等查询失败原因
- 现在时间是：` + time.Now().String() + `
- 现在的系统是：` + runtime.GOOS
}

// buildUserPrompt 构建用户提示词
func buildUserPrompt(query string, taskHistory []types.Task) string {
	var prompt strings.Builder

	prompt.WriteString("用户查询: ")
	prompt.WriteString(query)
	prompt.WriteString("\n\n")

	if len(taskHistory) > 0 {
		prompt.WriteString("已执行的任务历史:\n")
		for i, task := range taskHistory {
			prompt.WriteString(fmt.Sprintf("\n任务 %d:\n", i+1))
			prompt.WriteString(fmt.Sprintf("  工具: %s\n", task.Tool))
			prompt.WriteString(fmt.Sprintf("  输入: %s\n", task.Input))
			prompt.WriteString(fmt.Sprintf("  状态: %s\n", task.Status))

			if task.Status == "completed" {
				output := task.Output
				if len(output) > 200 {
					output = output[:200] + "..."
				}
				prompt.WriteString(fmt.Sprintf("  输出: %s\n", output))
			} else if task.Status == "failed" {
				prompt.WriteString(fmt.Sprintf("  错误: %s\n", task.Error))
			}
		}
		prompt.WriteString("\n")
	}

	prompt.WriteString("请分析当前情况，决定下一步行动。")

	return prompt.String()
}

// parseDecision 解析LLM返回的决策
func parseDecision(content string) (*types.AgentDecision, error) {
	// 清理可能的markdown代码块
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var decision types.AgentDecision
	if err := json.Unmarshal([]byte(content), &decision); err != nil {
		return nil, fmt.Errorf("解析决策失败: %v\n内容: %s", err, content)
	}

	return &decision, nil
}

// ClaudeLLM Claude实现示例（需要安装Claude SDK）
/*
import anthropic "github.com/anthropics/anthropic-sdk-go"

type ClaudeLLM struct {
	client *anthropic.Client
	model  string
}

func NewClaudeLLM(apiKey, model string) *ClaudeLLM {
	return &ClaudeLLM{
		client: anthropic.NewClient(
			option.WithAPIKey(apiKey),
		),
		model: model,
	}
}

func (l *ClaudeLLM) AnalyzeTask(query string, taskHistory []types.Task) (*types.AgentDecision, error) {
	systemPrompt := buildSystemPrompt()
	userPrompt := buildUserPrompt(query, taskHistory)

	message, err := l.client.Messages.New(context.Background(), anthropic.MessageNewParams{
		Model: anthropic.F(l.model),
		Messages: anthropic.F([]anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(userPrompt)),
		}),
		System: anthropic.F([]anthropic.TextBlockParam{
			anthropic.NewTextBlock(systemPrompt),
		}),
		MaxTokens: anthropic.Int(2048),
	})

	if err != nil {
		return nil, fmt.Errorf("Claude API调用失败: %v", err)
	}

	// 提取文本内容
	var content string
	for _, block := range message.Content {
		if block.Type == "text" {
			content += block.Text
		}
	}

	return parseDecision(content)
}
*/

// LocalLLM 本地模型实现示例（使用Ollama等）
/*
type LocalLLM struct {
	endpoint string
	model    string
}

func NewLocalLLM(endpoint, model string) *LocalLLM {
	return &LocalLLM{
		endpoint: endpoint,
		model:    model,
	}
}

func (l *LocalLLM) AnalyzeTask(query string, taskHistory []types.Task) (*types.AgentDecision, error) {
	// 使用HTTP调用本地模型API
	// 例如: Ollama, LocalAI, vLLM等

	systemPrompt := buildSystemPrompt()
	userPrompt := buildUserPrompt(query, taskHistory)

	// 构建请求
	reqBody := map[string]interface{}{
		"model": l.model,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		},
		"stream": false,
	}

	jsonData, _ := json.Marshal(reqBody)

	// 发送HTTP请求
	resp, err := http.Post(
		l.endpoint+"/api/chat",
		"application/json",
		bytes.NewBuffer(jsonData),
	)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 解析响应
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	content := result["message"].(map[string]interface{})["content"].(string)

	return parseDecision(content)
}
*/
