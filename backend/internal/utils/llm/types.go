package llm

import (
	"context"
	"io"
)

// LLMProvider 定义LLM提供商类型
type LLMProvider string

const (
	ProviderOpenAI  LLMProvider = "openai"
	ProviderHunyuan LLMProvider = "hunyuan"
)

// Message 表示聊天消息
type Message struct {
	Role    string `json:"role"` // system, user, assistant
	Content string `json:"content"`
}

// ChatRequest 表示聊天请求参数
type ChatRequest struct {
	Messages    []Message `json:"messages"`
	Model       string    `json:"model"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
	Stream      bool      `json:"stream,omitempty"`
}

// ChatResponse 表示聊天响应
type ChatResponse struct {
	ID      string         `json:"id"`
	Object  string         `json:"object"`
	Created int64          `json:"created"`
	Model   string         `json:"model"`
	Choices []Choice       `json:"choices"`
	Usage   *TokenUsage    `json:"usage,omitempty"`
	Error   *ErrorResponse `json:"error,omitempty"`
}

// Choice 表示响应选择
type Choice struct {
	Index        int      `json:"index"`
	Message      *Message `json:"message,omitempty"`
	Delta        *Message `json:"delta,omitempty"`
	FinishReason string   `json:"finish_reason,omitempty"`
}

// TokenUsage 表示token使用情况
type TokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ErrorResponse 表示错误响应
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Type    string `json:"type"`
}

// StreamEvent 表示流式响应事件
type StreamEvent struct {
	Data  *ChatResponse `json:"data,omitempty"`
	Error error         `json:"error,omitempty"`
	Done  bool          `json:"done"`
}

// LLMClient 定义LLM客户端接口
type LLMClient interface {
	// Chat 阻塞式聊天
	Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error)

	// StreamChat 流式聊天
	StreamChat(ctx context.Context, req *ChatRequest) (<-chan StreamEvent, error)

	// GetProvider 获取提供商类型
	GetProvider() LLMProvider

	// Close 关闭客户端
	Close() error
}

// Config 表示LLM配置
type Config struct {
	Provider    LLMProvider `json:"provider"`
	APIKey      string      `json:"api_key"`
	BaseURL     string      `json:"base_url,omitempty"`
	Model       string      `json:"model"`
	MaxTokens   int         `json:"max_tokens"`
	Temperature float64     `json:"temperature"`
	Timeout     int         `json:"timeout"` // seconds
}

// StreamReader 流式读取器接口
type StreamReader interface {
	io.ReadCloser
	ReadEvent() (*StreamEvent, error)
}
