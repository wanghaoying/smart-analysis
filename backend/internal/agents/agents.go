package agents

import (
	"context"
	"fmt"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
	"smart-analysis/internal/tools"
	"smart-analysis/internal/types"
)

// ReactAgent 基于React的智能体
type ReactAgent struct {
	agent     *react.Agent
	config    *types.AgentConfig
	tools     []tool.BaseTool
	agentType types.AgentType
}

// NewReactAgent 创建新的React智能体
func NewReactAgent(ctx context.Context, config *types.AgentConfig) (*ReactAgent, error) {
	// 创建工具
	toolsList := make([]tool.BaseTool, 0)

	if config.PythonSandbox != nil {
		// 添加Python分析工具（包含统计分析功能）
		pythonTool := tools.NewPythonAnalysisTool(config.PythonSandbox)
		toolsList = append(toolsList, pythonTool)

		// 添加ECharts可视化工具
		vizTool := tools.NewEChartsVisualizationTool(config.PythonSandbox)
		toolsList = append(toolsList, vizTool)

		// 添加文件读取工具
		fileReaderTool := tools.NewFileReaderTool(config.PythonSandbox)
		toolsList = append(toolsList, fileReaderTool)

		// 添加数据查询工具
		queryTool := tools.NewDataQueryTool(config.PythonSandbox)
		toolsList = append(toolsList, queryTool)

		// 添加数据预处理工具
		preprocessingTool := tools.NewDataPreprocessingTool(config.PythonSandbox)
		toolsList = append(toolsList, preprocessingTool)

		// 添加机器学习分析工具
		mlTool := tools.NewMLAnalysisTool(config.PythonSandbox)
		toolsList = append(toolsList, mlTool)

		// 添加文本分析工具
		textTool := tools.NewTextAnalysisTool(config.PythonSandbox)
		toolsList = append(toolsList, textTool)

		// 添加报告生成工具
		reportTool := tools.NewReportGeneratorTool(config.PythonSandbox)
		toolsList = append(toolsList, reportTool)
	}

	// 添加用户提供的工具
	toolsList = append(toolsList, config.Tools...)

	// 创建React智能体配置
	reactConfig := &react.AgentConfig{
		ToolCallingModel: config.ChatModel.(model.ToolCallingChatModel),
		ToolsConfig: compose.ToolsNodeConfig{
			Tools: toolsList,
		},
		MaxStep: config.MaxSteps,
		MessageModifier: func(ctx context.Context, input []*schema.Message) []*schema.Message {
			// 添加系统提示来增强数据分析能力
			systemMessage := &schema.Message{
				Role: schema.System,
				Content: `你是一个专业的数据分析助手。你拥有以下能力：

1. 使用python_analysis工具执行Python代码进行数据分析和统计计算
2. 使用echarts_visualization工具创建ECharts格式的交互式图表
3. 使用file_reader工具读取和预览数据文件
4. 使用data_query工具进行数据查询和筛选
5. 使用data_preprocessing工具进行数据预处理和特征工程
6. 使用ml_analysis工具进行机器学习分析（分类、回归、聚类）

当用户询问数据分析相关问题时：
- 首先理解用户的需求和数据
- 选择合适的工具来完成分析任务
- 为用户提供清晰的分析结果和解释
- 对于图表，优先使用echarts_visualization工具生成可交互的图表配置
- 对于复杂的数据处理，可以先使用data_preprocessing工具预处理数据
- 对于机器学习任务，使用ml_analysis工具进行建模和评估
- 如果需要，可以建议进一步的分析方向

请始终保持专业、准确，并提供有价值的洞察。`,
			}

			// 检查是否已经有系统消息
			hasSystemMessage := false
			for _, msg := range input {
				if msg.Role == schema.System {
					hasSystemMessage = true
					break
				}
			}

			if !hasSystemMessage {
				return append([]*schema.Message{systemMessage}, input...)
			}

			return input
		},
	}

	// 创建React智能体
	agent, err := react.NewAgent(ctx, reactConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create react agent: %w", err)
	}

	return &ReactAgent{
		agent:     agent,
		config:    config,
		tools:     toolsList,
		agentType: types.AgentTypeReact,
	}, nil
}

// GetType 获取智能体类型
func (a *ReactAgent) GetType() types.AgentType {
	return a.agentType
}

// Generate 生成响应
func (a *ReactAgent) Generate(ctx context.Context, messages []*schema.Message, opts ...interface{}) (*schema.Message, error) {
	return a.agent.Generate(ctx, messages)
}

// Stream 流式生成响应
func (a *ReactAgent) Stream(ctx context.Context, messages []*schema.Message, opts ...interface{}) (*schema.StreamReader[*schema.Message], error) {
	return a.agent.Stream(ctx, messages)
}

// Initialize 初始化智能体
func (a *ReactAgent) Initialize(ctx context.Context) error {
	// React智能体已经在创建时初始化
	return nil
}

// Shutdown 关闭智能体
func (a *ReactAgent) Shutdown(ctx context.Context) error {
	// 清理资源
	return nil
}

// MainAgent 主智能体（为了兼容性保留，内部使用MultiAgentManager）
type MainAgent struct {
	manager   *MultiAgentManager
	config    *types.AgentConfig
	agentType types.AgentType
}

// NewMainAgent 创建新的主智能体（兼容性接口）
func NewMainAgent(ctx context.Context, config *types.AgentConfig) (*MainAgent, error) {
	// 创建多智能体管理器
	manager, err := NewMultiAgentManager(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("创建多智能体管理器失败: %w", err)
	}

	return &MainAgent{
		manager:   manager,
		config:    config,
		agentType: types.AgentTypeMaster, // 更新为新的类型
	}, nil
}

// GetType 获取智能体类型
func (a *MainAgent) GetType() types.AgentType {
	return a.agentType
}

// Generate 生成响应
func (a *MainAgent) Generate(ctx context.Context, messages []*schema.Message, opts ...interface{}) (*schema.Message, error) {
	return a.manager.Generate(ctx, messages, opts...)
}

// Stream 流式生成响应
func (a *MainAgent) Stream(ctx context.Context, messages []*schema.Message, opts ...interface{}) (*schema.StreamReader[*schema.Message], error) {
	return a.manager.Stream(ctx, messages, opts...)
}

// Initialize 初始化智能体
func (a *MainAgent) Initialize(ctx context.Context) error {
	return a.manager.Initialize(ctx)
}

// Shutdown 关闭智能体
func (a *MainAgent) Shutdown(ctx context.Context) error {
	return a.manager.Shutdown(ctx)
}
