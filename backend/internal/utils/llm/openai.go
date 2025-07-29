package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// OpenAIClient OpenAI客户端实现
type OpenAIClient struct {
	config     *Config
	httpClient *http.Client
	baseURL    string
}

// NewOpenAIClient 创建新的OpenAI客户端
func NewOpenAIClient(config *Config) (*OpenAIClient, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	if config.APIKey == "" {
		return nil, fmt.Errorf("OpenAI API key is required")
	}

	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}

	timeout := config.Timeout
	if timeout <= 0 {
		timeout = 60 // 默认60秒
	}

	return &OpenAIClient{
		config: config,
		httpClient: &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		},
		baseURL: baseURL,
	}, nil
}

// Chat 阻塞式聊天
func (c *OpenAIClient) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	// 确保不是流式请求
	req.Stream = false

	// 设置默认模型
	if req.Model == "" {
		req.Model = c.config.Model
		if req.Model == "" {
			req.Model = "gpt-3.5-turbo"
		}
	}

	// 设置默认参数
	if req.MaxTokens == 0 {
		req.MaxTokens = c.config.MaxTokens
		if req.MaxTokens == 0 {
			req.MaxTokens = 1000
		}
	}

	if req.Temperature == 0 {
		req.Temperature = c.config.Temperature
		if req.Temperature == 0 {
			req.Temperature = 0.7
		}
	}

	// 发送请求
	resp, err := c.sendRequest(ctx, "/chat/completions", req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.handleErrorResponse(resp)
	}

	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &chatResp, nil
}

// StreamChat 流式聊天
func (c *OpenAIClient) StreamChat(ctx context.Context, req *ChatRequest) (<-chan StreamEvent, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	// 确保是流式请求
	req.Stream = true

	// 设置默认模型
	if req.Model == "" {
		req.Model = c.config.Model
		if req.Model == "" {
			req.Model = "gpt-3.5-turbo"
		}
	}

	// 设置默认参数
	if req.MaxTokens == 0 {
		req.MaxTokens = c.config.MaxTokens
		if req.MaxTokens == 0 {
			req.MaxTokens = 1000
		}
	}

	if req.Temperature == 0 {
		req.Temperature = c.config.Temperature
		if req.Temperature == 0 {
			req.Temperature = 0.7
		}
	}

	// 发送请求
	resp, err := c.sendRequest(ctx, "/chat/completions", req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		chatResp, err := c.handleErrorResponse(resp)
		if err != nil {
			resp.Body.Close()
			return nil, err
		}
		resp.Body.Close()

		// 创建错误事件通道
		eventChan := make(chan StreamEvent, 1)
		eventChan <- StreamEvent{
			Error: fmt.Errorf("API error: %s", chatResp.Error.Message),
			Done:  true,
		}
		close(eventChan)
		return eventChan, nil
	}

	// 创建事件通道
	eventChan := make(chan StreamEvent, 10)

	// 启动goroutine处理流式响应
	go c.handleStreamResponse(ctx, resp.Body, eventChan)

	return eventChan, nil
}

// GetProvider 获取提供商类型
func (c *OpenAIClient) GetProvider() LLMProvider {
	return ProviderOpenAI
}

// Close 关闭客户端
func (c *OpenAIClient) Close() error {
	// HTTP客户端不需要显式关闭
	return nil
}

// sendRequest 发送HTTP请求
func (c *OpenAIClient) sendRequest(ctx context.Context, endpoint string, payload interface{}) (*http.Response, error) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.config.APIKey)

	return c.httpClient.Do(req)
}

// handleErrorResponse 处理错误响应
func (c *OpenAIClient) handleErrorResponse(resp *http.Response) (*ChatResponse, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read error response: %w", err)
	}

	var errorResp struct {
		Error ErrorResponse `json:"error"`
	}

	if err := json.Unmarshal(body, &errorResp); err != nil {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	return &ChatResponse{
		Error: &errorResp.Error,
	}, fmt.Errorf("OpenAI API error: %s", errorResp.Error.Message)
}

// handleStreamResponse 处理流式响应
func (c *OpenAIClient) handleStreamResponse(ctx context.Context, body io.ReadCloser, eventChan chan<- StreamEvent) {
	defer body.Close()
	defer close(eventChan)

	scanner := bufio.NewScanner(body)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			eventChan <- StreamEvent{
				Error: ctx.Err(),
				Done:  true,
			}
			return
		default:
		}

		line := scanner.Text()
		if line == "" {
			continue
		}

		// OpenAI的SSE格式：data: {...}
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")

			// 检查是否是结束标志
			if data == "[DONE]" {
				eventChan <- StreamEvent{Done: true}
				return
			}

			var chatResp ChatResponse
			if err := json.Unmarshal([]byte(data), &chatResp); err != nil {
				eventChan <- StreamEvent{
					Error: fmt.Errorf("failed to decode stream response: %w", err),
					Done:  true,
				}
				return
			}

			eventChan <- StreamEvent{
				Data: &chatResp,
				Done: false,
			}
		}
	}

	if err := scanner.Err(); err != nil {
		eventChan <- StreamEvent{
			Error: fmt.Errorf("stream reading error: %w", err),
			Done:  true,
		}
	}
}
