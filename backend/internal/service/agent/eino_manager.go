package agent

import (
	"context"
	"fmt"
	"sync"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"smart-analysis/internal/utils/sanbox"
)

// EinoAgentManager Eino智能体管理器
type EinoAgentManager struct {
	agents map[EinoAgentType]EinoAgent
	config *EinoAgentConfig
	mu     sync.RWMutex
}

// NewEinoAgentManager 创建新的Eino智能体管理器
func NewEinoAgentManager(config *EinoAgentConfig) *EinoAgentManager {
	return &EinoAgentManager{
		agents: make(map[EinoAgentType]EinoAgent),
		config: config,
	}
}

// RegisterAgent 注册智能体
func (m *EinoAgentManager) RegisterAgent(agentType EinoAgentType, agent EinoAgent) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.agents[agentType] = agent
}

// GetAgent 获取智能体
func (m *EinoAgentManager) GetAgent(agentType EinoAgentType) (EinoAgent, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	agent, exists := m.agents[agentType]
	return agent, exists
}

// GetMainAgent 获取主智能体
func (m *EinoAgentManager) GetMainAgent() (EinoAgent, error) {
	agent, exists := m.GetAgent(EinoAgentTypeMain)
	if !exists {
		return nil, fmt.Errorf("main agent not found")
	}
	return agent, nil
}

// ProcessQuery 处理查询
func (m *EinoAgentManager) ProcessQuery(ctx context.Context, query string) (*schema.Message, error) {
	mainAgent, err := m.GetMainAgent()
	if err != nil {
		return nil, err
	}

	messages := []*schema.Message{
		{
			Role:    schema.User,
			Content: query,
		},
	}

	return mainAgent.Generate(ctx, messages)
}

// ProcessQueryWithHistory 处理带历史记录的查询
func (m *EinoAgentManager) ProcessQueryWithHistory(ctx context.Context, messages []*schema.Message) (*schema.Message, error) {
	mainAgent, err := m.GetMainAgent()
	if err != nil {
		return nil, err
	}

	return mainAgent.Generate(ctx, messages)
}

// StreamQuery 流式处理查询
func (m *EinoAgentManager) StreamQuery(ctx context.Context, query string) (*schema.StreamReader[*schema.Message], error) {
	mainAgent, err := m.GetMainAgent()
	if err != nil {
		return nil, err
	}

	messages := []*schema.Message{
		{
			Role:    schema.User,
			Content: query,
		},
	}

	return mainAgent.Stream(ctx, messages)
}

// StreamQueryWithHistory 流式处理带历史记录的查询
func (m *EinoAgentManager) StreamQueryWithHistory(ctx context.Context, messages []*schema.Message) (*schema.StreamReader[*schema.Message], error) {
	mainAgent, err := m.GetMainAgent()
	if err != nil {
		return nil, err
	}

	return mainAgent.Stream(ctx, messages)
}

// Initialize 初始化所有智能体
func (m *EinoAgentManager) Initialize(ctx context.Context) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for agentType, agent := range m.agents {
		if err := agent.Initialize(ctx); err != nil {
			return fmt.Errorf("failed to initialize agent %s: %w", agentType, err)
		}

		if m.config.EnableDebug {
			fmt.Printf("Initialized agent: %s\n", agentType)
		}
	}

	return nil
}

// Shutdown 关闭所有智能体
func (m *EinoAgentManager) Shutdown(ctx context.Context) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var errors []error
	for agentType, agent := range m.agents {
		if err := agent.Shutdown(ctx); err != nil {
			errors = append(errors, fmt.Errorf("failed to shutdown agent %s: %w", agentType, err))
		}

		if m.config.EnableDebug {
			fmt.Printf("Shutdown agent: %s\n", agentType)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("shutdown errors: %v", errors)
	}

	return nil
}

// EinoAgentSystemBuilder 基于Eino的智能体系统构建器
type EinoAgentSystemBuilder struct {
	chatModel     model.BaseChatModel
	pythonSandbox *sanbox.PythonSandbox
	tools         []tool.BaseTool
	maxSteps      int
	enableDebug   bool
}

// NewEinoAgentSystemBuilder 创建新的Eino智能体系统构建器
func NewEinoAgentSystemBuilder() *EinoAgentSystemBuilder {
	return &EinoAgentSystemBuilder{
		maxSteps: 10, // 默认最大步数
	}
}

// WithChatModel 设置聊天模型
func (b *EinoAgentSystemBuilder) WithChatModel(model model.BaseChatModel) *EinoAgentSystemBuilder {
	b.chatModel = model
	return b
}

// WithPythonSandbox 设置Python沙箱
func (b *EinoAgentSystemBuilder) WithPythonSandbox(sandbox *sanbox.PythonSandbox) *EinoAgentSystemBuilder {
	b.pythonSandbox = sandbox
	return b
}

// WithTools 设置工具
func (b *EinoAgentSystemBuilder) WithTools(tools []tool.BaseTool) *EinoAgentSystemBuilder {
	b.tools = tools
	return b
}

// WithMaxSteps 设置最大步数
func (b *EinoAgentSystemBuilder) WithMaxSteps(maxSteps int) *EinoAgentSystemBuilder {
	b.maxSteps = maxSteps
	return b
}

// WithDebug 启用调试模式
func (b *EinoAgentSystemBuilder) WithDebug(enable bool) *EinoAgentSystemBuilder {
	b.enableDebug = enable
	return b
}

// Build 构建智能体系统
func (b *EinoAgentSystemBuilder) Build(ctx context.Context) (*EinoAgentManager, error) {
	// 验证必需的组件
	if b.chatModel == nil {
		return nil, fmt.Errorf("chat model is required")
	}

	// 创建配置
	config := &EinoAgentConfig{
		ChatModel:     b.chatModel,
		PythonSandbox: b.pythonSandbox,
		Tools:         b.tools,
		MaxSteps:      b.maxSteps,
		EnableDebug:   b.enableDebug,
	}

	// 创建管理器
	manager := NewEinoAgentManager(config)

	// 创建主智能体
	mainAgent, err := NewEinoMainAgent(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create main agent: %w", err)
	}
	manager.RegisterAgent(EinoAgentTypeMain, mainAgent)

	// 创建React智能体
	reactAgent, err := NewEinoReactAgent(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create react agent: %w", err)
	}
	manager.RegisterAgent(EinoAgentTypeReact, reactAgent)

	// 创建分析智能体
	if b.pythonSandbox != nil {
		analysisAgent, err := NewEinoAnalysisAgent(ctx, config)
		if err != nil {
			return nil, fmt.Errorf("failed to create analysis agent: %w", err)
		}
		manager.RegisterAgent(EinoAgentTypeAnalysis, analysisAgent)
	}

	// 初始化所有智能体
	if err := manager.Initialize(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize agent system: %w", err)
	}

	if b.enableDebug {
		fmt.Println("Eino agent system built successfully")
	}

	return manager, nil
}

// EinoAgentSystem 基于Eino的智能体系统（兼容性接口）
type EinoAgentSystem struct {
	manager *EinoAgentManager
	config  *EinoAgentConfig
}

// NewEinoAgentSystem 创建新的Eino智能体系统
func NewEinoAgentSystem(manager *EinoAgentManager, config *EinoAgentConfig) *EinoAgentSystem {
	return &EinoAgentSystem{
		manager: manager,
		config:  config,
	}
}

// ProcessQuery 处理查询（兼容性方法）
func (s *EinoAgentSystem) ProcessQuery(ctx context.Context, query string) (*schema.Message, error) {
	return s.manager.ProcessQuery(ctx, query)
}

// ProcessQueryWithHistory 处理带历史记录的查询（兼容性方法）
func (s *EinoAgentSystem) ProcessQueryWithHistory(ctx context.Context, messages []*schema.Message) (*schema.Message, error) {
	return s.manager.ProcessQueryWithHistory(ctx, messages)
}

// StreamQuery 流式处理查询（兼容性方法）
func (s *EinoAgentSystem) StreamQuery(ctx context.Context, query string) (*schema.StreamReader[*schema.Message], error) {
	return s.manager.StreamQuery(ctx, query)
}

// StreamQueryWithHistory 流式处理带历史记录的查询（兼容性方法）
func (s *EinoAgentSystem) StreamQueryWithHistory(ctx context.Context, messages []*schema.Message) (*schema.StreamReader[*schema.Message], error) {
	return s.manager.StreamQueryWithHistory(ctx, messages)
}

// GetManager 获取管理器
func (s *EinoAgentSystem) GetManager() *EinoAgentManager {
	return s.manager
}

// Shutdown 关闭系统
func (s *EinoAgentSystem) Shutdown(ctx context.Context) error {
	return s.manager.Shutdown(ctx)
}
