package config

import (
	"fmt"
	"os"
	"time"

	"mini_agent/types"

	"github.com/BurntSushi/toml"
)

// Config 应用配置
type Config struct {
	Agent   AgentConfig   `toml:"agent"`
	LLM     LLMConfig     `toml:"llm"`
	Tools   ToolsConfig   `toml:"tools"`
	Logging LoggingConfig `toml:"logging"`
}

// AgentConfig Agent配置
type AgentConfig struct {
	MaxIterations int    `toml:"max_iterations"`
	Timeout       int    `toml:"timeout"` // 秒
	Debug         bool   `toml:"debug"`
	Workspace     string `toml:"workspace"`
}

// LLMConfig LLM配置
type LLMConfig struct {
	Provider    string  `toml:"provider"`
	Model       string  `toml:"model"`
	Temperature float64 `toml:"temperature"`
	APIKey      string  `toml:"api_key"`
	BaseURL     string  `toml:"base_url"`
}

// ToolsConfig 工具配置
type ToolsConfig struct {
	Python     ToolConfig    `toml:"python"`
	Shell      ShellConfig   `toml:"shell"`
	Search     SearchConfig  `toml:"search"`
	Calculator ToolConfig    `toml:"calculator"`
	SimpleMath ToolConfig    `toml:"simple_math"`
	FileOps    FileOpsConfig `toml:"file_ops"`
}

// ToolConfig 基础工具配置
type ToolConfig struct {
	Enabled bool `toml:"enabled"`
	Timeout int  `toml:"timeout"` // 秒
}

// ShellConfig Shell工具配置
type ShellConfig struct {
	ToolConfig
	AllowedCommands []string `toml:"allowed_commands"`
	BlockedCommands []string `toml:"blocked_commands"`
}

// SearchConfig 搜索工具配置
type SearchConfig struct {
	ToolConfig
	APIKey string `toml:"api_key"`
}

// FileOpsConfig 文件操作工具配置
type FileOpsConfig struct {
	ToolConfig
	Workspace    string   `toml:"workspace"`
	AllowedPaths []string `toml:"allowed_paths"`
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	Level    string `toml:"level"`
	Format   string `toml:"format"`
	Output   string `toml:"output"`
	FilePath string `toml:"file_path"`
}

// LoadConfig 从文件加载配置
func LoadConfig(path string) (*Config, error) {
	var config Config

	// 读取配置文件
	if _, err := toml.DecodeFile(path, &config); err != nil {
		return nil, fmt.Errorf("加载配置文件失败: %v", err)
	}

	// 从环境变量覆盖敏感信息
	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
		config.LLM.APIKey = apiKey
	}

	if baseURL := os.Getenv("OPENAI_BASE_URL"); baseURL != "" {
		config.LLM.BaseURL = baseURL
	}

	// 验证配置
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("配置验证失败: %v", err)
	}

	// 设置默认值
	config.SetDefaults()

	return &config, nil
}

// Validate 验证配置
func (c *Config) Validate() error {
	if c.Agent.MaxIterations <= 0 {
		return fmt.Errorf("max_iterations 必须大于0")
	}

	if c.Agent.MaxIterations > 20 {
		return fmt.Errorf("max_iterations 不能超过20（防止无限循环）")
	}

	if c.LLM.Provider == "openai" && c.LLM.APIKey == "" {
		return fmt.Errorf("使用OpenAI时必须设置API密钥（配置文件或环境变量OPENAI_API_KEY）")
	}

	if c.LLM.Temperature < 0 || c.LLM.Temperature > 2 {
		return fmt.Errorf("temperature 必须在0-2之间")
	}

	return nil
}

// SetDefaults 设置默认值
func (c *Config) SetDefaults() {
	if c.Agent.Timeout == 0 {
		c.Agent.Timeout = 30
	}

	if c.Agent.Workspace == "" {
		c.Agent.Workspace = "/tmp/agent-workspace"
	}

	if c.LLM.Model == "" {
		if c.LLM.Provider == "openai" {
			c.LLM.Model = "gpt-4-turbo-preview"
		} else {
			c.LLM.Model = "mock"
		}
	}

	if c.LLM.Temperature == 0 {
		c.LLM.Temperature = 0.7
	}

	if c.Logging.Level == "" {
		c.Logging.Level = "info"
	}

	if c.Logging.Format == "" {
		c.Logging.Format = "text"
	}

	if c.Logging.Output == "" {
		c.Logging.Output = "stdout"
	}

	// 工具默认超时
	if c.Tools.Python.Timeout == 0 {
		c.Tools.Python.Timeout = 30
	}
	if c.Tools.Shell.Timeout == 0 {
		c.Tools.Shell.Timeout = 30
	}
	if c.Tools.Search.Timeout == 0 {
		c.Tools.Search.Timeout = 10
	}
}

// ToAgentConfig 转换为Agent配置
func (c *Config) ToAgentConfig() types.AgentConfig {
	return types.AgentConfig{
		MaxIterations: c.Agent.MaxIterations,
		Model:         c.LLM.Model,
		Temperature:   c.LLM.Temperature,
		Timeout:       time.Duration(c.Agent.Timeout) * time.Second,
		Debug:         c.Agent.Debug,
	}
}

// GetToolTimeout 获取工具超时时间
func (c *Config) GetToolTimeout(toolName string) time.Duration {
	switch toolName {
	case "python":
		return time.Duration(c.Tools.Python.Timeout) * time.Second
	case "shell":
		return time.Duration(c.Tools.Shell.Timeout) * time.Second
	case "search":
		return time.Duration(c.Tools.Search.Timeout) * time.Second
	default:
		return 30 * time.Second
	}
}

// IsToolEnabled 检查工具是否启用
func (c *Config) IsToolEnabled(toolName string) bool {
	switch toolName {
	case "python":
		return c.Tools.Python.Enabled
	case "shell":
		return c.Tools.Shell.Enabled
	case "search":
		return c.Tools.Search.Enabled
	case "calculator":
		return c.Tools.Calculator.Enabled
	case "simple_math":
		return c.Tools.SimpleMath.Enabled
	case "file_ops":
		return c.Tools.FileOps.Enabled
	default:
		return true
	}
}

// SaveConfig 保存配置到文件
func SaveConfig(config *Config, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("创建配置文件失败: %v", err)
	}
	defer file.Close()

	encoder := toml.NewEncoder(file)
	if err := encoder.Encode(config); err != nil {
		return fmt.Errorf("编码配置失败: %v", err)
	}

	return nil
}

// NewDefaultConfig 创建默认配置
func NewDefaultConfig() *Config {
	config := &Config{
		Agent: AgentConfig{
			MaxIterations: 5,
			Timeout:       30,
			Debug:         true,
			Workspace:     "/tmp/agent-workspace",
		},
		LLM: LLMConfig{
			Provider:    "mock",
			Model:       "mock",
			Temperature: 0.7,
			APIKey:      "",
		},
		Tools: ToolsConfig{
			Python: ToolConfig{Enabled: true, Timeout: 30},
			Shell: ShellConfig{
				ToolConfig:      ToolConfig{Enabled: true, Timeout: 30},
				BlockedCommands: []string{"rm -rf /", "mkfs", "dd"},
			},
			Search:     SearchConfig{ToolConfig: ToolConfig{Enabled: true, Timeout: 10}},
			Calculator: ToolConfig{Enabled: true},
			SimpleMath: ToolConfig{Enabled: true},
			FileOps: FileOpsConfig{
				ToolConfig: ToolConfig{Enabled: true},
				Workspace:  "/tmp/agent-workspace",
			},
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "text",
			Output: "stdout",
		},
	}

	return config
}
