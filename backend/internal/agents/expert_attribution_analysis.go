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

// AttributionAnalysisAgent 归因分析专家智能体
type AttributionAnalysisAgent struct {
	chatModel model.BaseChatModel
	sandbox   *sanbox.PythonSandbox
	config    *types.AgentConfig
	agentType types.AgentType
	mlTool    *tools.MLAnalysisTool
}

// NewAttributionAnalysisAgent 创建归因分析专家智能体
func NewAttributionAnalysisAgent(ctx context.Context, config *types.AgentConfig) (*AttributionAnalysisAgent, error) {
	var mlTool *tools.MLAnalysisTool
	if config.PythonSandbox != nil {
		mlTool = tools.NewMLAnalysisTool(config.PythonSandbox)
	}

	return &AttributionAnalysisAgent{
		chatModel: config.ChatModel,
		sandbox:   config.PythonSandbox,
		config:    config,
		agentType: types.AgentTypeAttributionAnalysis,
		mlTool:    mlTool,
	}, nil
}

// GetType 获取智能体类型
func (a *AttributionAnalysisAgent) GetType() types.AgentType {
	return a.agentType
}

// GetCapabilities 获取能力描述
func (a *AttributionAnalysisAgent) GetCapabilities() []string {
	return []string{
		"因果关系分析",
		"根因分析",
		"贡献度分析",
		"影响因子识别",
		"特征重要性分析",
		"相关性与因果性分析",
		"回归分析",
		"变化归因分析",
	}
}

// CanHandle 判断是否能处理特定任务
func (a *AttributionAnalysisAgent) CanHandle(task *types.Task) bool {
	return task.Type == "attribution_analysis" ||
		task.Type == "root_cause" ||
		task.Type == "causal_analysis" ||
		task.Type == "contribution_analysis" ||
		task.AgentType == types.AgentTypeAttributionAnalysis
}

// Generate 生成响应
func (a *AttributionAnalysisAgent) Generate(ctx context.Context, messages []*schema.Message, opts ...interface{}) (*schema.Message, error) {
	// 生成归因分析代码
	attributionCode, err := a.generateAttributionCode(ctx, messages)
	if err != nil {
		return &schema.Message{
			Role:    schema.Assistant,
			Content: fmt.Sprintf("归因分析代码生成失败: %v", err),
		}, nil
	}

	// 执行归因分析
	result, err := a.executeAttributionAnalysis(attributionCode)
	if err != nil {
		return &schema.Message{
			Role:    schema.Assistant,
			Content: fmt.Sprintf("归因分析执行失败: %v", err),
		}, nil
	}

	return &schema.Message{
		Role:    schema.Assistant,
		Content: result,
	}, nil
}

// Stream 流式生成响应
func (a *AttributionAnalysisAgent) Stream(ctx context.Context, messages []*schema.Message, opts ...interface{}) (*schema.StreamReader[*schema.Message], error) {
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
func (a *AttributionAnalysisAgent) Initialize(ctx context.Context) error {
	return nil
}

// Shutdown 关闭智能体
func (a *AttributionAnalysisAgent) Shutdown(ctx context.Context) error {
	return nil
}

// ExecuteTask 执行任务
func (a *AttributionAnalysisAgent) ExecuteTask(ctx context.Context, task *types.Task) (*types.TaskResult, error) {
	if !a.CanHandle(task) {
		return &types.TaskResult{
			Success:    false,
			Error:      "无法处理此类型的任务",
			ExecutedBy: a.agentType,
		}, nil
	}

	// 解析任务输入
	attributionReq, err := a.parseTaskInput(task.Input)
	if err != nil {
		return &types.TaskResult{
			Success:    false,
			Error:      fmt.Sprintf("解析任务输入失败: %v", err),
			ExecutedBy: a.agentType,
		}, nil
	}

	// 生成归因分析代码
	attributionCode, err := a.generateAttributionCodeFromRequest(ctx, attributionReq)
	if err != nil {
		return &types.TaskResult{
			Success:    false,
			Error:      fmt.Sprintf("生成归因分析代码失败: %v", err),
			ExecutedBy: a.agentType,
		}, nil
	}

	// 执行归因分析
	result, err := a.executeAttributionAnalysis(attributionCode)
	if err != nil {
		return &types.TaskResult{
			Success:    false,
			Error:      fmt.Sprintf("执行归因分析失败: %v", err),
			ExecutedBy: a.agentType,
		}, nil
	}

	return &types.TaskResult{
		Success:    true,
		Output:     result,
		ExecutedBy: a.agentType,
		Metadata: map[string]interface{}{
			"attribution_code": attributionCode,
			"task_type":        task.Type,
		},
	}, nil
}

// parseTaskInput 解析任务输入
func (a *AttributionAnalysisAgent) parseTaskInput(input interface{}) (map[string]interface{}, error) {
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

	// 默认返回空map
	return make(map[string]interface{}), nil
}

// generateAttributionCode 从消息生成归因分析代码
func (a *AttributionAnalysisAgent) generateAttributionCode(ctx context.Context, messages []*schema.Message) (string, error) {
	// 构建归因分析的系统提示
	systemPrompt := `你是一个专业的归因分析和根因分析专家。请根据用户需求生成高质量的Python归因分析代码。

代码要求：
1. 使用pandas, numpy, matplotlib, seaborn, scikit-learn, scipy, statsmodels等专业库
2. 包含完整的归因分析流程：
   - 数据预处理和特征工程
   - 相关性分析
   - 因果关系分析
   - 贡献度计算
   - 影响因子排序
   - 结果可视化
3. 支持多种归因分析方法：
   - 线性回归分析
   - 特征重要性分析（随机森林、XGBoost）
   - 相关性分析（皮尔逊、斯皮尔曼）
   - 主成分分析（PCA）
   - 方差分析（ANOVA）
   - 沙普利值分析（SHAP）
4. 提供量化的贡献度评分
5. 生成归因分析图表和报告
6. 包含统计显著性检验
7. 包含适当的注释和错误处理
8. 保存图表为'attribution_output.png'

请只返回Python代码，不要添加其他解释。`

	attributionMessages := []*schema.Message{
		{
			Role:    schema.System,
			Content: systemPrompt,
		},
	}
	attributionMessages = append(attributionMessages, messages...)

	response, err := a.chatModel.Generate(ctx, attributionMessages)
	if err != nil {
		return "", err
	}

	return a.extractPythonCode(response.Content), nil
}

// generateAttributionCodeFromRequest 从归因分析请求生成代码
func (a *AttributionAnalysisAgent) generateAttributionCodeFromRequest(ctx context.Context, attributionReq map[string]interface{}) (string, error) {
	// 构建归因分析描述
	attributionDesc := a.buildAttributionDescription(attributionReq)

	messages := []*schema.Message{
		{
			Role:    schema.User,
			Content: attributionDesc,
		},
	}

	return a.generateAttributionCode(ctx, messages)
}

// buildAttributionDescription 构建归因分析描述
func (a *AttributionAnalysisAgent) buildAttributionDescription(attributionReq map[string]interface{}) string {
	var desc []string

	// 检查数据源
	if dataSource, ok := attributionReq["data_source"]; ok {
		desc = append(desc, fmt.Sprintf("基于数据源: %v", dataSource))
	}

	// 检查目标变量
	if target, ok := attributionReq["target_variable"]; ok {
		desc = append(desc, fmt.Sprintf("分析目标变量: %v", target))
	}

	// 检查影响因子
	if factors, ok := attributionReq["factors"]; ok {
		desc = append(desc, fmt.Sprintf("分析影响因子: %v", factors))
	}

	// 检查分析方法
	if method, ok := attributionReq["method"]; ok {
		desc = append(desc, fmt.Sprintf("使用 %v 方法", method))
	} else {
		desc = append(desc, "使用多种归因分析方法进行综合分析")
	}

	// 检查分析期间
	if period, ok := attributionReq["analysis_period"]; ok {
		desc = append(desc, fmt.Sprintf("分析时间期间: %v", period))
	}

	// 检查分组维度
	if groupBy, ok := attributionReq["group_by"]; ok {
		desc = append(desc, fmt.Sprintf("按 %v 分组进行归因分析", groupBy))
	}

	// 检查变化事件
	if changeEvent, ok := attributionReq["change_event"]; ok {
		desc = append(desc, fmt.Sprintf("分析变化事件: %v", changeEvent))
	}

	// 检查特殊要求
	if requirements, ok := attributionReq["requirements"]; ok {
		desc = append(desc, fmt.Sprintf("特殊要求: %v", requirements))
	}

	if len(desc) == 0 {
		return "执行综合归因分析，包括相关性分析、因果关系分析和贡献度分析"
	}

	return strings.Join(desc, "\n")
}

// executeAttributionAnalysis 执行归因分析
func (a *AttributionAnalysisAgent) executeAttributionAnalysis(code string) (string, error) {
	if a.sandbox == nil {
		return "", fmt.Errorf("Python沙盒未配置")
	}

	result, err := a.sandbox.ExecutePython(code)
	if err != nil {
		return "", err
	}

	if !result.Success {
		return "", fmt.Errorf("归因分析执行失败: %s", result.Error)
	}

	response := "归因分析完成！\n\n"

	if result.Stdout != "" {
		response += "分析结果:\n" + result.Stdout + "\n\n"
	}

	if result.ImagePath != "" {
		response += "生成的归因分析图表: " + result.ImagePath + "\n"
	}

	return response, nil
}

// extractPythonCode 从响应中提取Python代码
func (a *AttributionAnalysisAgent) extractPythonCode(content string) string {
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
