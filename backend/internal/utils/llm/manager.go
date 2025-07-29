package llm

import (
	"context"
	"fmt"
	"sync"
)

// ClientManager 管理不同的LLM客户端
type ClientManager struct {
	clients map[LLMProvider]LLMClient
	configs map[LLMProvider]*Config
	mutex   sync.RWMutex
}

// NewClientManager 创建新的客户端管理器
func NewClientManager() *ClientManager {
	return &ClientManager{
		clients: make(map[LLMProvider]LLMClient),
		configs: make(map[LLMProvider]*Config),
	}
}

// RegisterProvider 注册LLM提供商
func (cm *ClientManager) RegisterProvider(provider LLMProvider, config *Config) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	if config == nil {
		return fmt.Errorf("config cannot be nil for provider %s", provider)
	}

	if config.APIKey == "" {
		return fmt.Errorf("API key is required for provider %s", provider)
	}

	// 创建客户端
	client, err := cm.createClient(provider, config)
	if err != nil {
		return fmt.Errorf("failed to create client for provider %s: %w", provider, err)
	}

	cm.clients[provider] = client
	cm.configs[provider] = config

	return nil
}

// GetClient 获取指定提供商的客户端
func (cm *ClientManager) GetClient(provider LLMProvider) (LLMClient, error) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	client, exists := cm.clients[provider]
	if !exists {
		return nil, fmt.Errorf("client for provider %s not found", provider)
	}

	return client, nil
}

// GetDefaultClient 获取默认客户端（优先OpenAI，然后Hunyuan）
func (cm *ClientManager) GetDefaultClient() (LLMClient, error) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	// 优先返回OpenAI客户端
	if client, exists := cm.clients[ProviderOpenAI]; exists {
		return client, nil
	}

	// 其次返回Hunyuan客户端
	if client, exists := cm.clients[ProviderHunyuan]; exists {
		return client, nil
	}

	return nil, fmt.Errorf("no available LLM clients")
}

// ListProviders 列出所有已注册的提供商
func (cm *ClientManager) ListProviders() []LLMProvider {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	providers := make([]LLMProvider, 0, len(cm.clients))
	for provider := range cm.clients {
		providers = append(providers, provider)
	}

	return providers
}

// Chat 使用指定提供商进行阻塞式聊天
func (cm *ClientManager) Chat(ctx context.Context, provider LLMProvider, req *ChatRequest) (*ChatResponse, error) {
	client, err := cm.GetClient(provider)
	if err != nil {
		return nil, err
	}

	return client.Chat(ctx, req)
}

// StreamChat 使用指定提供商进行流式聊天
func (cm *ClientManager) StreamChat(ctx context.Context, provider LLMProvider, req *ChatRequest) (<-chan StreamEvent, error) {
	client, err := cm.GetClient(provider)
	if err != nil {
		return nil, err
	}

	return client.StreamChat(ctx, req)
}

// ChatWithDefault 使用默认客户端进行阻塞式聊天
func (cm *ClientManager) ChatWithDefault(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	client, err := cm.GetDefaultClient()
	if err != nil {
		return nil, err
	}

	return client.Chat(ctx, req)
}

// StreamChatWithDefault 使用默认客户端进行流式聊天
func (cm *ClientManager) StreamChatWithDefault(ctx context.Context, req *ChatRequest) (<-chan StreamEvent, error) {
	client, err := cm.GetDefaultClient()
	if err != nil {
		return nil, err
	}

	return client.StreamChat(ctx, req)
}

// Close 关闭所有客户端
func (cm *ClientManager) Close() error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	var lastErr error
	for provider, client := range cm.clients {
		if err := client.Close(); err != nil {
			lastErr = fmt.Errorf("failed to close client for provider %s: %w", provider, err)
		}
	}

	// 清空clients和configs
	cm.clients = make(map[LLMProvider]LLMClient)
	cm.configs = make(map[LLMProvider]*Config)

	return lastErr
}

// createClient 创建指定提供商的客户端
func (cm *ClientManager) createClient(provider LLMProvider, config *Config) (LLMClient, error) {
	switch provider {
	case ProviderOpenAI:
		return NewOpenAIClient(config)
	case ProviderHunyuan:
		return NewHunyuanClient(config)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}

// 全局客户端管理器实例
var globalManager *ClientManager
var globalManagerOnce sync.Once

// GetGlobalManager 获取全局客户端管理器
func GetGlobalManager() *ClientManager {
	globalManagerOnce.Do(func() {
		globalManager = NewClientManager()
	})
	return globalManager
}
