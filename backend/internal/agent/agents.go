package agent

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
	"smart-analysis/internal/utils/sanbox"
)

// ReactAgent 基于React的智能体
type ReactAgent struct {
	agent     *react.Agent
	config    *AgentConfig
	tools     []tool.BaseTool
	agentType AgentType
}

// NewReactAgent 创建新的React智能体
func NewReactAgent(ctx context.Context, config *AgentConfig) (*ReactAgent, error) {
	// 创建工具
	tools := make([]tool.BaseTool, 0)

	if config.PythonSandbox != nil {
		// 添加Python分析工具（包含统计分析功能）
		pythonTool := NewPythonAnalysisTool(config.PythonSandbox)
		tools = append(tools, pythonTool)

		// 添加ECharts可视化工具
		vizTool := NewEChartsVisualizationTool(config.PythonSandbox)
		tools = append(tools, vizTool)

		// 添加文件读取工具
		fileReaderTool := NewFileReaderTool(config.PythonSandbox)
		tools = append(tools, fileReaderTool)

		// 添加数据查询工具
		queryTool := NewDataQueryTool(config.PythonSandbox)
		tools = append(tools, queryTool)
	}

	// 添加用户提供的工具
	tools = append(tools, config.Tools...)

	// 创建React智能体配置
	reactConfig := &react.AgentConfig{
		ToolCallingModel: config.ChatModel.(model.ToolCallingChatModel),
		ToolsConfig: compose.ToolsNodeConfig{
			Tools: tools,
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

当用户询问数据分析相关问题时：
- 首先理解用户的需求和数据
- 选择合适的工具来完成分析任务
- 为用户提供清晰的分析结果和解释
- 对于图表，优先使用echarts_visualization工具生成可交互的图表配置
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
		tools:     tools,
		agentType: AgentTypeReact,
	}, nil
}

// GetType 获取智能体类型
func (a *ReactAgent) GetType() AgentType {
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

// MainAgent 主智能体
type MainAgent struct {
	chatModel  model.BaseChatModel
	config     *AgentConfig
	agentType  AgentType
	reactAgent *ReactAgent
}

// NewMainAgent 创建新的主智能体
func NewMainAgent(ctx context.Context, config *AgentConfig) (*MainAgent, error) {
	// 创建React智能体作为执行引擎
	reactAgent, err := NewReactAgent(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create react agent: %w", err)
	}

	return &MainAgent{
		chatModel:  config.ChatModel,
		config:     config,
		agentType:  AgentTypeMain,
		reactAgent: reactAgent,
	}, nil
}

// GetType 获取智能体类型
func (a *MainAgent) GetType() AgentType {
	return a.agentType
}

// Generate 生成响应
func (a *MainAgent) Generate(ctx context.Context, messages []*schema.Message, opts ...interface{}) (*schema.Message, error) {
	// 进行意图识别和查询改写
	rewrittenMessages, err := a.processUserIntent(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("failed to process user intent: %w", err)
	}

	// 使用React智能体执行实际的分析任务
	return a.reactAgent.Generate(ctx, rewrittenMessages, opts...)
}

// Stream 流式生成响应
func (a *MainAgent) Stream(ctx context.Context, messages []*schema.Message, opts ...interface{}) (*schema.StreamReader[*schema.Message], error) {
	// 进行意图识别和查询改写
	rewrittenMessages, err := a.processUserIntent(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("failed to process user intent: %w", err)
	}

	// 使用React智能体执行实际的分析任务
	return a.reactAgent.Stream(ctx, rewrittenMessages, opts...)
}

// Initialize 初始化智能体
func (a *MainAgent) Initialize(ctx context.Context) error {
	return a.reactAgent.Initialize(ctx)
}

// Shutdown 关闭智能体
func (a *MainAgent) Shutdown(ctx context.Context) error {
	return a.reactAgent.Shutdown(ctx)
}

// processUserIntent 处理用户意图识别和查询改写
func (a *MainAgent) processUserIntent(ctx context.Context, messages []*schema.Message) ([]*schema.Message, error) {
	if len(messages) == 0 {
		return messages, nil
	}

	// 获取用户的最后一条消息
	lastMessage := messages[len(messages)-1]
	if lastMessage.Role != schema.User {
		return messages, nil
	}

	// 构建意图识别的系统提示
	systemPrompt := `你是一个数据分析意图识别专家。请分析用户的查询，并根据需要进行改写以便更好地处理。

分析用户查询时，请考虑：
1. 用户是否需要数据分析？
2. 用户是否需要数据可视化？
3. 用户是否需要统计分析？
4. 查询是否需要更清晰、具体的表达？

如果查询已经足够清晰，请直接返回原查询。
如果查询需要改写，请提供更清晰、具体的版本。

用户查询: ` + lastMessage.Content

	// 使用LLM进行意图识别和查询改写
	intentMessages := []*schema.Message{
		{
			Role:    schema.System,
			Content: systemPrompt,
		},
		{
			Role:    schema.User,
			Content: "请分析并改写这个查询（如果需要的话）。",
		},
	}

	response, err := a.chatModel.Generate(ctx, intentMessages)
	if err != nil {
		// 如果意图识别失败，返回原始消息
		if a.config.EnableDebug {
			fmt.Printf("Intent recognition failed: %v\n", err)
		}
		return messages, nil
	}

	// 如果改写后的查询更好，使用改写后的版本
	rewrittenContent := strings.TrimSpace(response.Content)
	if rewrittenContent != "" && rewrittenContent != lastMessage.Content {
		// 创建新的消息列表，替换最后一条用户消息
		newMessages := make([]*schema.Message, len(messages))
		copy(newMessages, messages)
		newMessages[len(messages)-1] = &schema.Message{
			Role:    schema.User,
			Content: rewrittenContent,
		}
		return newMessages, nil
	}

	return messages, nil
}

// EinoAnalysisAgent 基于Eino的数据分析智能体（简化版，主要用于特定分析任务）
type EinoAnalysisAgent struct {
	chatModel model.BaseChatModel
	sandbox   *sanbox.PythonSandbox
	config    *AgentConfig
	agentType AgentType
}

// NewEinoAnalysisAgent 创建新的Eino分析智能体
func NewEinoAnalysisAgent(ctx context.Context, config *AgentConfig) (*EinoAnalysisAgent, error) {
	return &EinoAnalysisAgent{
		chatModel: config.ChatModel,
		sandbox:   config.PythonSandbox,
		config:    config,
		agentType: AgentTypeAnalysis,
	}, nil
}

// GetType 获取智能体类型
func (a *EinoAnalysisAgent) GetType() AgentType {
	return a.agentType
}

// Generate 生成响应
func (a *EinoAnalysisAgent) Generate(ctx context.Context, messages []*schema.Message, opts ...interface{}) (*schema.Message, error) {
	// 使用LLM生成Python代码
	code, err := a.generateAnalysisCode(ctx, messages)
	if err != nil {
		return &schema.Message{
			Role:    schema.Assistant,
			Content: fmt.Sprintf("代码生成失败: %v", err),
		}, nil
	}

	// 执行Python代码
	result, err := a.sandbox.ExecutePython(code)
	if err != nil {
		return &schema.Message{
			Role:    schema.Assistant,
			Content: fmt.Sprintf("代码执行失败: %v", err),
		}, nil
	}

	if !result.Success {
		return &schema.Message{
			Role:    schema.Assistant,
			Content: fmt.Sprintf("分析执行失败: %s", result.Error),
		}, nil
	}

	// 格式化结果
	response := "数据分析完成！\n\n"
	if result.Stdout != "" {
		response += "分析结果:\n" + result.Stdout + "\n\n"
	}
	if result.ImagePath != "" {
		response += "生成的图表: " + result.ImagePath + "\n"
	}

	return &schema.Message{
		Role:    schema.Assistant,
		Content: response,
	}, nil
}

// Stream 流式生成响应
func (a *EinoAnalysisAgent) Stream(ctx context.Context, messages []*schema.Message, opts ...interface{}) (*schema.StreamReader[*schema.Message], error) {
	// 简单实现：先生成完整响应，然后流式返回
	response, err := a.Generate(ctx, messages, opts...)
	if err != nil {
		return nil, err
	}

	// 创建流式读取器
	sr, sw := schema.Pipe[*schema.Message](1)
	go func() {
		defer sw.Close()
		sw.Send(response, nil)
	}()

	return sr, nil
}

// Initialize 初始化智能体
func (a *EinoAnalysisAgent) Initialize(ctx context.Context) error {
	return nil
}

// Shutdown 关闭智能体
func (a *EinoAnalysisAgent) Shutdown(ctx context.Context) error {
	return nil
}

// generateAnalysisCode 生成分析代码
func (a *EinoAnalysisAgent) generateAnalysisCode(ctx context.Context, messages []*schema.Message) (string, error) {
	// 构建代码生成的系统提示
	systemPrompt := `你是一个专业的Python数据分析师。请根据用户的需求生成高质量的Python代码。

代码要求：
1. 使用pandas, numpy, matplotlib, seaborn等常用库
2. 代码完整且可执行
3. 包含适当的注释
4. 处理可能的错误情况
5. 如果生成图表，保存为'output.png'

请只返回Python代码，不要添加其他解释。`

	// 构建代码生成消息
	codeMessages := []*schema.Message{
		{
			Role:    schema.System,
			Content: systemPrompt,
		},
	}
	codeMessages = append(codeMessages, messages...)

	// 调用LLM生成代码
	response, err := a.chatModel.Generate(ctx, codeMessages)
	if err != nil {
		return "", err
	}

	// 提取Python代码
	code := a.extractPythonCode(response.Content)
	return code, nil
}

// extractPythonCode 从响应中提取Python代码
func (a *EinoAnalysisAgent) extractPythonCode(content string) string {
	// 查找代码块
	start := strings.Index(content, "```python")
	if start == -1 {
		start = strings.Index(content, "```")
		if start == -1 {
			return strings.TrimSpace(content)
		}
	} else {
		start += len("```python")
	}

	end := strings.Index(content[start:], "```")
	if end == -1 {
		return strings.TrimSpace(content[start:])
	}

	return strings.TrimSpace(content[start : start+end])
}
