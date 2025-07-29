package llm

import (
	"bufio"
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

// HunyuanClient 混元客户端实现
type HunyuanClient struct {
	config     *Config
	httpClient *http.Client
	baseURL    string
	secretId   string
	secretKey  string
}

// HunyuanMessage 混元消息格式
type HunyuanMessage struct {
	Role    string `json:"Role"`
	Content string `json:"Content"`
}

// HunyuanChatRequest 混元聊天请求
type HunyuanChatRequest struct {
	Model       string           `json:"Model"`
	Messages    []HunyuanMessage `json:"Messages"`
	Stream      bool             `json:"Stream,omitempty"`
	MaxTokens   int              `json:"MaxTokens,omitempty"`
	Temperature float64          `json:"Temperature,omitempty"`
}

// HunyuanChatResponse 混元聊天响应
type HunyuanChatResponse struct {
	Response struct {
		Choices []struct {
			Index   int `json:"Index"`
			Message struct {
				Role    string `json:"Role"`
				Content string `json:"Content"`
			} `json:"Message"`
			Delta struct {
				Role    string `json:"Role"`
				Content string `json:"Content"`
			} `json:"Delta"`
			FinishReason string `json:"FinishReason"`
		} `json:"Choices"`
		Usage struct {
			PromptTokens     int `json:"PromptTokens"`
			CompletionTokens int `json:"CompletionTokens"`
			TotalTokens      int `json:"TotalTokens"`
		} `json:"Usage"`
		RequestId string `json:"RequestId"`
	} `json:"Response"`
	Error *struct {
		Code    string `json:"Code"`
		Message string `json:"Message"`
	} `json:"Error,omitempty"`
}

// NewHunyuanClient 创建新的混元客户端
func NewHunyuanClient(config *Config) (*HunyuanClient, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	if config.APIKey == "" {
		return nil, fmt.Errorf("Hunyuan API key is required")
	}

	// 解析API Key，格式为：secretId:secretKey
	parts := strings.Split(config.APIKey, ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid Hunyuan API key format, expected 'secretId:secretKey'")
	}

	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = "https://hunyuan.tencentcloudapi.com"
	}

	timeout := config.Timeout
	if timeout <= 0 {
		timeout = 60 // 默认60秒
	}

	return &HunyuanClient{
		config: config,
		httpClient: &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		},
		baseURL:   baseURL,
		secretId:  parts[0],
		secretKey: parts[1],
	}, nil
}

// Chat 阻塞式聊天
func (c *HunyuanClient) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	// 转换为混元格式
	hunyuanReq := c.convertToChatRequest(req, false)

	// 发送请求
	hunyuanResp, err := c.sendChatRequest(ctx, hunyuanReq)
	if err != nil {
		return nil, err
	}

	// 转换响应格式
	return c.convertFromChatResponse(hunyuanResp), nil
}

// StreamChat 流式聊天
func (c *HunyuanClient) StreamChat(ctx context.Context, req *ChatRequest) (<-chan StreamEvent, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	// 转换为混元格式
	hunyuanReq := c.convertToChatRequest(req, true)

	// 发送流式请求
	resp, err := c.sendStreamRequest(ctx, hunyuanReq)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("HTTP %d: request failed", resp.StatusCode)
	}

	// 创建事件通道
	eventChan := make(chan StreamEvent, 10)

	// 启动goroutine处理流式响应
	go c.handleStreamResponse(ctx, resp.Body, eventChan)

	return eventChan, nil
}

// GetProvider 获取提供商类型
func (c *HunyuanClient) GetProvider() LLMProvider {
	return ProviderHunyuan
}

// Close 关闭客户端
func (c *HunyuanClient) Close() error {
	// HTTP客户端不需要显式关闭
	return nil
}

// convertToChatRequest 转换为混元请求格式
func (c *HunyuanClient) convertToChatRequest(req *ChatRequest, stream bool) *HunyuanChatRequest {
	// 转换消息格式
	messages := make([]HunyuanMessage, len(req.Messages))
	for i, msg := range req.Messages {
		messages[i] = HunyuanMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	// 设置默认模型
	model := req.Model
	if model == "" {
		model = c.config.Model
		if model == "" {
			model = "hunyuan-lite"
		}
	}

	// 设置默认参数
	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = c.config.MaxTokens
		if maxTokens == 0 {
			maxTokens = 1000
		}
	}

	temperature := req.Temperature
	if temperature == 0 {
		temperature = c.config.Temperature
		if temperature == 0 {
			temperature = 0.7
		}
	}

	return &HunyuanChatRequest{
		Model:       model,
		Messages:    messages,
		Stream:      stream,
		MaxTokens:   maxTokens,
		Temperature: temperature,
	}
}

// convertFromChatResponse 转换混元响应格式
func (c *HunyuanClient) convertFromChatResponse(hunyuanResp *HunyuanChatResponse) *ChatResponse {
	if hunyuanResp.Error != nil {
		return &ChatResponse{
			Error: &ErrorResponse{
				Code:    hunyuanResp.Error.Code,
				Message: hunyuanResp.Error.Message,
				Type:    "hunyuan_error",
			},
		}
	}

	choices := make([]Choice, len(hunyuanResp.Response.Choices))
	for i, choice := range hunyuanResp.Response.Choices {
		choices[i] = Choice{
			Index: choice.Index,
			Message: &Message{
				Role:    choice.Message.Role,
				Content: choice.Message.Content,
			},
			FinishReason: choice.FinishReason,
		}
	}

	return &ChatResponse{
		ID:      hunyuanResp.Response.RequestId,
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Choices: choices,
		Usage: &TokenUsage{
			PromptTokens:     hunyuanResp.Response.Usage.PromptTokens,
			CompletionTokens: hunyuanResp.Response.Usage.CompletionTokens,
			TotalTokens:      hunyuanResp.Response.Usage.TotalTokens,
		},
	}
}

// sendChatRequest 发送聊天请求
func (c *HunyuanClient) sendChatRequest(ctx context.Context, req *HunyuanChatRequest) (*HunyuanChatResponse, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// 创建HTTP请求
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 设置请求头和签名
	c.setHeaders(httpReq, jsonData)

	// 发送请求
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var hunyuanResp HunyuanChatResponse
	if err := json.Unmarshal(body, &hunyuanResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &hunyuanResp, nil
}

// sendStreamRequest 发送流式请求
func (c *HunyuanClient) sendStreamRequest(ctx context.Context, req *HunyuanChatRequest) (*http.Response, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// 创建HTTP请求
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 设置请求头和签名
	c.setHeaders(httpReq, jsonData)

	// 发送请求
	return c.httpClient.Do(httpReq)
}

// setHeaders 设置请求头和签名
func (c *HunyuanClient) setHeaders(req *http.Request, payload []byte) {
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Host", req.URL.Host)
	req.Header.Set("X-TC-Action", "ChatCompletions")
	req.Header.Set("X-TC-Version", "2023-09-01")
	req.Header.Set("X-TC-Timestamp", timestamp)
	req.Header.Set("Authorization", c.buildAuthorization(req, payload, timestamp))
}

// buildAuthorization 构建腾讯云签名
func (c *HunyuanClient) buildAuthorization(req *http.Request, payload []byte, timestamp string) string {
	// 构建规范请求串
	canonicalRequest := c.buildCanonicalRequest(req, payload)

	// 构建待签名字符串
	credentialScope := fmt.Sprintf("%s/hunyuan/tc3_request", time.Unix(0, 0).UTC().Format("2006-01-02"))
	stringToSign := fmt.Sprintf("TC3-HMAC-SHA256\n%s\n%s\n%s",
		timestamp,
		credentialScope,
		c.sha256Hex([]byte(canonicalRequest)))

	// 计算签名
	signature := c.sign(stringToSign, timestamp)

	// 构建Authorization头
	return fmt.Sprintf("TC3-HMAC-SHA256 Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		c.secretId,
		credentialScope,
		c.getSignedHeaders(req),
		signature)
}

// buildCanonicalRequest 构建规范请求串
func (c *HunyuanClient) buildCanonicalRequest(req *http.Request, payload []byte) string {
	return fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s",
		req.Method,
		req.URL.Path,
		req.URL.RawQuery,
		c.getCanonicalHeaders(req),
		c.getSignedHeaders(req),
		c.sha256Hex(payload))
}

// getCanonicalHeaders 获取规范头部
func (c *HunyuanClient) getCanonicalHeaders(req *http.Request) string {
	var headers []string
	for k, v := range req.Header {
		if len(v) > 0 {
			headers = append(headers, fmt.Sprintf("%s:%s", strings.ToLower(k), strings.TrimSpace(v[0])))
		}
	}
	sort.Strings(headers)
	return strings.Join(headers, "\n") + "\n"
}

// getSignedHeaders 获取签名头部
func (c *HunyuanClient) getSignedHeaders(req *http.Request) string {
	var headers []string
	for k := range req.Header {
		headers = append(headers, strings.ToLower(k))
	}
	sort.Strings(headers)
	return strings.Join(headers, ";")
}

// sign 计算签名
func (c *HunyuanClient) sign(stringToSign, timestamp string) string {
	date := time.Unix(0, 0).UTC().Format("2006-01-02")
	kDate := c.hmacSha256([]byte("TC3"+c.secretKey), date)
	kService := c.hmacSha256(kDate, "hunyuan")
	kSigning := c.hmacSha256(kService, "tc3_request")
	signature := c.hmacSha256(kSigning, stringToSign)
	return hex.EncodeToString(signature)
}

// sha256Hex 计算SHA256哈希
func (c *HunyuanClient) sha256Hex(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// hmacSha256 计算HMAC-SHA256
func (c *HunyuanClient) hmacSha256(key []byte, data string) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(data))
	return mac.Sum(nil)
}

// handleStreamResponse 处理流式响应
func (c *HunyuanClient) handleStreamResponse(ctx context.Context, body io.ReadCloser, eventChan chan<- StreamEvent) {
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

		// 混元的SSE格式：data: {...}
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")

			// 检查是否是结束标志
			if data == "[DONE]" {
				eventChan <- StreamEvent{Done: true}
				return
			}

			var hunyuanResp HunyuanChatResponse
			if err := json.Unmarshal([]byte(data), &hunyuanResp); err != nil {
				eventChan <- StreamEvent{
					Error: fmt.Errorf("failed to decode stream response: %w", err),
					Done:  true,
				}
				return
			}

			// 转换为标准格式
			chatResp := c.convertFromStreamResponse(&hunyuanResp)
			eventChan <- StreamEvent{
				Data: chatResp,
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

// convertFromStreamResponse 转换流式响应
func (c *HunyuanClient) convertFromStreamResponse(hunyuanResp *HunyuanChatResponse) *ChatResponse {
	if hunyuanResp.Error != nil {
		return &ChatResponse{
			Error: &ErrorResponse{
				Code:    hunyuanResp.Error.Code,
				Message: hunyuanResp.Error.Message,
				Type:    "hunyuan_error",
			},
		}
	}

	choices := make([]Choice, len(hunyuanResp.Response.Choices))
	for i, choice := range hunyuanResp.Response.Choices {
		choices[i] = Choice{
			Index: choice.Index,
			Delta: &Message{
				Role:    choice.Delta.Role,
				Content: choice.Delta.Content,
			},
			FinishReason: choice.FinishReason,
		}
	}

	return &ChatResponse{
		ID:      hunyuanResp.Response.RequestId,
		Object:  "chat.completion.chunk",
		Created: time.Now().Unix(),
		Choices: choices,
	}
}
