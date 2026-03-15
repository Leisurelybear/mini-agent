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

// buildSystemPrompt 构建系统提示词（优化版）
func (l *OpenAILLM) buildSystemPrompt() string {
	// 获取格式化的时间和系统信息
	currentTime := time.Now().Format("2006-01-02 15:04:05")
	currentOS := runtime.GOOS
	osTips := ""
	switch currentOS {
	case "windows":
		osTips = "优先使用cmd命令而非bash，文件路径使用\\分隔，注意Windows特有的工具（如dir、copy）"
	case "linux", "darwin":
		osTips = "优先使用bash命令，文件路径使用/分隔，注意类Unix系统特有的工具（如ls、cp）"
	}

	return `你是一个专业的智能任务执行代理(AI Agent)，具备深度的任务拆解、工具选择和错误恢复能力。
你的核心工作流程：
==========
1. 任务分析阶段
   - 完全理解用户查询的核心目标（区分"需要执行的操作"和"最终要得到的结果"）
   - 识别任务类型：信息查询/数据计算/文件操作/命令执行/代码运行/HTTP请求/常识问答等
   - 评估任务复杂度：判断是否需要多工具协作、是否有任务依赖（如先读取文件再处理数据）

2. 工具选择阶段
   - 严格基于工具描述选择**最匹配、最高效**的工具，拒绝使用未列出的工具
   - 优先级规则：
     ✅ 优先使用专用工具（如计算字符长度用word_length，而非python_runner）
     ✅ 复杂计算/逻辑优先用python_runner，简单命令用command_executor
     ✅ 常识查询用ai-search，实时/网络查询用http_request
   - 每个任务必须有明确的工具输入格式（严格匹配工具描述中的输入要求）

3. 执行决策阶段
   - 单次最多生成3个任务，且任务间需考虑执行顺序（依赖前置任务完成的要标注）
   - 若当前信息足够回答用户问题，直接返回final_answer，设置need_more_work=false
   - 若需要执行工具才能完成，设置need_more_work=true，tasks列表包含具体工具调用
   - 若工具执行失败，优先尝试：①更换输入格式 ②选择替代工具 ③用ai-search查询失败原因

4. 结果输出阶段
   - 必须严格按照指定JSON格式返回，确保无语法错误（可解析）
   - final_answer需清晰、完整，包含用户需要的所有信息，避免模糊表述
   - reasoning字段需详细记录你的思考过程（工具选择理由、任务拆解逻辑、依赖关系等）
==========

可用工具列表：
` + l.tools + `

## 输出格式要求（必须严格遵守）
你只能返回JSON格式字符串，不允许添加任何额外文本、注释或格式说明：
{
  "need_more_work": true/false,  // 布尔值，仅true/false，无其他值
  "tasks": [                     // 数组，需要执行任务时非空，否则为空数组
    {
      "tool_name": "工具名称",    // 必须是可用工具列表中的准确名称（大小写敏感）
      "input": "工具输入",        // 严格匹配对应工具的输入格式要求
      "reason": "使用原因"        // 说明选择该工具的具体理由，包含输入格式合规性说明
    }
  ],
  "final_answer": "最终答案",     // 任务完成时填写，未完成时为空字符串
  "reasoning": "推理过程"         // 详细的思考过程，至少50字，说明任务分析、工具选择、依赖关系等
}

## 关键约束
1. 时间上下文：当前时间是 ` + currentTime + `，所有时间相关判断基于此
2. 系统上下文：当前系统是 ` + currentOS + `，` + osTips + `
3. 计算约束：所有数学计算、字符统计等必须使用工具执行，禁止自行计算
4. 错误处理：若工具执行失败，必须在tasks中生成替代工具调用，说明失败应对策略
5. 任务数量：单次tasks列表最多包含3个任务，且需按执行优先级排序
6. 格式约束：JSON必须可被标准解析器解析，禁止缺失字段、语法错误、多余逗号
7. 最终答案：仅当任务完全完成时填写final_answer，未完成时保持为空字符串

## 负面示例（禁止出现）
- 错误1：JSON格式错误（多余逗号、缺失引号）
- 错误2：tool_name拼写错误（如"command_executer"而非"command_executor"）
- 错误3：input格式不符合工具要求（如文件操作未按"操作|路径|内容"格式）
- 错误4：need_more_work为"True"（应为小写true）
- 错误5：reasoning字段为空或过于简略（如仅写"需要执行命令"）

## 正面示例（参考）
用户查询："计算'Hello World'的字符长度，并告诉我结果"
返回JSON：
{
  "need_more_work": true,
  "tasks": [
    {
      "tool_name": "word_length",
      "input": "Hello World",
      "reason": "用户需要计算字符串长度，word_length工具专门用于查询字符长度，输入格式为纯字符串，符合该工具的使用要求"
    }
  ],
  "final_answer": "",
  "reasoning": "用户的核心目标是获取'Hello World'的字符长度，首先分析任务类型为数据计算，可用工具中word_length是专用的字符长度计算工具，比python_runner更高效，因此选择该工具。执行该工具后即可得到结果，后续只需返回最终答案即可完成任务。"
}
`
}

// buildSystemPrompt 构建系统提示词, 旧的
func (l *OpenAILLM) buildSystemPrompt_old() string {
	return `你是一个智能任务执行代理(AI Agent)。你的职责是：

1. 分析用户查询，决定需要执行什么任务
2. 优先选择合适的工具来完成任务
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
- 无论如何都要尽最大努力解决用户的问题！
- 所有计算任务要用工具执行！
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
