package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"smart-analysis/internal/types"
)

// MasterAgent 主控智能体，负责意图识别和会话重写
type MasterAgent struct {
	chatModel model.BaseChatModel
	config    *types.AgentConfig
	agentType types.AgentType
}

// NewMasterAgent 创建新的主控智能体
func NewMasterAgent(ctx context.Context, config *types.AgentConfig) (*MasterAgent, error) {
	return &MasterAgent{
		chatModel: config.ChatModel,
		config:    config,
		agentType: types.AgentTypeMaster,
	}, nil
}

// GetType 获取智能体类型
func (a *MasterAgent) GetType() types.AgentType {
	return a.agentType
}

// Generate 生成响应 - 主要用于意图识别和查询重写
func (a *MasterAgent) Generate(ctx context.Context, messages []*schema.Message, opts ...interface{}) (*schema.Message, error) {
	// 解析数据模式（如果提供）
	var dataSchema *types.DataSchema
	if len(opts) > 0 {
		if schema, ok := opts[0].(*types.DataSchema); ok {
			dataSchema = schema
		}
	}

	// 进行意图识别和查询重写
	queryIntent, err := a.identifyIntent(ctx, messages, dataSchema)
	if err != nil {
		return &schema.Message{
			Role:    schema.Assistant,
			Content: fmt.Sprintf("意图识别失败: %v", err),
		}, nil
	}

	// 将结果序列化为JSON返回
	intentJSON, err := json.Marshal(queryIntent)
	if err != nil {
		return &schema.Message{
			Role:    schema.Assistant,
			Content: fmt.Sprintf("结果序列化失败: %v", err),
		}, nil
	}

	return &schema.Message{
		Role:    schema.Assistant,
		Content: string(intentJSON),
	}, nil
}

// Stream 流式生成响应
func (a *MasterAgent) Stream(ctx context.Context, messages []*schema.Message, opts ...interface{}) (*schema.StreamReader[*schema.Message], error) {
	// 对于MasterAgent，直接生成完整响应然后流式返回
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
func (a *MasterAgent) Initialize(ctx context.Context) error {
	return nil
}

// Shutdown 关闭智能体
func (a *MasterAgent) Shutdown(ctx context.Context) error {
	return nil
}

// identifyIntent 意图识别和查询重写
func (a *MasterAgent) identifyIntent(ctx context.Context, messages []*schema.Message, dataSchema *types.DataSchema) (*types.QueryIntent, error) {
	if len(messages) == 0 {
		return nil, fmt.Errorf("没有输入消息")
	}

	// 获取用户的最后一条消息
	var userQuery string
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == schema.User {
			userQuery = messages[i].Content
			break
		}
	}

	if userQuery == "" {
		return nil, fmt.Errorf("未找到用户查询")
	}

	// 构建数据模式信息
	var schemaInfo string
	if dataSchema != nil {
		schemaJSON, _ := json.MarshalIndent(dataSchema, "", "  ")
		schemaInfo = string(schemaJSON)
	} else {
		schemaInfo = "数据模式未提供"
	}

	// 构建意图识别的系统提示
	systemPrompt := fmt.Sprintf(`你是一个专业的数据分析意图识别专家。请分析用户的查询，基于数据模式信息，识别其数据查询或分析诉求中的重点信息。

数据模式信息:
%s

请分析用户查询并提取以下信息：
1. 意图类型：确定用户是想要数据查询、数据分析、可视化、趋势预测、异动检测还是归因分析
2. 事件(Events)：用户关心的业务事件或指标
3. 维度(Dimensions)：用户想要分析的维度，如时间、地区、产品类型等
4. 度量(Metrics)：用户关心的度量指标，如数量、金额、比率等
5. 过滤条件(Filters)：用户提到的筛选条件
6. 时间范围(TimeRange)：用户指定的时间范围
7. 分组和排序要求
8. 其他特殊要求

请返回标准的JSON格式结果，包含以下字段：
- intent_type: 意图类型 (data_query/analysis/visualization/trend_forecast/anomaly_detection/attribution_analysis)
- query_object: 包含events, dimensions, metrics, filters, time_range, group_by, order_by等
- requirements: 用户的具体要求列表

用户查询: %s`, schemaInfo, userQuery)

	// 使用LLM进行意图识别
	intentMessages := []*schema.Message{
		{
			Role:    schema.System,
			Content: systemPrompt,
		},
		{
			Role:    schema.User,
			Content: "请分析并提取这个查询的意图和结构化信息。",
		},
	}

	response, err := a.chatModel.Generate(ctx, intentMessages)
	if err != nil {
		return nil, fmt.Errorf("调用LLM失败: %w", err)
	}

	// 解析LLM返回的JSON结果
	queryIntent, err := a.parseIntentResponse(response.Content)
	if err != nil {
		// 如果解析失败，创建一个基本的意图对象
		return &types.QueryIntent{
			IntentType: "analysis",
			DataSchema: dataSchema,
			QueryObject: &types.QueryObject{
				Events: []string{userQuery},
			},
			Requirements: []string{userQuery},
		}, nil
	}

	// 设置数据模式
	queryIntent.DataSchema = dataSchema

	return queryIntent, nil
}

// parseIntentResponse 解析意图识别响应
func (a *MasterAgent) parseIntentResponse(content string) (*types.QueryIntent, error) {
	// 尝试提取JSON内容
	jsonStr := a.extractJSON(content)
	if jsonStr == "" {
		return nil, fmt.Errorf("未找到有效的JSON内容")
	}

	var intent types.QueryIntent
	err := json.Unmarshal([]byte(jsonStr), &intent)
	if err != nil {
		return nil, fmt.Errorf("JSON解析失败: %w", err)
	}

	return &intent, nil
}

// extractJSON 从文本中提取JSON内容
func (a *MasterAgent) extractJSON(content string) string {
	// 查找JSON代码块
	start := strings.Index(content, "```json")
	if start != -1 {
		start += len("```json")
		end := strings.Index(content[start:], "```")
		if end != -1 {
			return strings.TrimSpace(content[start : start+end])
		}
	}

	// 查找普通代码块
	start = strings.Index(content, "```")
	if start != -1 {
		start += 3
		end := strings.Index(content[start:], "```")
		if end != -1 {
			return strings.TrimSpace(content[start : start+end])
		}
	}

	// 尝试查找JSON对象
	start = strings.Index(content, "{")
	if start != -1 {
		braceCount := 0
		for i := start; i < len(content); i++ {
			if content[i] == '{' {
				braceCount++
			} else if content[i] == '}' {
				braceCount--
				if braceCount == 0 {
					return strings.TrimSpace(content[start : i+1])
				}
			}
		}
	}

	return ""
}
