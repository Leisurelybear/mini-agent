package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"mini_agent/agent"
	"mini_agent/config"
	"mini_agent/llm"
	"mini_agent/tools"
)

func main() { // 帮我在桌面创建一个文件，然后把圆周率的100位写入，然后改一个合适的名字
	fmt.Println("===========================================")
	fmt.Println("    AI Agent - 智能任务执行系统")
	fmt.Println("===========================================")
	fmt.Println()

	// 命令行参数
	configFile := flag.String("config", "config.toml", "配置文件路径")
	useOpenAI := flag.Bool("openai", false, "使用OpenAI（覆盖配置文件）")
	model := flag.String("model", "", "LLM模型名称（覆盖配置文件）")
	maxIter := flag.Int("max-iter", 0, "最大迭代次数（覆盖配置文件）")
	debug := flag.Bool("debug", false, "调试模式")
	generateConfig := flag.Bool("generate-config", false, "生成默认配置文件")
	flag.Parse()

	// 生成配置文件
	if *generateConfig {
		defaultConfig := config.NewDefaultConfig()
		if err := config.SaveConfig(defaultConfig, "config.toml"); err != nil {
			fmt.Printf("❌ 生成配置文件失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✅ 已生成默认配置文件: config.toml")
		os.Exit(0)
	}

	// 加载配置
	cfg, err := config.LoadConfig(*configFile)
	if err != nil {
		fmt.Printf("⚠️  加载配置文件失败: %v\n", err)
		fmt.Println("   使用默认配置...")
		cfg = config.NewDefaultConfig()
	} else {
		fmt.Printf("✅ 已加载配置文件: %s\n", *configFile)
	}

	// 命令行参数覆盖配置文件
	if *useOpenAI {
		cfg.LLM.Provider = "openai"
	}
	if *model != "" {
		cfg.LLM.Model = *model
	}
	if *maxIter > 0 {
		cfg.Agent.MaxIterations = *maxIter
	}
	if *debug {
		cfg.Agent.Debug = true
	}

	// 创建LLM实例
	var llmInstance llm.LLM
	if cfg.LLM.Provider == "openai" {
		if cfg.LLM.APIKey == "" {
			fmt.Println("❌ 错误: 使用OpenAI时必须设置API密钥")
			fmt.Println("   方式1: 在config.toml中设置 llm.api_key")
			fmt.Println("   方式2: 设置环境变量 export OPENAI_API_KEY='your-key'")
			os.Exit(1)
		}
		fmt.Printf("🤖 使用OpenAI模型: %s\n", cfg.LLM.Model)
		if cfg.LLM.BaseURL != "" {
			fmt.Printf("   API地址: %s\n", cfg.LLM.BaseURL)
		}
		llmInstance = llm.NewOpenAILLM(cfg.LLM.APIKey, cfg.LLM.Model, cfg.LLM.BaseURL, float32(cfg.LLM.Temperature))
	} else {
		fmt.Println("🤖 使用模拟LLM（适合测试）")
		fmt.Println("   提示: 在config.toml中设置 llm.provider = \"openai\" 启用真实模型")
		llmInstance = llm.NewMockLLM(cfg.LLM.Model)
	}
	fmt.Println()

	// 显示配置信息
	if cfg.Agent.Debug {
		fmt.Printf("⚙️  配置信息:\n")
		fmt.Printf("   - 最大迭代次数: %d\n", cfg.Agent.MaxIterations)
		fmt.Printf("   - 超时时间: %ds\n", cfg.Agent.Timeout)
		fmt.Printf("   - 工作目录: %s\n", cfg.Agent.Workspace)
		fmt.Printf("   - 调试模式: %v\n", cfg.Agent.Debug)
		fmt.Println()
	}

	// 创建Agent
	agentConfig := cfg.ToAgentConfig()
	aiAgent := agent.NewAgent(agentConfig, llmInstance)

	// 注册工具
	registerTools(aiAgent)
	llmInstance.BindToolsDesc(aiAgent.GetToolDescriptions())

	// 显示可用工具
	fmt.Println("已注册工具:")
	for i, toolName := range aiAgent.ListTools() {
		fmt.Printf("  %d. %s\n", i+1, toolName)
	}
	fmt.Println()

	// 交互式模式
	if len(os.Args) > 1 {
		// 命令行模式
		query := strings.Join(os.Args[1:], " ")
		executeQuery(aiAgent, query)
	} else {
		// 交互模式
		interactiveMode(aiAgent)
	}
}

func registerTools(a *agent.Agent) {
	// 注册Python工具
	a.RegisterTool(tools.NewPythonTool(30 * time.Second))

	// 注册Shell工具
	a.RegisterTool(tools.NewShellTool(30 * time.Second))

	// 注册搜索工具
	a.RegisterTool(tools.NewSearchTool(10*time.Second, ""))

	// 注册计算器工具
	a.RegisterTool(tools.NewCalculatorTool())
	a.RegisterTool(tools.NewSimpleMathTool())

	// 注册文件操作工具
	a.RegisterTool(tools.NewFileOperationTool("/tmp/agent-workspace"))

	// 注册天气工具
	a.RegisterTool(tools.NewWeatherTool(10 * time.Second))

	// 注册股票工具
	a.RegisterTool(tools.NewStockTool(10 * time.Second))

	// 注册AI搜索工具
	a.RegisterTool(tools.NewAISearchTool(a.GetLLMInstance()))

	// 注册字符长度工具
	a.RegisterTool(tools.NewWordLengthTool())
}

func interactiveMode(a *agent.Agent) {
	fmt.Println("进入交互模式（输入 'quit' 或 'exit' 退出，'help' 查看帮助）")
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print(">>> ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		// 处理特殊命令
		switch strings.ToLower(input) {
		case "quit", "exit":
			fmt.Println("再见！")
			return
		case "help":
			showHelp(a)
			continue
		case "tools":
			showTools(a)
			continue
		case "clear":
			fmt.Print("\033[H\033[2J")
			continue
		}

		// 执行查询
		executeQuery(a, input)
		fmt.Println()
	}
}

func executeQuery(a *agent.Agent, query string) {
	fmt.Println("\n" + strings.Repeat("=", 60))

	result := a.Execute(query)

	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("\n📊 执行结果:")
	fmt.Println(strings.Repeat("-", 60))

	if result.Success {
		fmt.Println("✅ 状态: 成功")
		fmt.Printf("\n💡 最终答案:\n%s\n", result.FinalAnswer)
	} else {
		fmt.Println("❌ 状态: 失败")
		if result.ErrorMessage != "" {
			fmt.Printf("\n❗ 错误: %s\n", result.ErrorMessage)
		}
		if result.FinalAnswer != "" {
			fmt.Printf("\n📝 部分结果:\n%s\n", result.FinalAnswer)
		}
	}

	fmt.Println(strings.Repeat("-", 60))
	fmt.Printf("⏱️  总耗时: %.2f秒\n", result.TotalTime)
	fmt.Printf("🔄 迭代次数: %d/%d\n", result.Iterations, 5)
	fmt.Printf("📋 执行任务数: %d\n", len(result.Tasks))

	// 显示任务详情
	if len(result.Tasks) > 0 {
		fmt.Println("\n📝 任务详情:")
		for i, task := range result.Tasks {
			status := "✓"
			if task.Status == "failed" {
				status = "✗"
			}
			fmt.Printf("\n  %s 任务 %d: %s\n", status, i+1, task.Description)
			fmt.Printf("     工具: %s\n", task.Tool)
			if task.Status == "completed" {
				fmt.Printf("     结果: %s\n", truncate(task.Output, 100))
			} else if task.Status == "failed" {
				fmt.Printf("     错误: %s\n", task.Error)
			}
		}
	}

	fmt.Println(strings.Repeat("=", 60))
}

func showHelp(a *agent.Agent) {
	fmt.Println("\n📚 帮助信息:")
	fmt.Println(strings.Repeat("-", 60))
	fmt.Println("命令:")
	fmt.Println("  help   - 显示此帮助信息")
	fmt.Println("  tools  - 显示所有可用工具及其描述")
	fmt.Println("  clear  - 清屏")
	fmt.Println("  quit   - 退出程序")
	fmt.Println("\n使用示例:")
	fmt.Println("  >>> 计算 15 * 23")
	fmt.Println("  >>> 用Python计算1到100的和")
	fmt.Println("  >>> 搜索Go语言最新版本")
	fmt.Println("  >>> 列出当前目录的文件")
	fmt.Println(strings.Repeat("-", 60))
}

func showTools(a *agent.Agent) {
	fmt.Println("\n🛠️  可用工具:")
	fmt.Println(strings.Repeat("-", 60))
	fmt.Println(a.GetToolDescriptions())
	fmt.Println(strings.Repeat("-", 60))
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// 辅助函数：打印JSON
func printJSON(v interface{}) {
	data, _ := json.MarshalIndent(v, "", "  ")
	fmt.Println(string(data))
}
