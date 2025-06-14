package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config 结构体定义了应用程序的配置项
type Config struct {
	APIKey       string   `mapstructure:"api_key"`       // LLM API Key
	Model        string   `mapstructure:"model"`         // LLM 模型名称
	BaseURL      string   `mapstructure:"base_url"`      // LLM API 端点
	AllowedTools []string `mapstructure:"allowed_tools"` // 允许使用的工具列表
	DeniedTools  []string `mapstructure:"denied_tools"`  // 禁止使用的工具列表
}

// LoadConfig 从配置文件和环境变量加载配置
// 优先级: 环境变量 > 配置文件 > 默认值
func LoadConfig() (*Config, error) {
	v := viper.New()

	// 设置默认配置
	v.SetDefault("model", "deepseek-coder")
	v.SetDefault("base_url", "https://api.deepseek.com/v1")
	v.SetDefault("allowed_tools", []string{"ps", "find", "grep", "wget", "ss", "lsof"})
	v.SetDefault("denied_tools", []string{})

	// 配置 Viper
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./configs")

	// 读取配置文件
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("读取配置文件失败: %w", err)
		}
	}

	// 配置环境变量
	v.SetEnvPrefix("AGENT")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 解析配置到结构体
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}

	// 验证必要配置
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("未设置 AGENT_API_KEY 环境变量或配置值")
	}

	return &cfg, nil
}
