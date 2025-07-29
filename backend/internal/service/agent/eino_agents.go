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

// EinoReactAgent 基于Eino的React智能体
type EinoReactAgent struct {
	agent     *react.Agent
	config    *EinoAgentConfig
	tools     []tool.BaseTool
	agentType EinoAgentType
}

// NewEinoReactAgent 创建新的Eino React智能体
func NewEinoReactAgent(ctx context.Context, config *EinoAgentConfig) (*EinoReactAgent, error) {
	// 创建工具
	tools := make([]tool.BaseTool, 0)

	if config.PythonSandbox != nil {
		// 添加Python分析工具
		pythonTool := NewPythonAnalysisTool(config.PythonSandbox)
		tools = append(tools, pythonTool)

		// 添加数据可视化工具
		vizTool := NewDataVisualizationTool(config.PythonSandbox)
		tools = append(tools, vizTool)

		// 添加统计分析工具
		statTool := NewStatisticalAnalysisTool(config.PythonSandbox)
		tools = append(tools, statTool)
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

1. 使用python_analysis工具执行Python代码进行数据分析
2. 使用data_visualization工具创建各种数据可视化图表
3. 使用statistical_analysis工具进行统计分析

当用户询问数据分析相关问题时：
- 首先理解用户的需求和数据
- 选择合适的工具来完成分析任务
- 为用户提供清晰的分析结果和解释
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

	return &EinoReactAgent{
		agent:     agent,
		config:    config,
		tools:     tools,
		agentType: EinoAgentTypeReact,
	}, nil
}

// GetType 获取智能体类型
func (a *EinoReactAgent) GetType() EinoAgentType {
	return a.agentType
}

// Generate 生成响应
func (a *EinoReactAgent) Generate(ctx context.Context, messages []*schema.Message, opts ...interface{}) (*schema.Message, error) {
	return a.agent.Generate(ctx, messages)
}

// Stream 流式生成响应
func (a *EinoReactAgent) Stream(ctx context.Context, messages []*schema.Message, opts ...interface{}) (*schema.StreamReader[*schema.Message], error) {
	return a.agent.Stream(ctx, messages)
}

// Initialize 初始化智能体
func (a *EinoReactAgent) Initialize(ctx context.Context) error {
	// React智能体已经在创建时初始化
	return nil
}

// Shutdown 关闭智能体
func (a *EinoReactAgent) Shutdown(ctx context.Context) error {
	// 清理资源
	return nil
}

// EinoMainAgent 基于Eino的主智能体
type EinoMainAgent struct {
	chatModel  model.BaseChatModel
	config     *EinoAgentConfig
	agentType  EinoAgentType
	reactAgent *EinoReactAgent
}

// NewEinoMainAgent 创建新的Eino主智能体
func NewEinoMainAgent(ctx context.Context, config *EinoAgentConfig) (*EinoMainAgent, error) {
	// 创建React智能体作为执行引擎
	reactAgent, err := NewEinoReactAgent(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create react agent: %w", err)
	}

	return &EinoMainAgent{
		chatModel:  config.ChatModel,
		config:     config,
		agentType:  EinoAgentTypeMain,
		reactAgent: reactAgent,
	}, nil
}

// GetType 获取智能体类型
func (a *EinoMainAgent) GetType() EinoAgentType {
	return a.agentType
}

// Generate 生成响应
func (a *EinoMainAgent) Generate(ctx context.Context, messages []*schema.Message, opts ...interface{}) (*schema.Message, error) {
	// 进行意图识别和查询改写
	rewrittenMessages, err := a.processUserIntent(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("failed to process user intent: %w", err)
	}

	// 使用React智能体执行实际的分析任务
	return a.reactAgent.Generate(ctx, rewrittenMessages, opts...)
}

// Stream 流式生成响应
func (a *EinoMainAgent) Stream(ctx context.Context, messages []*schema.Message, opts ...interface{}) (*schema.StreamReader[*schema.Message], error) {
	// 进行意图识别和查询改写
	rewrittenMessages, err := a.processUserIntent(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("failed to process user intent: %w", err)
	}

	// 使用React智能体执行实际的分析任务
	return a.reactAgent.Stream(ctx, rewrittenMessages, opts...)
}

// Initialize 初始化智能体
func (a *EinoMainAgent) Initialize(ctx context.Context) error {
	return a.reactAgent.Initialize(ctx)
}

// Shutdown 关闭智能体
func (a *EinoMainAgent) Shutdown(ctx context.Context) error {
	return a.reactAgent.Shutdown(ctx)
}

// processUserIntent 处理用户意图识别和查询改写
func (a *EinoMainAgent) processUserIntent(ctx context.Context, messages []*schema.Message) ([]*schema.Message, error) {
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
	config    *EinoAgentConfig
	agentType EinoAgentType
}

// NewEinoAnalysisAgent 创建新的Eino分析智能体
func NewEinoAnalysisAgent(ctx context.Context, config *EinoAgentConfig) (*EinoAnalysisAgent, error) {
	return &EinoAnalysisAgent{
		chatModel: config.ChatModel,
		sandbox:   config.PythonSandbox,
		config:    config,
		agentType: EinoAgentTypeAnalysis,
	}, nil
}

// GetType 获取智能体类型
func (a *EinoAnalysisAgent) GetType() EinoAgentType {
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
