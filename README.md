# AI Agent - 智能任务执行系统

一个用Go语言编写的功能完整的AI Agent系统，支持循环分析、任务分解、工具调用和结果聚合。

## 🌟 特性

- ✅ **循环分析**: 自动分析任务并决定是否需要继续执行
- ✅ **任务分解**: 将复杂任务分解为多个子任务
- ✅ **工具系统**: 支持多种工具的注册和调用
- ✅ **结果聚合**: 自动聚合多个任务的执行结果
- ✅ **最大迭代限制**: 防止无限循环
- ✅ **调试模式**: 详细的执行日志

## 🛠️ 内置工具

1. **Python工具** (`python`) - 执行Python代码
2. **Shell工具** (`shell`) - 执行Shell命令
3. **搜索工具** (`search`) - 网络搜索（DuckDuckGo API）
4. **计算器工具** (`calculator`) - 数学计算
5. **简单计算器** (`simple_math`) - 基础算术运算
6. **文件操作工具** (`file_ops`) - 文件读写、列表、删除

## 📦 项目结构

```
ai-agent/
├── cmd/
│   └── main.go              # 主程序入口
├── internal/
│   ├── agent/
│   │   └── agent.go         # Agent核心实现
│   ├── llm/
│   │   └── mock_llm.go      # LLM接口（模拟版）
│   └── tools/
│       ├── python.go        # Python工具
│       ├── shell.go         # Shell工具
│       ├── search.go        # 搜索工具
│       ├── calculator.go    # 计算器工具
│       └── file_ops.go      # 文件操作工具
└── pkg/
    └── types/
        └── types.go         # 类型定义
```

## 🚀 快速开始

### 编译

```bash
cd ai-agent
go build -o agent ./cmd
```

### 运行

**交互模式:**
```bash
./agent
```

**命令行模式:**
```bash
./agent "计算 25 * 4"
./agent "用Python计算1到100的和"
```

## 📖 使用示例

### 示例 1: 数学计算

```
>>> 计算 15 * 23
```

### 示例 2: Python代码执行

```
>>> 用Python计算斐波那契数列前10项
```

输入的Python代码会被自动识别并执行。

### 示例 3: 文件操作

```
>>> 在test.txt文件中写入"Hello World"
```

### 示例 4: 网络搜索

```
>>> 搜索Go语言最新版本
```

### 示例 5: 组合任务

```
>>> 用Python生成1-10的平方数，保存到文件squares.txt
```

## 🔧 配置选项

在 `cmd/main.go` 中可以配置：

```go
config := types.AgentConfig{
    MaxIterations: 5,                    // 最大迭代次数
    Model:         "mock-gpt",           // LLM模型
    Temperature:   0.7,                  // 温度参数
    Timeout:       30 * time.Second,     // 超时时间
    Debug:         true,                 // 调试模式
}
```

## 🏗️ 架构设计

### 核心流程

```
用户查询
    ↓
[循环开始] (最多N次)
    ↓
1. 任务分析
    - LLM分析当前状态
    - 决定是否需要更多工作
    - 生成任务列表
    ↓
2. 任务执行
    - 并发执行多个工具
    - 收集执行结果
    ↓
3. 结果聚合
    - 整合所有任务输出
    - 更新任务历史
    ↓
4. 决策判断
    - 检查是否完成
    - 决定是否继续循环
    ↓
[循环结束] 或 返回步骤1
    ↓
返回最终结果
```

### Agent决策逻辑

```go
type AgentDecision struct {
    NeedMoreWork bool       // 是否需要继续工作
    Tasks        []ToolCall // 下一步要执行的任务
    FinalAnswer  string     // 最终答案（如果完成）
    Reasoning    string     // 推理过程
}
```

## 🔌 添加自定义工具

实现 `types.Tool` 接口：

```go
type Tool interface {
    Name() string
    Description() string
    Execute(input string) (string, error)
}
```

示例：

```go
type WeatherTool struct{}

func (t *WeatherTool) Name() string {
    return "weather"
}

func (t *WeatherTool) Description() string {
    return "获取天气信息。输入：城市名称"
}

func (t *WeatherTool) Execute(input string) (string, error) {
    // 实现天气查询逻辑
    return fmt.Sprintf("%s的天气：晴天，25°C", input), nil
}

// 注册工具
agent.RegisterTool(&WeatherTool{})
```

## 🔄 集成真实LLM

替换 `internal/llm/mock_llm.go` 中的模拟实现：

### OpenAI示例

```go
import "github.com/sashabaranov/go-openai"

type OpenAILLM struct {
    client *openai.Client
}

func (l *OpenAILLM) AnalyzeTask(query string, history []types.Task) (*types.AgentDecision, error) {
    // 构建prompt
    prompt := buildPrompt(query, history)
    
    // 调用OpenAI API
    resp, err := l.client.CreateChatCompletion(
        context.Background(),
        openai.ChatCompletionRequest{
            Model: openai.GPT4,
            Messages: []openai.ChatCompletionMessage{
                {Role: "system", Content: systemPrompt},
                {Role: "user", Content: prompt},
            },
        },
    )
    
    // 解析响应为AgentDecision
    return parseDecision(resp.Choices[0].Message.Content)
}
```

### Claude示例

```go
import anthropic "github.com/anthropics/anthropic-sdk-go"

type ClaudeLLM struct {
    client *anthropic.Client
}

func (l *ClaudeLLM) AnalyzeTask(query string, history []types.Task) (*types.AgentDecision, error) {
    // 类似OpenAI的实现
}
```

## 🧪 测试

创建测试文件：

```go
func TestAgent(t *testing.T) {
    config := types.AgentConfig{
        MaxIterations: 3,
        Debug:         false,
    }
    
    llm := llm.NewMockLLM("test")
    agent := agent.NewAgent(config, llm)
    agent.RegisterTool(tools.NewCalculatorTool())
    
    result := agent.Execute("计算 10 + 20")
    
    assert.True(t, result.Success)
    assert.Contains(t, result.FinalAnswer, "30")
}
```

## 📝 工作原理

### 1. 任务分解
Agent接收用户查询后，使用LLM分析需要执行的操作，并分解为多个子任务。

### 2. 工具选择
根据任务类型，Agent自动选择合适的工具（计算器、Python、搜索等）。

### 3. 循环执行
- 第1轮：执行初始任务
- 第2轮：根据第1轮结果，决定是否需要更多操作
- ...
- 最多执行N轮（可配置）

### 4. 结果聚合
将所有任务的输出整合成最终答案返回给用户。

## 🛡️ 安全注意事项

- Python和Shell工具有安全风险，在生产环境中应该：
    - 使用沙箱环境（Docker容器）
    - 限制可执行的命令
    - 添加输入验证
    - 设置资源限制

## 🔮 未来改进

- [ ] 支持真实LLM（OpenAI、Claude、本地模型）
- [ ] 添加工具调用的并发执行
- [ ] 实现更智能的任务规划算法
- [ ] 添加记忆和上下文管理
- [ ] 支持函数调用（Function Calling）
- [ ] 添加Web界面
- [ ] 持久化任务历史
- [ ] 支持插件系统

## 📄 许可证

MIT License

## 🤝 贡献

欢迎提交Issue和Pull Request！

## 📧 联系方式

如有问题，请创建GitHub Issue。