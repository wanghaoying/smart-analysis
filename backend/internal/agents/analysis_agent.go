package agents

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"smart-analysis/internal/types"
	"smart-analysis/internal/utils/sanbox"
)

// AnalysisAgent 数据分析智能体（简化版，主要用于特定分析任务）
type AnalysisAgent struct {
	chatModel model.BaseChatModel
	sandbox   *sanbox.PythonSandbox
	config    *types.AgentConfig
	agentType types.AgentType
}

// NewAnalysisAgent 创建新的分析智能体
func NewAnalysisAgent(ctx context.Context, config *types.AgentConfig) (*AnalysisAgent, error) {
	return &AnalysisAgent{
		chatModel: config.ChatModel,
		sandbox:   config.PythonSandbox,
		config:    config,
		agentType: types.AgentTypeAnalysis,
	}, nil
}

// GetType 获取智能体类型
func (a *AnalysisAgent) GetType() types.AgentType {
	return a.agentType
}

// Generate 生成响应
func (a *AnalysisAgent) Generate(ctx context.Context, messages []*schema.Message, opts ...interface{}) (*schema.Message, error) {
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
func (a *AnalysisAgent) Stream(ctx context.Context, messages []*schema.Message, opts ...interface{}) (*schema.StreamReader[*schema.Message], error) {
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
func (a *AnalysisAgent) Initialize(ctx context.Context) error {
	return nil
}

// Shutdown 关闭智能体
func (a *AnalysisAgent) Shutdown(ctx context.Context) error {
	return nil
}

// generateAnalysisCode 生成分析代码
func (a *AnalysisAgent) generateAnalysisCode(ctx context.Context, messages []*schema.Message) (string, error) {
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
func (a *AnalysisAgent) extractPythonCode(content string) string {
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
