package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"smart-analysis/internal/tools"
	"smart-analysis/internal/types"
	"smart-analysis/internal/utils/sanbox"
)

// DataAnalysisAgent 数据分析专家智能体
type DataAnalysisAgent struct {
	chatModel      model.BaseChatModel
	sandbox        *sanbox.PythonSandbox
	config         *types.AgentConfig
	agentType      types.AgentType
	pythonTool     *tools.PythonAnalysisTool
	vizTool        *tools.EChartsVisualizationTool
	preprocessTool *tools.DataPreprocessingTool
}

// NewDataAnalysisAgent 创建数据分析专家智能体
func NewDataAnalysisAgent(ctx context.Context, config *types.AgentConfig) (*DataAnalysisAgent, error) {
	var pythonTool *tools.PythonAnalysisTool
	var vizTool *tools.EChartsVisualizationTool
	var preprocessTool *tools.DataPreprocessingTool

	if config.PythonSandbox != nil {
		pythonTool = tools.NewPythonAnalysisTool(config.PythonSandbox)
		vizTool = tools.NewEChartsVisualizationTool(config.PythonSandbox)
		preprocessTool = tools.NewDataPreprocessingTool(config.PythonSandbox)
	}

	return &DataAnalysisAgent{
		chatModel:      config.ChatModel,
		sandbox:        config.PythonSandbox,
		config:         config,
		agentType:      types.AgentTypeDataAnalysis,
		pythonTool:     pythonTool,
		vizTool:        vizTool,
		preprocessTool: preprocessTool,
	}, nil
}

// GetType 获取智能体类型
func (a *DataAnalysisAgent) GetType() types.AgentType {
	return a.agentType
}

// GetCapabilities 获取能力描述
func (a *DataAnalysisAgent) GetCapabilities() []string {
	return []string{
		"描述性统计分析",
		"相关性分析",
		"分布分析",
		"对比分析",
		"数据可视化",
		"数据预处理",
		"特征工程",
	}
}

// CanHandle 判断是否能处理特定任务
func (a *DataAnalysisAgent) CanHandle(task *types.Task) bool {
	return task.Type == "analysis" ||
		task.Type == "data_analysis" ||
		task.Type == "visualization" ||
		task.AgentType == types.AgentTypeDataAnalysis
}

// Generate 生成响应
func (a *DataAnalysisAgent) Generate(ctx context.Context, messages []*schema.Message, opts ...interface{}) (*schema.Message, error) {
	// 生成分析代码
	analysisCode, err := a.generateAnalysisCode(ctx, messages)
	if err != nil {
		return &schema.Message{
			Role:    schema.Assistant,
			Content: fmt.Sprintf("分析代码生成失败: %v", err),
		}, nil
	}

	// 执行分析
	result, err := a.executeAnalysis(analysisCode)
	if err != nil {
		return &schema.Message{
			Role:    schema.Assistant,
			Content: fmt.Sprintf("分析执行失败: %v", err),
		}, nil
	}

	return &schema.Message{
		Role:    schema.Assistant,
		Content: result,
	}, nil
}

// Stream 流式生成响应
func (a *DataAnalysisAgent) Stream(ctx context.Context, messages []*schema.Message, opts ...interface{}) (*schema.StreamReader[*schema.Message], error) {
	response, err := a.Generate(ctx, messages, opts...)
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

// Initialize 初始化智能体
func (a *DataAnalysisAgent) Initialize(ctx context.Context) error {
	return nil
}

// Shutdown 关闭智能体
func (a *DataAnalysisAgent) Shutdown(ctx context.Context) error {
	return nil
}

// ExecuteTask 执行任务
func (a *DataAnalysisAgent) ExecuteTask(ctx context.Context, task *types.Task) (*types.TaskResult, error) {
	if !a.CanHandle(task) {
		return &types.TaskResult{
			Success:    false,
			Error:      "无法处理此类型的任务",
			ExecutedBy: a.agentType,
		}, nil
	}

	// 解析任务输入
	analysisReq, err := a.parseTaskInput(task.Input)
	if err != nil {
		return &types.TaskResult{
			Success:    false,
			Error:      fmt.Sprintf("解析任务输入失败: %v", err),
			ExecutedBy: a.agentType,
		}, nil
	}

	// 生成分析代码
	analysisCode, err := a.generateAnalysisCodeFromRequest(ctx, analysisReq)
	if err != nil {
		return &types.TaskResult{
			Success:    false,
			Error:      fmt.Sprintf("生成分析代码失败: %v", err),
			ExecutedBy: a.agentType,
		}, nil
	}

	// 执行分析
	result, err := a.executeAnalysis(analysisCode)
	if err != nil {
		return &types.TaskResult{
			Success:    false,
			Error:      fmt.Sprintf("执行分析失败: %v", err),
			ExecutedBy: a.agentType,
		}, nil
	}

	return &types.TaskResult{
		Success:    true,
		Output:     result,
		ExecutedBy: a.agentType,
		Metadata: map[string]interface{}{
			"analysis_code": analysisCode,
			"task_type":     task.Type,
		},
	}, nil
}

// parseTaskInput 解析任务输入
func (a *DataAnalysisAgent) parseTaskInput(input interface{}) (map[string]interface{}, error) {
	// 尝试直接转换
	if inputMap, ok := input.(map[string]interface{}); ok {
		return inputMap, nil
	}

	// 尝试从JSON字符串转换
	if inputStr, ok := input.(string); ok {
		var inputMap map[string]interface{}
		err := json.Unmarshal([]byte(inputStr), &inputMap)
		if err == nil {
			return inputMap, nil
		}
	}

	// 默认返回空map，表示通用分析
	return make(map[string]interface{}), nil
}

// generateAnalysisCode 从消息生成分析代码
func (a *DataAnalysisAgent) generateAnalysisCode(ctx context.Context, messages []*schema.Message) (string, error) {
	// 构建分析生成的系统提示
	systemPrompt := `你是一个专业的数据分析专家。请根据用户需求生成高质量的Python数据分析代码。

代码要求：
1. 使用pandas, numpy, matplotlib, seaborn等专业库
2. 包含完整的数据分析流程：数据加载、清洗、分析、可视化
3. 提供统计分析结果和洞察
4. 生成清晰的图表和报告
5. 包含适当的注释和错误处理
6. 保存图表为'analysis_output.png'

分析类型包括但不限于：
- 描述性统计
- 相关性分析
- 分布分析
- 趋势分析
- 对比分析
- 聚类分析

请只返回Python代码，不要添加其他解释。`

	analysisMessages := []*schema.Message{
		{
			Role:    schema.System,
			Content: systemPrompt,
		},
	}
	analysisMessages = append(analysisMessages, messages...)

	response, err := a.chatModel.Generate(ctx, analysisMessages)
	if err != nil {
		return "", err
	}

	return a.extractPythonCode(response.Content), nil
}

// generateAnalysisCodeFromRequest 从分析请求生成代码
func (a *DataAnalysisAgent) generateAnalysisCodeFromRequest(ctx context.Context, analysisReq map[string]interface{}) (string, error) {
	// 构建分析描述
	analysisDesc := a.buildAnalysisDescription(analysisReq)

	messages := []*schema.Message{
		{
			Role:    schema.User,
			Content: analysisDesc,
		},
	}

	return a.generateAnalysisCode(ctx, messages)
}

// buildAnalysisDescription 构建分析描述
func (a *DataAnalysisAgent) buildAnalysisDescription(analysisReq map[string]interface{}) string {
	var desc []string

	// 检查数据源
	if dataSource, ok := analysisReq["data_source"]; ok {
		desc = append(desc, fmt.Sprintf("基于数据源: %v", dataSource))
	}

	// 检查分析类型
	if analysisType, ok := analysisReq["analysis_type"]; ok {
		desc = append(desc, fmt.Sprintf("执行%s分析", analysisType))
	} else {
		desc = append(desc, "执行综合数据分析")
	}

	// 检查分析目标
	if target, ok := analysisReq["target"]; ok {
		desc = append(desc, fmt.Sprintf("分析目标: %v", target))
	}

	// 检查关注字段
	if fields, ok := analysisReq["fields"]; ok {
		desc = append(desc, fmt.Sprintf("关注字段: %v", fields))
	}

	// 检查特殊要求
	if requirements, ok := analysisReq["requirements"]; ok {
		desc = append(desc, fmt.Sprintf("特殊要求: %v", requirements))
	}

	if len(desc) == 0 {
		return "执行基本数据分析，包括描述性统计、分布分析和相关性分析"
	}

	return strings.Join(desc, "\n")
}

// executeAnalysis 执行分析
func (a *DataAnalysisAgent) executeAnalysis(code string) (string, error) {
	if a.sandbox == nil {
		return "", fmt.Errorf("Python沙盒未配置")
	}

	result, err := a.sandbox.ExecutePython(code)
	if err != nil {
		return "", err
	}

	if !result.Success {
		return "", fmt.Errorf("分析执行失败: %s", result.Error)
	}

	response := "数据分析完成！\n\n"

	if result.Stdout != "" {
		response += "分析结果:\n" + result.Stdout + "\n\n"
	}

	if result.ImagePath != "" {
		response += "生成的图表: " + result.ImagePath + "\n"
	}

	return response, nil
}

// extractPythonCode 从响应中提取Python代码
func (a *DataAnalysisAgent) extractPythonCode(content string) string {
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
