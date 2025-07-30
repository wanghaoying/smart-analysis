package agent

import (
	"context"
	"strings"
	"testing"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"smart-analysis/internal/utils/llm"
	"smart-analysis/internal/utils/sanbox"
)

// MockEinoLLMModel 模拟的Eino LLM模型
type MockEinoLLMModel struct {
	responses map[string]string
	tools     []*schema.ToolInfo
}

// NewMockEinoLLMModel 创建新的模拟LLM模型
func NewMockEinoLLMModel() *MockEinoLLMModel {
	return &MockEinoLLMModel{
		responses: map[string]string{
			"分析数据":         "我将为您分析数据。请提供需要分析的数据文件或数据内容。",
			"create chart": "我将为您创建图表。请使用数据可视化工具。",
			"统计分析":         "我将进行统计分析。让我使用统计分析工具来处理数据。",
			"hello":        "Hello! I'm ready to help with data analysis.",
		},
		tools: nil,
	}
}

// Generate 生成响应
func (m *MockEinoLLMModel) Generate(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.Message, error) {
	if len(messages) == 0 {
		return &schema.Message{
			Role:    schema.Assistant,
			Content: "No input provided",
		}, nil
	}

	lastMessage := messages[len(messages)-1]

	// 查找匹配的响应
	for key, response := range m.responses {
		if containsEino(lastMessage.Content, key) {
			return &schema.Message{
				Role:    schema.Assistant,
				Content: response,
			}, nil
		}
	}

	// 默认响应
	return &schema.Message{
		Role:    schema.Assistant,
		Content: "I understand your request. Let me help you with data analysis.",
	}, nil
}

// Stream 流式生成响应
func (m *MockEinoLLMModel) Stream(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	response, err := m.Generate(ctx, messages, opts...)
	if err != nil {
		return nil, err
	}

	sr, sw := schema.Pipe[*schema.Message](1)
	go func() {
		defer sw.Close()
		sw.Send(response, nil)
	}()

	return sr, nil
}

// WithTools 实现ToolCallingChatModel接口
func (m *MockEinoLLMModel) WithTools(tools []*schema.ToolInfo) (model.ToolCallingChatModel, error) {
	return &MockEinoLLMModel{
		responses: m.responses,
		tools:     tools,
	}, nil
}

// containsEino 检查字符串是否包含子字符串（不区分大小写）
func containsEino(s, substr string) bool {
	s = strings.ToLower(s)
	substr = strings.ToLower(substr)
	return strings.Contains(s, substr)
}

// TestEinoAgentManager 测试Eino智能体管理器
func TestEinoAgentManager(t *testing.T) {
	ctx := context.Background()

	// 创建模拟组件
	mockModel := NewMockEinoLLMModel()
	mockSandbox := &sanbox.PythonSandbox{}

	// 使用构建器创建系统
	manager, err := NewEinoAgentSystemBuilder().
		WithChatModel(mockModel).
		WithPythonSandbox(mockSandbox).
		WithDebug(true).
		Build(ctx)
	if err != nil {
		t.Fatalf("Failed to create eino agent manager: %v", err)
	}

	// 创建系统
	config := &EinoAgentConfig{
		MaxSteps:    10,
		EnableDebug: true,
	}
	system := NewEinoAgentSystem(manager, config)
	defer system.Shutdown(ctx)

	// 测试查询处理
	t.Run("ProcessQuery", func(t *testing.T) {
		response, err := system.ProcessQuery(ctx, "分析数据")
		if err != nil {
			t.Errorf("ProcessQuery failed: %v", err)
			return
		}

		if response == nil {
			t.Error("Response is nil")
			return
		}

		if response.Content == "" {
			t.Error("Response content is empty")
		}

		t.Logf("Response: %s", response.Content)
	})

	// 测试流式查询
	t.Run("StreamQuery", func(t *testing.T) {
		stream, err := system.StreamQuery(ctx, "hello")
		if err != nil {
			t.Errorf("StreamQuery failed: %v", err)
			return
		}

		// 读取流式响应
		for {
			response, err := stream.Recv()
			if err != nil {
				if err.Error() == "EOF" {
					break
				}
				t.Errorf("Stream recv failed: %v", err)
				break
			}

			if response != nil {
				t.Logf("Stream response: %s", response.Content)
			}
		}

		stream.Close()
	})
}

// TestEinoAgentSystemBuilder 测试构建器
func TestEinoAgentSystemBuilder(t *testing.T) {
	ctx := context.Background()

	mockModel := NewMockEinoLLMModel()
	mockSandbox := &sanbox.PythonSandbox{}

	// 测试构建过程
	t.Run("Builder", func(t *testing.T) {
		manager, err := NewEinoAgentSystemBuilder().
			WithChatModel(mockModel).
			WithPythonSandbox(mockSandbox).
			WithMaxSteps(5).
			WithDebug(true).
			Build(ctx)

		if err != nil {
			t.Errorf("Builder failed: %v", err)
			return
		}

		defer manager.Shutdown(ctx)

		// 验证智能体是否正确注册
		mainAgent, exists := manager.GetAgent(EinoAgentTypeMain)
		if !exists {
			t.Error("Main agent not registered")
		} else {
			if mainAgent.GetType() != EinoAgentTypeMain {
				t.Error("Main agent type mismatch")
			}
		}

		reactAgent, exists := manager.GetAgent(EinoAgentTypeReact)
		if !exists {
			t.Error("React agent not registered")
		} else {
			if reactAgent.GetType() != EinoAgentTypeReact {
				t.Error("React agent type mismatch")
			}
		}

		analysisAgent, exists := manager.GetAgent(EinoAgentTypeAnalysis)
		if !exists {
			t.Error("Analysis agent not registered")
		} else {
			if analysisAgent.GetType() != EinoAgentTypeAnalysis {
				t.Error("Analysis agent type mismatch")
			}
		}
	})
}

// TestEinoIntegration 测试集成
// MockLLMClientForEino 为Eino测试准备的模拟LLM客户端
type MockLLMClientForEino struct {
	model *MockEinoLLMModel
}

// Chat 实现LLM客户端接口
func (m *MockLLMClientForEino) Chat(ctx context.Context, req *llm.ChatRequest) (*llm.ChatResponse, error) {
	// 转换消息格式
	einoMessages := make([]*schema.Message, len(req.Messages))
	for i, msg := range req.Messages {
		einoMessages[i] = &schema.Message{
			Role:    m.convertRole(msg.Role),
			Content: msg.Content,
		}
	}

	// 调用模拟模型
	response, err := m.model.Generate(ctx, einoMessages)
	if err != nil {
		return nil, err
	}

	// 转换响应格式
	return &llm.ChatResponse{
		Choices: []llm.Choice{
			{
				Message: &llm.Message{
					Role:    "assistant",
					Content: response.Content,
				},
			},
		},
	}, nil
}

// StreamChat 实现流式聊天
func (m *MockLLMClientForEino) StreamChat(ctx context.Context, req *llm.ChatRequest) (<-chan llm.StreamEvent, error) {
	// 先生成完整响应
	response, err := m.Chat(ctx, req)
	if err != nil {
		return nil, err
	}

	// 创建流式通道
	streamChan := make(chan llm.StreamEvent, 2)

	go func() {
		defer close(streamChan)

		// 发送响应
		streamChan <- llm.StreamEvent{
			Data: response,
			Done: false,
		}

		// 发送结束标记
		streamChan <- llm.StreamEvent{
			Done: true,
		}
	}()

	return streamChan, nil
}

// GetProvider 获取提供商
func (m *MockLLMClientForEino) GetProvider() llm.LLMProvider {
	return "mock"
}

// Close 关闭客户端
func (m *MockLLMClientForEino) Close() error {
	return nil
}

// convertRole 转换角色
func (m *MockLLMClientForEino) convertRole(role string) schema.RoleType {
	switch role {
	case "system":
		return schema.System
	case "user":
		return schema.User
	case "assistant":
		return schema.Assistant
	default:
		return schema.User
	}
}
