package llm

import (
	"fmt"
	"smart-analysis/internal/config"
	"strings"
)

// InitConfig 从应用配置初始化LLM配置
type InitConfig struct {
	OpenAI  *Config `json:"openai,omitempty"`
	Hunyuan *Config `json:"hunyuan,omitempty"`
}

// LoadConfigFromAppConfig 从应用配置加载LLM配置
func LoadConfigFromAppConfig(appConfig *config.Config) *InitConfig {
	initConfig := &InitConfig{}

	// 配置OpenAI
	if appConfig.OpenAIKey != "" {
		initConfig.OpenAI = &Config{
			Provider:    ProviderOpenAI,
			APIKey:      appConfig.OpenAIKey,
			BaseURL:     "", // 使用默认URL
			Model:       "gpt-3.5-turbo",
			MaxTokens:   1000,
			Temperature: 0.7,
			Timeout:     60,
		}
	}

	// 配置混元
	if appConfig.HunyuanKey != "" {
		initConfig.Hunyuan = &Config{
			Provider:    ProviderHunyuan,
			APIKey:      appConfig.HunyuanKey,
			BaseURL:     "", // 使用默认URL
			Model:       "hunyuan-lite",
			MaxTokens:   1000,
			Temperature: 0.7,
			Timeout:     60,
		}
	}

	return initConfig
}

// InitializeGlobalManager 初始化全局管理器
func InitializeGlobalManager(appConfig *config.Config) error {
	manager := GetGlobalManager()
	initConfig := LoadConfigFromAppConfig(appConfig)

	var registered int

	// 注册OpenAI客户端
	if initConfig.OpenAI != nil {
		if err := manager.RegisterProvider(ProviderOpenAI, initConfig.OpenAI); err != nil {
			return fmt.Errorf("failed to register OpenAI provider: %w", err)
		}
		registered++
	}

	// 注册混元客户端
	if initConfig.Hunyuan != nil {
		if err := manager.RegisterProvider(ProviderHunyuan, initConfig.Hunyuan); err != nil {
			return fmt.Errorf("failed to register Hunyuan provider: %w", err)
		}
		registered++
	}

	if registered == 0 {
		return fmt.Errorf("no LLM providers configured")
	}

	return nil
}

// DefaultConfig 获取默认配置
func DefaultConfig(provider LLMProvider, apiKey string) *Config {
	baseConfig := &Config{
		Provider:    provider,
		APIKey:      apiKey,
		MaxTokens:   1000,
		Temperature: 0.7,
		Timeout:     60,
	}

	switch provider {
	case ProviderOpenAI:
		baseConfig.BaseURL = "https://api.openai.com/v1"
		baseConfig.Model = "gpt-3.5-turbo"
	case ProviderHunyuan:
		baseConfig.BaseURL = "https://hunyuan.tencentcloudapi.com"
		baseConfig.Model = "hunyuan-lite"
	}

	return baseConfig
}

// ValidateConfig 验证配置
func ValidateConfig(config *Config) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	if config.APIKey == "" {
		return fmt.Errorf("API key is required")
	}

	if config.Provider != ProviderOpenAI && config.Provider != ProviderHunyuan {
		return fmt.Errorf("unsupported provider: %s", config.Provider)
	}

	if config.MaxTokens < 0 {
		return fmt.Errorf("max tokens must be non-negative")
	}

	if config.Temperature < 0 || config.Temperature > 2 {
		return fmt.Errorf("temperature must be between 0 and 2")
	}

	if config.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}

	// 验证混元API Key格式
	if config.Provider == ProviderHunyuan {
		parts := len(strings.Split(config.APIKey, ":"))
		if parts != 2 {
			return fmt.Errorf("Hunyuan API key must be in format 'secretId:secretKey'")
		}
	}

	return nil
}

// GetAvailableProviders 获取可用的提供商列表
func GetAvailableProviders() []LLMProvider {
	return []LLMProvider{ProviderOpenAI, ProviderHunyuan}
}

// IsProviderSupported 检查提供商是否支持
func IsProviderSupported(provider LLMProvider) bool {
	for _, p := range GetAvailableProviders() {
		if p == provider {
			return true
		}
	}
	return false
}
