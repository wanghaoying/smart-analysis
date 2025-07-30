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

// AgentManager Eino智能体管理器
type AgentManager struct {
	agents map[AgentType]Agent
	config *AgentConfig
	mu     sync.RWMutex
}

// NewAgentManager 创建新的Eino智能体管理器
func NewAgentManager(config *AgentConfig) *AgentManager {
	return &AgentManager{
		agents: make(map[AgentType]Agent),
		config: config,
	}
}

// RegisterAgent 注册智能体
func (m *AgentManager) RegisterAgent(agentType AgentType, agent Agent) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.agents[agentType] = agent
}

// GetAgent 获取智能体
func (m *AgentManager) GetAgent(agentType AgentType) (Agent, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	agent, exists := m.agents[agentType]
	return agent, exists
}

// GetMainAgent 获取主智能体
func (m *AgentManager) GetMainAgent() (Agent, error) {
	agent, exists := m.GetAgent(AgentTypeMain)
	if !exists {
		return nil, fmt.Errorf("main agent not found")
	}
	return agent, nil
}

// ProcessQuery 处理查询
func (m *AgentManager) ProcessQuery(ctx context.Context, query string) (*schema.Message, error) {
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
func (m *AgentManager) ProcessQueryWithHistory(ctx context.Context, messages []*schema.Message) (*schema.Message, error) {
	mainAgent, err := m.GetMainAgent()
	if err != nil {
		return nil, err
	}

	return mainAgent.Generate(ctx, messages)
}

// StreamQuery 流式处理查询
func (m *AgentManager) StreamQuery(ctx context.Context, query string) (*schema.StreamReader[*schema.Message], error) {
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
func (m *AgentManager) StreamQueryWithHistory(ctx context.Context, messages []*schema.Message) (*schema.StreamReader[*schema.Message], error) {
	mainAgent, err := m.GetMainAgent()
	if err != nil {
		return nil, err
	}

	return mainAgent.Stream(ctx, messages)
}

// Initialize 初始化所有智能体
func (m *AgentManager) Initialize(ctx context.Context) error {
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
func (m *AgentManager) Shutdown(ctx context.Context) error {
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

// AgentSystemBuilder 基于Eino的智能体系统构建器
type AgentSystemBuilder struct {
	chatModel     model.BaseChatModel
	pythonSandbox *sanbox.PythonSandbox
	tools         []tool.BaseTool
	maxSteps      int
	enableDebug   bool
}

// NewAgentSystemBuilder 创建新的Eino智能体系统构建器
func NewAgentSystemBuilder() *AgentSystemBuilder {
	return &AgentSystemBuilder{
		maxSteps: 10, // 默认最大步数
	}
}

// WithChatModel 设置聊天模型
func (b *AgentSystemBuilder) WithChatModel(model model.BaseChatModel) *AgentSystemBuilder {
	b.chatModel = model
	return b
}

// WithPythonSandbox 设置Python沙箱
func (b *AgentSystemBuilder) WithPythonSandbox(sandbox *sanbox.PythonSandbox) *AgentSystemBuilder {
	b.pythonSandbox = sandbox
	return b
}

// WithTools 设置工具
func (b *AgentSystemBuilder) WithTools(tools []tool.BaseTool) *AgentSystemBuilder {
	b.tools = tools
	return b
}

// WithMaxSteps 设置最大步数
func (b *AgentSystemBuilder) WithMaxSteps(maxSteps int) *AgentSystemBuilder {
	b.maxSteps = maxSteps
	return b
}

// WithDebug 启用调试模式
func (b *AgentSystemBuilder) WithDebug(enable bool) *AgentSystemBuilder {
	b.enableDebug = enable
	return b
}

// Build 构建智能体系统
func (b *AgentSystemBuilder) Build(ctx context.Context) (*AgentManager, error) {
	// 验证必需的组件
	if b.chatModel == nil {
		return nil, fmt.Errorf("chat model is required")
	}

	// 创建配置
	config := &AgentConfig{
		ChatModel:     b.chatModel,
		PythonSandbox: b.pythonSandbox,
		Tools:         b.tools,
		MaxSteps:      b.maxSteps,
		EnableDebug:   b.enableDebug,
	}

	// 创建管理器
	manager := NewAgentManager(config)

	// 创建主智能体
	mainAgent, err := NewMainAgent(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create main agent: %w", err)
	}
	manager.RegisterAgent(AgentTypeMain, mainAgent)

	// 创建React智能体
	reactAgent, err := NewReactAgent(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create react agent: %w", err)
	}
	manager.RegisterAgent(AgentTypeReact, reactAgent)

	// 创建分析智能体
	if b.pythonSandbox != nil {
		analysisAgent, err := NewEinoAnalysisAgent(ctx, config)
		if err != nil {
			return nil, fmt.Errorf("failed to create analysis agent: %w", err)
		}
		manager.RegisterAgent(AgentTypeAnalysis, analysisAgent)
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

// AgentSystem 基于Eino的智能体系统（兼容性接口）
type AgentSystem struct {
	manager *AgentManager
	config  *AgentConfig
}

// NewAgentSystem 创建新的Eino智能体系统
func NewAgentSystem(manager *AgentManager, config *AgentConfig) *AgentSystem {
	return &AgentSystem{
		manager: manager,
		config:  config,
	}
}

// ProcessQuery 处理查询（兼容性方法）
func (s *AgentSystem) ProcessQuery(ctx context.Context, query string) (*schema.Message, error) {
	return s.manager.ProcessQuery(ctx, query)
}

// ProcessQueryWithHistory 处理带历史记录的查询（兼容性方法）
func (s *AgentSystem) ProcessQueryWithHistory(ctx context.Context, messages []*schema.Message) (*schema.Message, error) {
	return s.manager.ProcessQueryWithHistory(ctx, messages)
}

// StreamQuery 流式处理查询（兼容性方法）
func (s *AgentSystem) StreamQuery(ctx context.Context, query string) (*schema.StreamReader[*schema.Message], error) {
	return s.manager.StreamQuery(ctx, query)
}

// StreamQueryWithHistory 流式处理带历史记录的查询（兼容性方法）
func (s *AgentSystem) StreamQueryWithHistory(ctx context.Context, messages []*schema.Message) (*schema.StreamReader[*schema.Message], error) {
	return s.manager.StreamQueryWithHistory(ctx, messages)
}

// GetManager 获取管理器
func (s *AgentSystem) GetManager() *AgentManager {
	return s.manager
}

// Shutdown 关闭系统
func (s *AgentSystem) Shutdown(ctx context.Context) error {
	return s.manager.Shutdown(ctx)
}
