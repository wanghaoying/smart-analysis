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

// TrendForecastAgent 趋势分析与预测专家智能体
type TrendForecastAgent struct {
	chatModel model.BaseChatModel
	sandbox   *sanbox.PythonSandbox
	config    *types.AgentConfig
	agentType types.AgentType
	mlTool    *tools.MLAnalysisTool
}

// NewTrendForecastAgent 创建趋势预测专家智能体
func NewTrendForecastAgent(ctx context.Context, config *types.AgentConfig) (*TrendForecastAgent, error) {
	var mlTool *tools.MLAnalysisTool
	if config.PythonSandbox != nil {
		mlTool = tools.NewMLAnalysisTool(config.PythonSandbox)
	}

	return &TrendForecastAgent{
		chatModel: config.ChatModel,
		sandbox:   config.PythonSandbox,
		config:    config,
		agentType: types.AgentTypeTrendForecast,
		mlTool:    mlTool,
	}, nil
}

// GetType 获取智能体类型
func (a *TrendForecastAgent) GetType() types.AgentType {
	return a.agentType
}

// GetCapabilities 获取能力描述
func (a *TrendForecastAgent) GetCapabilities() []string {
	return []string{
		"时间序列分析",
		"趋势预测",
		"季节性分析",
		"周期性检测",
		"回归分析",
		"ARIMA建模",
		"指数平滑",
		"机器学习预测",
	}
}

// CanHandle 判断是否能处理特定任务
func (a *TrendForecastAgent) CanHandle(task *types.Task) bool {
	return task.Type == "trend_forecast" ||
		task.Type == "time_series" ||
		task.Type == "forecast" ||
		task.AgentType == types.AgentTypeTrendForecast
}

// Generate 生成响应
func (a *TrendForecastAgent) Generate(ctx context.Context, messages []*schema.Message, opts ...interface{}) (*schema.Message, error) {
	// 生成趋势分析代码
	forecastCode, err := a.generateForecastCode(ctx, messages)
	if err != nil {
		return &schema.Message{
			Role:    schema.Assistant,
			Content: fmt.Sprintf("趋势分析代码生成失败: %v", err),
		}, nil
	}

	// 执行趋势分析
	result, err := a.executeForecast(forecastCode)
	if err != nil {
		return &schema.Message{
			Role:    schema.Assistant,
			Content: fmt.Sprintf("趋势分析执行失败: %v", err),
		}, nil
	}

	return &schema.Message{
		Role:    schema.Assistant,
		Content: result,
	}, nil
}

// Stream 流式生成响应
func (a *TrendForecastAgent) Stream(ctx context.Context, messages []*schema.Message, opts ...interface{}) (*schema.StreamReader[*schema.Message], error) {
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
func (a *TrendForecastAgent) Initialize(ctx context.Context) error {
	return nil
}

// Shutdown 关闭智能体
func (a *TrendForecastAgent) Shutdown(ctx context.Context) error {
	return nil
}

// ExecuteTask 执行任务
func (a *TrendForecastAgent) ExecuteTask(ctx context.Context, task *types.Task) (*types.TaskResult, error) {
	if !a.CanHandle(task) {
		return &types.TaskResult{
			Success:    false,
			Error:      "无法处理此类型的任务",
			ExecutedBy: a.agentType,
		}, nil
	}

	// 解析任务输入
	forecastReq, err := a.parseTaskInput(task.Input)
	if err != nil {
		return &types.TaskResult{
			Success:    false,
			Error:      fmt.Sprintf("解析任务输入失败: %v", err),
			ExecutedBy: a.agentType,
		}, nil
	}

	// 生成趋势分析代码
	forecastCode, err := a.generateForecastCodeFromRequest(ctx, forecastReq)
	if err != nil {
		return &types.TaskResult{
			Success:    false,
			Error:      fmt.Sprintf("生成趋势分析代码失败: %v", err),
			ExecutedBy: a.agentType,
		}, nil
	}

	// 执行趋势分析
	result, err := a.executeForecast(forecastCode)
	if err != nil {
		return &types.TaskResult{
			Success:    false,
			Error:      fmt.Sprintf("执行趋势分析失败: %v", err),
			ExecutedBy: a.agentType,
		}, nil
	}

	return &types.TaskResult{
		Success:    true,
		Output:     result,
		ExecutedBy: a.agentType,
		Metadata: map[string]interface{}{
			"forecast_code": forecastCode,
			"task_type":     task.Type,
		},
	}, nil
}

// parseTaskInput 解析任务输入
func (a *TrendForecastAgent) parseTaskInput(input interface{}) (map[string]interface{}, error) {
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

// generateForecastCode 从消息生成趋势分析代码
func (a *TrendForecastAgent) generateForecastCode(ctx context.Context, messages []*schema.Message) (string, error) {
	// 构建趋势分析的系统提示
	systemPrompt := `你是一个专业的时间序列分析和趋势预测专家。请根据用户需求生成高质量的Python趋势分析代码。

代码要求：
1. 使用pandas, numpy, matplotlib, seaborn, scikit-learn, statsmodels等专业库
2. 包含完整的时间序列分析流程：
   - 数据预处理和清洗
   - 趋势和季节性分析
   - 平稳性检验
   - 模型拟合和预测
   - 预测结果可视化
3. 支持多种预测方法：
   - 线性回归
   - ARIMA模型
   - 指数平滑
   - 机器学习方法
4. 提供预测精度评估
5. 生成趋势图表和预测结果
6. 包含适当的注释和错误处理
7. 保存图表为'forecast_output.png'

请只返回Python代码，不要添加其他解释。`

	forecastMessages := []*schema.Message{
		{
			Role:    schema.System,
			Content: systemPrompt,
		},
	}
	forecastMessages = append(forecastMessages, messages...)

	response, err := a.chatModel.Generate(ctx, forecastMessages)
	if err != nil {
		return "", err
	}

	return a.extractPythonCode(response.Content), nil
}

// generateForecastCodeFromRequest 从预测请求生成代码
func (a *TrendForecastAgent) generateForecastCodeFromRequest(ctx context.Context, forecastReq map[string]interface{}) (string, error) {
	// 构建预测描述
	forecastDesc := a.buildForecastDescription(forecastReq)

	messages := []*schema.Message{
		{
			Role:    schema.User,
			Content: forecastDesc,
		},
	}

	return a.generateForecastCode(ctx, messages)
}

// buildForecastDescription 构建预测描述
func (a *TrendForecastAgent) buildForecastDescription(forecastReq map[string]interface{}) string {
	var desc []string

	// 检查数据源
	if dataSource, ok := forecastReq["data_source"]; ok {
		desc = append(desc, fmt.Sprintf("基于数据源: %v", dataSource))
	}

	// 检查预测目标
	if target, ok := forecastReq["target_variable"]; ok {
		desc = append(desc, fmt.Sprintf("预测目标变量: %v", target))
	}

	// 检查时间字段
	if timeField, ok := forecastReq["time_field"]; ok {
		desc = append(desc, fmt.Sprintf("时间字段: %v", timeField))
	}

	// 检查预测期数
	if periods, ok := forecastReq["forecast_periods"]; ok {
		desc = append(desc, fmt.Sprintf("预测未来 %v 个时间点", periods))
	} else {
		desc = append(desc, "预测未来趋势")
	}

	// 检查预测方法
	if method, ok := forecastReq["method"]; ok {
		desc = append(desc, fmt.Sprintf("使用 %v 方法", method))
	} else {
		desc = append(desc, "使用自动选择最佳预测方法")
	}

	// 检查季节性
	if seasonal, ok := forecastReq["seasonal"]; ok && seasonal.(bool) {
		desc = append(desc, "考虑季节性因素")
	}

	// 检查特殊要求
	if requirements, ok := forecastReq["requirements"]; ok {
		desc = append(desc, fmt.Sprintf("特殊要求: %v", requirements))
	}

	if len(desc) == 0 {
		return "执行时间序列趋势分析和预测，包括趋势检测、季节性分析和未来值预测"
	}

	return strings.Join(desc, "\n")
}

// executeForecast 执行趋势分析
func (a *TrendForecastAgent) executeForecast(code string) (string, error) {
	if a.sandbox == nil {
		return "", fmt.Errorf("Python沙盒未配置")
	}

	result, err := a.sandbox.ExecutePython(code)
	if err != nil {
		return "", err
	}

	if !result.Success {
		return "", fmt.Errorf("趋势分析执行失败: %s", result.Error)
	}

	response := "趋势分析和预测完成！\n\n"

	if result.Stdout != "" {
		response += "分析结果:\n" + result.Stdout + "\n\n"
	}

	if result.ImagePath != "" {
		response += "生成的趋势图表: " + result.ImagePath + "\n"
	}

	return response, nil
}

// extractPythonCode 从响应中提取Python代码
func (a *TrendForecastAgent) extractPythonCode(content string) string {
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
