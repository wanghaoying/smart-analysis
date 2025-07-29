# LLM 调用封装

这个包提供了对OpenAI和混元大模型的统一调用接口，支持流式和阻塞式调用，具有良好的可配置性和扩展性。

## 特性

- ✅ **统一接口**: 支持OpenAI和混元两种大模型提供商
- ✅ **双模式调用**: 支持阻塞式和流式调用
- ✅ **配置化**: 支持灵活的配置管理
- ✅ **设计模式**: 使用工厂模式和策略模式，易于扩展
- ✅ **错误处理**: 完善的错误处理和类型安全
- ✅ **并发安全**: 线程安全的客户端管理

## 架构设计

```
├── types.go      # 通用类型定义和接口
├── manager.go    # 客户端管理器（工厂模式）
├── openai.go     # OpenAI客户端实现
├── hunyuan.go    # 混元客户端实现
├── init.go       # 初始化和配置加载
└── example.go    # 使用示例
```

## 核心组件

### 1. LLMClient 接口
```go
type LLMClient interface {
    Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error)
    StreamChat(ctx context.Context, req *ChatRequest) (<-chan StreamEvent, error)
    GetProvider() LLMProvider
    Close() error
}
```

### 2. ClientManager 管理器
- 管理多个LLM客户端
- 提供统一的调用接口
- 支持默认客户端选择
- 线程安全

### 3. 配置系统
```go
type Config struct {
    Provider    LLMProvider `json:"provider"`
    APIKey      string      `json:"api_key"`
    BaseURL     string      `json:"base_url,omitempty"`
    Model       string      `json:"model"`
    MaxTokens   int         `json:"max_tokens"`
    Temperature float64     `json:"temperature"`
    Timeout     int         `json:"timeout"`
}
```

## 快速开始

### 1. 从应用配置初始化
```go
import (
    "smart-analysis/internal/config"
    "smart-analysis/internal/utils/llm"
)

// 加载应用配置
appConfig := config.Load()

// 初始化全局管理器
if err := llm.InitializeGlobalManager(appConfig); err != nil {
    log.Fatal(err)
}

// 获取管理器
manager := llm.GetGlobalManager()
```

### 2. 阻塞式调用
```go
ctx := context.Background()
req := &llm.ChatRequest{
    Messages: []llm.Message{
        {Role: "system", Content: "You are a helpful assistant."},
        {Role: "user", Content: "Hello!"},
    },
}

resp, err := manager.ChatWithDefault(ctx, req)
if err != nil {
    log.Fatal(err)
}

fmt.Println(resp.Choices[0].Message.Content)
```

### 3. 流式调用
```go
eventChan, err := manager.StreamChatWithDefault(ctx, req)
if err != nil {
    log.Fatal(err)
}

for event := range eventChan {
    if event.Error != nil {
        log.Printf("Error: %v", event.Error)
        break
    }
    
    if event.Data != nil && len(event.Data.Choices) > 0 {
        if delta := event.Data.Choices[0].Delta; delta != nil {
            fmt.Print(delta.Content)
        }
    }
    
    if event.Done {
        break
    }
}
```

### 4. 使用特定提供商
```go
// 使用OpenAI
resp, err := manager.Chat(ctx, llm.ProviderOpenAI, req)

// 使用混元
resp, err := manager.Chat(ctx, llm.ProviderHunyuan, req)
```

## 配置说明

### 环境变量配置
```bash
# OpenAI配置
OPENAI_API_KEY=your-openai-api-key

# 混元配置（格式：secretId:secretKey）
HUNYUAN_API_KEY=your-secret-id:your-secret-key
```

### 手动配置
```go
manager := llm.NewClientManager()

// OpenAI配置
openaiConfig := &llm.Config{
    Provider:    llm.ProviderOpenAI,
    APIKey:      "your-api-key",
    BaseURL:     "https://api.openai.com/v1",
    Model:       "gpt-3.5-turbo",
    MaxTokens:   1000,
    Temperature: 0.7,
    Timeout:     60,
}

manager.RegisterProvider(llm.ProviderOpenAI, openaiConfig)
```

## 支持的模型

### OpenAI
- gpt-3.5-turbo (默认)
- gpt-4
- gpt-4-turbo
- 其他OpenAI兼容模型

### 混元
- hunyuan-lite (默认)
- hunyuan-standard
- hunyuan-pro

## 错误处理

```go
resp, err := manager.Chat(ctx, llm.ProviderOpenAI, req)
if err != nil {
    // 网络或系统错误
    log.Printf("System error: %v", err)
    return
}

if resp.Error != nil {
    // API错误
    log.Printf("API error: %s (code: %s)", resp.Error.Message, resp.Error.Code)
    return
}

// 正常处理响应
fmt.Println(resp.Choices[0].Message.Content)
```

## 扩展新的提供商

1. 实现 `LLMClient` 接口
2. 在 `manager.go` 的 `createClient` 方法中添加新的case
3. 在 `types.go` 中添加新的提供商常量

```go
// 1. 定义新提供商
const ProviderNewLLM LLMProvider = "newllm"

// 2. 实现客户端
type NewLLMClient struct {
    // ... 实现细节
}

func (c *NewLLMClient) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
    // ... 实现逻辑
}

// 3. 在工厂方法中注册
func (cm *ClientManager) createClient(provider LLMProvider, config *Config) (LLMClient, error) {
    switch provider {
    case ProviderNewLLM:
        return NewNewLLMClient(config)
    // ...
    }
}
```

## 最佳实践

1. **资源管理**: 在应用关闭时调用 `manager.Close()`
2. **错误处理**: 区分系统错误和API错误
3. **超时控制**: 为context设置合适的超时时间
4. **流式处理**: 及时消费流式响应通道，避免阻塞
5. **配置管理**: 使用环境变量管理敏感信息

## 性能考虑

- 客户端实例可以复用，避免频繁创建
- 流式调用适合长文本生成场景
- 阻塞式调用适合简短对话场景
- 合理设置MaxTokens以控制成本

## 注意事项

1. **API密钥安全**: 不要在代码中硬编码API密钥
2. **混元签名**: 混元使用腾讯云签名算法，确保时间同步
3. **网络代理**: 某些地区可能需要配置代理访问OpenAI
4. **速率限制**: 注意各提供商的API调用频率限制
