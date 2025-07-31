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

// DataQueryAgent 数据查询专家智能体
type DataQueryAgent struct {
	chatModel model.BaseChatModel
	sandbox   *sanbox.PythonSandbox
	config    *types.AgentConfig
	agentType types.AgentType
	tools     *tools.DataQueryTool
}

// NewDataQueryAgent 创建数据查询专家智能体
func NewDataQueryAgent(ctx context.Context, config *types.AgentConfig) (*DataQueryAgent, error) {
	var queryTool *tools.DataQueryTool
	if config.PythonSandbox != nil {
		queryTool = tools.NewDataQueryTool(config.PythonSandbox)
	}

	return &DataQueryAgent{
		chatModel: config.ChatModel,
		sandbox:   config.PythonSandbox,
		config:    config,
		agentType: types.AgentTypeDataQuery,
		tools:     queryTool,
	}, nil
}

// GetType 获取智能体类型
func (a *DataQueryAgent) GetType() types.AgentType {
	return a.agentType
}

// GetCapabilities 获取能力描述
func (a *DataQueryAgent) GetCapabilities() []string {
	return []string{
		"数据查询与筛选",
		"SQL查询生成",
		"数据过滤和聚合",
		"多表关联查询",
		"数据预览和统计",
	}
}

// CanHandle 判断是否能处理特定任务
func (a *DataQueryAgent) CanHandle(task *types.Task) bool {
	return task.Type == "data_query" || task.AgentType == types.AgentTypeDataQuery
}

// Generate 生成响应
func (a *DataQueryAgent) Generate(ctx context.Context, messages []*schema.Message, opts ...interface{}) (*schema.Message, error) {
	// 生成数据查询代码
	queryCode, err := a.generateQueryCode(ctx, messages)
	if err != nil {
		return &schema.Message{
			Role:    schema.Assistant,
			Content: fmt.Sprintf("查询代码生成失败: %v", err),
		}, nil
	}

	// 执行查询
	result, err := a.executeQuery(queryCode)
	if err != nil {
		return &schema.Message{
			Role:    schema.Assistant,
			Content: fmt.Sprintf("查询执行失败: %v", err),
		}, nil
	}

	return &schema.Message{
		Role:    schema.Assistant,
		Content: result,
	}, nil
}

// Stream 流式生成响应
func (a *DataQueryAgent) Stream(ctx context.Context, messages []*schema.Message, opts ...interface{}) (*schema.StreamReader[*schema.Message], error) {
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
func (a *DataQueryAgent) Initialize(ctx context.Context) error {
	return nil
}

// Shutdown 关闭智能体
func (a *DataQueryAgent) Shutdown(ctx context.Context) error {
	return nil
}

// ExecuteTask 执行任务
func (a *DataQueryAgent) ExecuteTask(ctx context.Context, task *types.Task) (*types.TaskResult, error) {
	if !a.CanHandle(task) {
		return &types.TaskResult{
			Success:    false,
			Error:      "无法处理此类型的任务",
			ExecutedBy: a.agentType,
		}, nil
	}

	// 解析任务输入
	queryObject, err := a.parseTaskInput(task.Input)
	if err != nil {
		return &types.TaskResult{
			Success:    false,
			Error:      fmt.Sprintf("解析任务输入失败: %v", err),
			ExecutedBy: a.agentType,
		}, nil
	}

	// 生成查询代码
	queryCode, err := a.generateQueryCodeFromObject(ctx, queryObject)
	if err != nil {
		return &types.TaskResult{
			Success:    false,
			Error:      fmt.Sprintf("生成查询代码失败: %v", err),
			ExecutedBy: a.agentType,
		}, nil
	}

	// 执行查询
	result, err := a.executeQuery(queryCode)
	if err != nil {
		return &types.TaskResult{
			Success:    false,
			Error:      fmt.Sprintf("执行查询失败: %v", err),
			ExecutedBy: a.agentType,
		}, nil
	}

	return &types.TaskResult{
		Success:    true,
		Output:     result,
		ExecutedBy: a.agentType,
		Metadata: map[string]interface{}{
			"query_code": queryCode,
			"task_type":  task.Type,
		},
	}, nil
}

// parseTaskInput 解析任务输入
func (a *DataQueryAgent) parseTaskInput(input interface{}) (*types.QueryObject, error) {
	// 尝试直接转换
	if queryObj, ok := input.(*types.QueryObject); ok {
		return queryObj, nil
	}

	// 尝试从map转换
	if inputMap, ok := input.(map[string]interface{}); ok {
		jsonData, err := json.Marshal(inputMap)
		if err != nil {
			return nil, err
		}

		var queryObj types.QueryObject
		err = json.Unmarshal(jsonData, &queryObj)
		if err != nil {
			return nil, err
		}

		return &queryObj, nil
	}

	return nil, fmt.Errorf("无法解析任务输入")
}

// generateQueryCode 从消息生成查询代码
func (a *DataQueryAgent) generateQueryCode(ctx context.Context, messages []*schema.Message) (string, error) {
	// 构建查询生成的系统提示
	systemPrompt := `你是一个专业的数据查询专家。请根据用户需求生成高质量的Python数据查询代码。

代码要求：
1. 使用pandas进行数据操作
2. 支持数据筛选、聚合、排序等操作
3. 处理各种数据类型和格式
4. 包含错误处理
5. 输出清晰的查询结果

请只返回Python代码，不要添加其他解释。代码应该可以直接执行。`

	queryMessages := []*schema.Message{
		{
			Role:    schema.System,
			Content: systemPrompt,
		},
	}
	queryMessages = append(queryMessages, messages...)

	response, err := a.chatModel.Generate(ctx, queryMessages)
	if err != nil {
		return "", err
	}

	return a.extractPythonCode(response.Content), nil
}

// generateQueryCodeFromObject 从查询对象生成查询代码
func (a *DataQueryAgent) generateQueryCodeFromObject(ctx context.Context, queryObj *types.QueryObject) (string, error) {
	// 构建查询描述
	queryDesc := a.buildQueryDescription(queryObj)

	messages := []*schema.Message{
		{
			Role:    schema.User,
			Content: queryDesc,
		},
	}

	return a.generateQueryCode(ctx, messages)
}

// buildQueryDescription 构建查询描述
func (a *DataQueryAgent) buildQueryDescription(queryObj *types.QueryObject) string {
	var desc []string

	if len(queryObj.Events) > 0 {
		desc = append(desc, fmt.Sprintf("查询事件: %v", queryObj.Events))
	}

	if len(queryObj.Dimensions) > 0 {
		desc = append(desc, fmt.Sprintf("按维度分组: %v", queryObj.Dimensions))
	}

	if len(queryObj.Metrics) > 0 {
		desc = append(desc, fmt.Sprintf("计算指标: %v", queryObj.Metrics))
	}

	if len(queryObj.Filters) > 0 {
		for _, filter := range queryObj.Filters {
			desc = append(desc, fmt.Sprintf("过滤条件: %s %s %v", filter.Column, filter.Operator, filter.Value))
		}
	}

	if queryObj.TimeRange != nil {
		desc = append(desc, fmt.Sprintf("时间范围: %s 到 %s", queryObj.TimeRange.StartTime, queryObj.TimeRange.EndTime))
	}

	if len(queryObj.GroupBy) > 0 {
		desc = append(desc, fmt.Sprintf("分组字段: %v", queryObj.GroupBy))
	}

	if len(queryObj.OrderBy) > 0 {
		for _, order := range queryObj.OrderBy {
			desc = append(desc, fmt.Sprintf("排序: %s %s", order.Column, order.Direction))
		}
	}

	if queryObj.Limit > 0 {
		desc = append(desc, fmt.Sprintf("限制数量: %d", queryObj.Limit))
	}

	if len(desc) == 0 {
		return "执行基本数据查询"
	}

	return "根据以下要求查询数据：\n" + strings.Join(desc, "\n")
}

// executeQuery 执行查询
func (a *DataQueryAgent) executeQuery(code string) (string, error) {
	if a.sandbox == nil {
		return "", fmt.Errorf("Python沙盒未配置")
	}

	result, err := a.sandbox.ExecutePython(code)
	if err != nil {
		return "", err
	}

	if !result.Success {
		return "", fmt.Errorf("查询执行失败: %s", result.Error)
	}

	response := "数据查询完成！\n\n"
	if result.Stdout != "" {
		response += "查询结果:\n" + result.Stdout + "\n"
	}

	return response, nil
}

// extractPythonCode 从响应中提取Python代码
func (a *DataQueryAgent) extractPythonCode(content string) string {
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
