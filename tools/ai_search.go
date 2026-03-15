package tools

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	// 这里替换为你实际的 llm 包路径
	"mini_agent/llm"
	"mini_agent/types"
)

// AISearchTool AI搜索工具（离线模型）
// 优化点：增加超时配置、输入校验、提示词优化、错误包装、返回格式化
type AISearchTool struct {
	agent   llm.LLM       // 底层LLM模型实例
	Timeout time.Duration // 搜索超时时间
	// 可选：添加提示词模板，方便定制不同场景的搜索策略
	promptTemplate string
}

// NewAISearchTool 创建AI搜索工具实例
// 参数：
//   - agent: LLM模型实例
//   - timeout: 可选，搜索超时时间（默认30秒）
func NewAISearchTool(agent llm.LLM, timeout ...time.Duration) *AISearchTool {
	// 默认超时30秒
	timeoutVal := 30 * time.Second
	if len(timeout) > 0 && timeout[0] > 0 {
		timeoutVal = timeout[0]
	}

	// 优化的提示词模板（引导模型输出结构化、准确的常识类答案）
	defaultPromptTemplate := `你是一个专业的离线知识问答助手，仅回答常识、定理、基础概念、通用方法等非实时内容：
1. 严格基于你的训练知识回答，不编造信息，不确定的内容明确说明"无法确定该常识信息"
2. 回答结构清晰，分点说明（如果适用），语言简洁易懂
3. 仅回答常识类问题，拒绝实时信息、最新数据、动态事件相关查询
4. 对于Windows、Linux、编程等实操类问题，给出具体步骤或示例

用户问题：%s`

	return &AISearchTool{
		agent:          agent,
		Timeout:        timeoutVal,
		promptTemplate: defaultPromptTemplate,
	}
}

// Name 工具名称（保持简洁且语义明确）
func (t *AISearchTool) Name() string {
	return "ai-search"
}

// Description 工具描述（详细、清晰，包含使用约束和示例）
func (t *AISearchTool) Description() string {
	return fmt.Sprintf(`AI搜索工具（离线模型）- 仅用于查询各领域常识、定理、基础概念、通用方法等非实时知识，不支持实时数据/动态事件查询！
使用约束：
1. 禁止查询实时内容（如：2026年最新Python版本、今日股市行情、实时新闻等）
2. 禁止查询主观评价、未公开信息、隐私相关内容
3. 超时时间：%s

输入格式：纯文本搜索查询字符串（无需特殊格式）

示例：
输入: Python 列表推导式用法
输出: 列表推导式是Python中创建列表的简洁方式，语法：[表达式 for 变量 in 可迭代对象 if 条件]
示例：
  squares = [x*x for x in range(10)]
  # 输出: [0,1,4,9,16,25,36,49,64,81]

输入: Windows系统如何创建文件夹
输出: Windows创建文件夹的常用方法：
1. 图形界面：右键空白处 → 新建 → 文件夹
2. CMD命令：md 文件夹名 或 mkdir 文件夹名
3. PowerShell：New-Item -Path 路径 -Name 文件夹名 -ItemType Directory

输入: 牛顿三大定律
输出: 牛顿三大定律是经典力学的核心：
1. 第一定律（惯性定律）：物体保持静止或匀速直线运动，除非受外力作用
2. 第二定律（F=ma）：加速度与合外力成正比，与质量成反比
3. 第三定律（作用力与反作用力）：两个物体间的作用力和反作用力大小相等、方向相反`, t.Timeout)
}

// Execute 执行AI搜索（核心优化：输入校验、超时控制、提示词优化、错误处理）
func (t *AISearchTool) Execute(input string) (string, error) {
	// 1. 输入校验
	input = strings.TrimSpace(input)
	if input == "" {
		return "", errors.New("搜索查询内容不能为空，请输入具体的常识类问题")
	}

	// 2. 敏感查询过滤（可选：根据需求扩展）
	sensitiveKeywords := []string{"实时", "2026", "最新", "今日", "现在", "股价", "新闻", "隐私", "密码"}
	for _, kw := range sensitiveKeywords {
		if strings.Contains(strings.ToLower(input), strings.ToLower(kw)) {
			return "", fmt.Errorf("查询包含禁止关键词「%s」，该工具仅支持常识类非实时内容查询", kw)
		}
	}

	// 3. 构建优化的提示词（避免模型跑偏）
	prompt := fmt.Sprintf(t.promptTemplate, input)

	// 4. 创建带超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), t.Timeout)
	defer cancel() // 确保超时后释放资源

	// 5. 调用LLM模型（适配上下文超时）
	// 注意：如果你的 llm.LLM.Chat 不支持context，需要改造底层Chat方法，或用goroutine+channel实现超时
	var resp string
	var err error
	done := make(chan struct{})
	go func() {
		// 调用底层LLM模型
		resp, err = t.agent.Chat([]types.Message{
			{
				Role:    "user", // 规范Role值（通常用小写user/assistant/system）
				Content: prompt,
			},
		})
		close(done)
	}()

	// 6. 超时控制
	select {
	case <-done:
		// 正常完成
	case <-ctx.Done():
		return "", fmt.Errorf("AI搜索超时（超时时间：%s），请简化查询或延长超时时间", t.Timeout)
	}

	// 7. 错误包装（增强可读性）
	if err != nil {
		return "", fmt.Errorf("AI搜索执行失败：%w", err)
	}

	// 8. 结果格式化（空结果处理）
	resp = strings.TrimSpace(resp)
	if resp == "" {
		resp = "未查询到相关常识信息，请调整查询关键词重试"
	}

	return resp, nil
}

// 可选：扩展方法 - 自定义提示词模板（适配不同场景）
func (t *AISearchTool) SetPromptTemplate(template string) {
	if template != "" {
		t.promptTemplate = template
	}
}
