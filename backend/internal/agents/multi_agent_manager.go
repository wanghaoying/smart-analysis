package agents

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/schema"
	"smart-analysis/internal/types"
)

// MultiAgentManager 多智能体管理器
type MultiAgentManager struct {
	masterAgent  *MasterAgent
	plannerAgent *PlannerAgent
	config       *types.AgentConfig
	agentType    types.AgentType
}

// NewMultiAgentManager 创建多智能体管理器
func NewMultiAgentManager(ctx context.Context, config *types.AgentConfig) (*MultiAgentManager, error) {
	// 创建主控智能体
	masterAgent, err := NewMasterAgent(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("创建主控智能体失败: %w", err)
	}

	// 创建规划智能体
	plannerAgent, err := NewPlannerAgent(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("创建规划智能体失败: %w", err)
	}

	manager := &MultiAgentManager{
		masterAgent:  masterAgent,
		plannerAgent: plannerAgent,
		config:       config,
		agentType:    types.AgentTypeMulti,
	}

	// 注册专家智能体
	if err := manager.registerExpertAgents(ctx); err != nil {
		return nil, fmt.Errorf("注册专家智能体失败: %w", err)
	}

	return manager, nil
}

// GetType 获取智能体类型
func (m *MultiAgentManager) GetType() types.AgentType {
	return m.agentType
}

// Generate 生成响应
func (m *MultiAgentManager) Generate(ctx context.Context, messages []*schema.Message, opts ...interface{}) (*schema.Message, error) {
	// 第一步：使用MasterAgent进行意图识别和查询重写
	var dataSchema *types.DataSchema
	if len(opts) > 0 {
		if schema, ok := opts[0].(*types.DataSchema); ok {
			dataSchema = schema
		}
	}

	intentResponse, err := m.masterAgent.Generate(ctx, messages, dataSchema)
	if err != nil {
		return &schema.Message{
			Role:    schema.Assistant,
			Content: fmt.Sprintf("意图识别失败: %v", err),
		}, nil
	}

	// 解析意图结果
	var queryIntent types.QueryIntent
	if err := json.Unmarshal([]byte(intentResponse.Content), &queryIntent); err != nil {
		return &schema.Message{
			Role:    schema.Assistant,
			Content: fmt.Sprintf("解析意图失败: %v", err),
		}, nil
	}

	// 第二步：使用PlannerAgent进行任务规划和执行
	plannerResponse, err := m.plannerAgent.Generate(ctx, messages, &queryIntent)
	if err != nil {
		return &schema.Message{
			Role:    schema.Assistant,
			Content: fmt.Sprintf("任务规划执行失败: %v", err),
		}, nil
	}

	return plannerResponse, nil
}

// Stream 流式生成响应
func (m *MultiAgentManager) Stream(ctx context.Context, messages []*schema.Message, opts ...interface{}) (*schema.StreamReader[*schema.Message], error) {
	sr, sw := schema.Pipe[*schema.Message](10)

	go func() {
		defer sw.Close()

		// 发送开始消息
		sw.Send(&schema.Message{
			Role:    schema.Assistant,
			Content: "🚀 启动多智能体分析系统...",
		}, nil)

		// 第一步：意图识别
		sw.Send(&schema.Message{
			Role:    schema.Assistant,
			Content: "🧠 MasterAgent 正在分析您的查询意图...",
		}, nil)

		var dataSchema *types.DataSchema
		if len(opts) > 0 {
			if schema, ok := opts[0].(*types.DataSchema); ok {
				dataSchema = schema
			}
		}

		intentResponse, err := m.masterAgent.Generate(ctx, messages, dataSchema)
		if err != nil {
			sw.Send(&schema.Message{
				Role:    schema.Assistant,
				Content: fmt.Sprintf("❌ 意图识别失败: %v", err),
			}, nil)
			return
		}

		// 解析意图结果
		var queryIntent types.QueryIntent
		if err := json.Unmarshal([]byte(intentResponse.Content), &queryIntent); err != nil {
			sw.Send(&schema.Message{
				Role:    schema.Assistant,
				Content: fmt.Sprintf("❌ 解析意图失败: %v", err),
			}, nil)
			return
		}

		sw.Send(&schema.Message{
			Role:    schema.Assistant,
			Content: fmt.Sprintf("✅ 意图识别完成，识别为: %s", queryIntent.IntentType),
		}, nil)

		// 第二步：任务规划和执行
		sw.Send(&schema.Message{
			Role:    schema.Assistant,
			Content: "📋 PlannerAgent 正在创建执行计划...",
		}, nil)

		// 使用流式执行
		plannerStream, err := m.plannerAgent.Stream(ctx, messages, &queryIntent)
		if err != nil {
			sw.Send(&schema.Message{
				Role:    schema.Assistant,
				Content: fmt.Sprintf("❌ 任务规划失败: %v", err),
			}, nil)
			return
		}

		// 转发PlannerAgent的流式输出
		for {
			response, err := plannerStream.Recv()
			if err != nil {
				break
			}
			sw.Send(response, nil)
		}
	}()

	return sr, nil
}

// Initialize 初始化智能体
func (m *MultiAgentManager) Initialize(ctx context.Context) error {
	// 初始化主控智能体
	if err := m.masterAgent.Initialize(ctx); err != nil {
		return fmt.Errorf("初始化主控智能体失败: %w", err)
	}

	// 初始化规划智能体
	if err := m.plannerAgent.Initialize(ctx); err != nil {
		return fmt.Errorf("初始化规划智能体失败: %w", err)
	}

	return nil
}

// Shutdown 关闭智能体
func (m *MultiAgentManager) Shutdown(ctx context.Context) error {
	// 关闭规划智能体
	if err := m.plannerAgent.Shutdown(ctx); err != nil {
		return fmt.Errorf("关闭规划智能体失败: %w", err)
	}

	// 关闭主控智能体
	if err := m.masterAgent.Shutdown(ctx); err != nil {
		return fmt.Errorf("关闭主控智能体失败: %w", err)
	}

	return nil
}

// registerExpertAgents 注册专家智能体
func (m *MultiAgentManager) registerExpertAgents(ctx context.Context) error {
	// 创建并注册数据查询专家智能体
	dataQueryAgent, err := NewDataQueryAgent(ctx, m.config)
	if err != nil {
		return fmt.Errorf("创建数据查询专家智能体失败: %w", err)
	}
	m.plannerAgent.RegisterExpertAgent(dataQueryAgent)

	// 创建并注册数据分析专家智能体
	dataAnalysisAgent, err := NewDataAnalysisAgent(ctx, m.config)
	if err != nil {
		return fmt.Errorf("创建数据分析专家智能体失败: %w", err)
	}
	m.plannerAgent.RegisterExpertAgent(dataAnalysisAgent)

	// 创建并注册趋势预测专家智能体
	trendForecastAgent, err := NewTrendForecastAgent(ctx, m.config)
	if err != nil {
		return fmt.Errorf("创建趋势预测专家智能体失败: %w", err)
	}
	m.plannerAgent.RegisterExpertAgent(trendForecastAgent)

	// 创建并注册异动检测专家智能体
	anomalyDetectionAgent, err := NewAnomalyDetectionAgent(ctx, m.config)
	if err != nil {
		return fmt.Errorf("创建异动检测专家智能体失败: %w", err)
	}
	m.plannerAgent.RegisterExpertAgent(anomalyDetectionAgent)

	// 创建并注册归因分析专家智能体
	attributionAnalysisAgent, err := NewAttributionAnalysisAgent(ctx, m.config)
	if err != nil {
		return fmt.Errorf("创建归因分析专家智能体失败: %w", err)
	}
	m.plannerAgent.RegisterExpertAgent(attributionAnalysisAgent)

	return nil
}

// GetMasterAgent 获取主控智能体
func (m *MultiAgentManager) GetMasterAgent() *MasterAgent {
	return m.masterAgent
}

// GetPlannerAgent 获取规划智能体
func (m *MultiAgentManager) GetPlannerAgent() *PlannerAgent {
	return m.plannerAgent
}

// AddDataSchema 添加数据模式信息
func (m *MultiAgentManager) AddDataSchema(schema *types.DataSchema) {
	// 可以在这里保存数据模式信息，供后续使用
	if m.config.Metadata == nil {
		m.config.Metadata = make(map[string]interface{})
	}
	m.config.Metadata["data_schema"] = schema
}
