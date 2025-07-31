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

// AnomalyDetectionAgent 异动分析专家智能体
type AnomalyDetectionAgent struct {
	chatModel model.BaseChatModel
	sandbox   *sanbox.PythonSandbox
	config    *types.AgentConfig
	agentType types.AgentType
	mlTool    *tools.MLAnalysisTool
}

// NewAnomalyDetectionAgent 创建异动检测专家智能体
func NewAnomalyDetectionAgent(ctx context.Context, config *types.AgentConfig) (*AnomalyDetectionAgent, error) {
	var mlTool *tools.MLAnalysisTool
	if config.PythonSandbox != nil {
		mlTool = tools.NewMLAnalysisTool(config.PythonSandbox)
	}

	return &AnomalyDetectionAgent{
		chatModel: config.ChatModel,
		sandbox:   config.PythonSandbox,
		config:    config,
		agentType: types.AgentTypeAnomalyDetection,
		mlTool:    mlTool,
	}, nil
}

// GetType 获取智能体类型
func (a *AnomalyDetectionAgent) GetType() types.AgentType {
	return a.agentType
}

// GetCapabilities 获取能力描述
func (a *AnomalyDetectionAgent) GetCapabilities() []string {
	return []string{
		"异常值检测",
		"离群点分析",
		"时间序列异常检测",
		"统计异常检测",
		"机器学习异常检测",
		"异动根因分析",
		"异常模式识别",
		"实时异常监控",
	}
}

// CanHandle 判断是否能处理特定任务
func (a *AnomalyDetectionAgent) CanHandle(task *types.Task) bool {
	return task.Type == "anomaly_detection" ||
		task.Type == "outlier_detection" ||
		task.Type == "anomaly" ||
		task.AgentType == types.AgentTypeAnomalyDetection
}

// Generate 生成响应
func (a *AnomalyDetectionAgent) Generate(ctx context.Context, messages []*schema.Message, opts ...interface{}) (*schema.Message, error) {
	// 生成异动检测代码
	anomalyCode, err := a.generateAnomalyCode(ctx, messages)
	if err != nil {
		return &schema.Message{
			Role:    schema.Assistant,
			Content: fmt.Sprintf("异动检测代码生成失败: %v", err),
		}, nil
	}

	// 执行异动检测
	result, err := a.executeAnomalyDetection(anomalyCode)
	if err != nil {
		return &schema.Message{
			Role:    schema.Assistant,
			Content: fmt.Sprintf("异动检测执行失败: %v", err),
		}, nil
	}

	return &schema.Message{
		Role:    schema.Assistant,
		Content: result,
	}, nil
}

// Stream 流式生成响应
func (a *AnomalyDetectionAgent) Stream(ctx context.Context, messages []*schema.Message, opts ...interface{}) (*schema.StreamReader[*schema.Message], error) {
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
func (a *AnomalyDetectionAgent) Initialize(ctx context.Context) error {
	return nil
}

// Shutdown 关闭智能体
func (a *AnomalyDetectionAgent) Shutdown(ctx context.Context) error {
	return nil
}

// ExecuteTask 执行任务
func (a *AnomalyDetectionAgent) ExecuteTask(ctx context.Context, task *types.Task) (*types.TaskResult, error) {
	if !a.CanHandle(task) {
		return &types.TaskResult{
			Success:    false,
			Error:      "无法处理此类型的任务",
			ExecutedBy: a.agentType,
		}, nil
	}

	// 解析任务输入
	anomalyReq, err := a.parseTaskInput(task.Input)
	if err != nil {
		return &types.TaskResult{
			Success:    false,
			Error:      fmt.Sprintf("解析任务输入失败: %v", err),
			ExecutedBy: a.agentType,
		}, nil
	}

	// 生成异动检测代码
	anomalyCode, err := a.generateAnomalyCodeFromRequest(ctx, anomalyReq)
	if err != nil {
		return &types.TaskResult{
			Success:    false,
			Error:      fmt.Sprintf("生成异动检测代码失败: %v", err),
			ExecutedBy: a.agentType,
		}, nil
	}

	// 执行异动检测
	result, err := a.executeAnomalyDetection(anomalyCode)
	if err != nil {
		return &types.TaskResult{
			Success:    false,
			Error:      fmt.Sprintf("执行异动检测失败: %v", err),
			ExecutedBy: a.agentType,
		}, nil
	}

	return &types.TaskResult{
		Success:    true,
		Output:     result,
		ExecutedBy: a.agentType,
		Metadata: map[string]interface{}{
			"anomaly_code": anomalyCode,
			"task_type":    task.Type,
		},
	}, nil
}

// parseTaskInput 解析任务输入
func (a *AnomalyDetectionAgent) parseTaskInput(input interface{}) (map[string]interface{}, error) {
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

// generateAnomalyCode 从消息生成异动检测代码
func (a *AnomalyDetectionAgent) generateAnomalyCode(ctx context.Context, messages []*schema.Message) (string, error) {
	// 构建异动检测的系统提示
	systemPrompt := `你是一个专业的异常检测和异动分析专家。请根据用户需求生成高质量的Python异动检测代码。

代码要求：
1. 使用pandas, numpy, matplotlib, seaborn, scikit-learn, scipy等专业库
2. 包含完整的异动检测流程：
   - 数据预处理和清洗
   - 异常值检测和识别
   - 异常模式分析
   - 异动原因分析
   - 结果可视化
3. 支持多种异常检测方法：
   - 统计方法（Z-score、IQR、Grubbs测试）
   - 时间序列异常检测
   - 机器学习方法（Isolation Forest、Local Outlier Factor、One-Class SVM）
   - 基于密度的方法（DBSCAN）
4. 提供异常程度评分
5. 生成异常检测图表和报告
6. 包含异动根因分析
7. 包含适当的注释和错误处理
8. 保存图表为'anomaly_output.png'

请只返回Python代码，不要添加其他解释。`

	anomalyMessages := []*schema.Message{
		{
			Role:    schema.System,
			Content: systemPrompt,
		},
	}
	anomalyMessages = append(anomalyMessages, messages...)

	response, err := a.chatModel.Generate(ctx, anomalyMessages)
	if err != nil {
		return "", err
	}

	return a.extractPythonCode(response.Content), nil
}

// generateAnomalyCodeFromRequest 从异动检测请求生成代码
func (a *AnomalyDetectionAgent) generateAnomalyCodeFromRequest(ctx context.Context, anomalyReq map[string]interface{}) (string, error) {
	// 构建异动检测描述
	anomalyDesc := a.buildAnomalyDescription(anomalyReq)

	messages := []*schema.Message{
		{
			Role:    schema.User,
			Content: anomalyDesc,
		},
	}

	return a.generateAnomalyCode(ctx, messages)
}

// buildAnomalyDescription 构建异动检测描述
func (a *AnomalyDetectionAgent) buildAnomalyDescription(anomalyReq map[string]interface{}) string {
	var desc []string

	// 检查数据源
	if dataSource, ok := anomalyReq["data_source"]; ok {
		desc = append(desc, fmt.Sprintf("基于数据源: %v", dataSource))
	}

	// 检查检测目标
	if target, ok := anomalyReq["target_variables"]; ok {
		desc = append(desc, fmt.Sprintf("检测目标变量: %v", target))
	}

	// 检查检测方法
	if method, ok := anomalyReq["method"]; ok {
		desc = append(desc, fmt.Sprintf("使用 %v 方法", method))
	} else {
		desc = append(desc, "使用多种异常检测方法进行综合分析")
	}

	// 检查时间维度
	if timeField, ok := anomalyReq["time_field"]; ok {
		desc = append(desc, fmt.Sprintf("基于时间字段 %v 进行时序异常检测", timeField))
	}

	// 检查异常阈值
	if threshold, ok := anomalyReq["threshold"]; ok {
		desc = append(desc, fmt.Sprintf("异常阈值: %v", threshold))
	}

	// 检查分组维度
	if groupBy, ok := anomalyReq["group_by"]; ok {
		desc = append(desc, fmt.Sprintf("按 %v 分组进行异常检测", groupBy))
	}

	// 检查特殊要求
	if requirements, ok := anomalyReq["requirements"]; ok {
		desc = append(desc, fmt.Sprintf("特殊要求: %v", requirements))
	}

	if len(desc) == 0 {
		return "执行综合异常检测分析，包括统计异常、时序异常和机器学习异常检测"
	}

	return strings.Join(desc, "\n")
}

// executeAnomalyDetection 执行异动检测
func (a *AnomalyDetectionAgent) executeAnomalyDetection(code string) (string, error) {
	if a.sandbox == nil {
		return "", fmt.Errorf("Python沙盒未配置")
	}

	result, err := a.sandbox.ExecutePython(code)
	if err != nil {
		return "", err
	}

	if !result.Success {
		return "", fmt.Errorf("异动检测执行失败: %s", result.Error)
	}

	response := "异动检测分析完成！\n\n"

	if result.Stdout != "" {
		response += "检测结果:\n" + result.Stdout + "\n\n"
	}

	if result.ImagePath != "" {
		response += "生成的异动分析图表: " + result.ImagePath + "\n"
	}

	return response, nil
}

// extractPythonCode 从响应中提取Python代码
func (a *AnomalyDetectionAgent) extractPythonCode(content string) string {
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
