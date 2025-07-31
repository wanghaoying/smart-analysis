package main

import (
	"context"
	"fmt"
	"log"

	"github.com/cloudwego/eino/schema"
	"smart-analysis/internal/agents"
	"smart-analysis/internal/manager"
	"smart-analysis/internal/types"
	"smart-analysis/internal/utils/llm"
	"smart-analysis/internal/utils/sanbox"
)

// MultiAgentExample 展示Multi-Agent架构使用示例
func MultiAgentExample() {
	ctx := context.Background()

	// 1. 初始化LLM模型
	llmManager := llm.NewLLMManager()
	chatModel, err := llmManager.GetChatModel("openai")
	if err != nil {
		log.Fatalf("初始化LLM模型失败: %v", err)
	}

	// 2. 初始化Python沙盒
	pythonSandbox, err := sanbox.NewPythonSandbox()
	if err != nil {
		log.Fatalf("初始化Python沙盒失败: %v", err)
	}

	// 3. 构建智能体系统
	agentSystem, err := manager.NewAgentSystemBuilder().
		WithChatModel(chatModel).
		WithPythonSandbox(pythonSandbox).
		WithMaxSteps(15).
		WithDebug(true).
		Build(ctx)
	if err != nil {
		log.Fatalf("构建智能体系统失败: %v", err)
	}

	defer agentSystem.Shutdown(ctx)

	// 4. 定义数据模式
	dataSchema := &types.DataSchema{
		TableName: "sales_data",
		Columns: []types.ColumnInfo{
			{Name: "date", Type: "datetime", Description: "销售日期"},
			{Name: "amount", Type: "float", Description: "销售金额"},
			{Name: "region", Type: "string", Description: "销售区域", Values: []string{"北京", "上海", "广州", "深圳"}},
			{Name: "product", Type: "string", Description: "产品类型", Values: []string{"产品A", "产品B", "产品C"}},
			{Name: "customer_type", Type: "string", Description: "客户类型", Values: []string{"VIP", "普通", "新客户"}},
		},
		Constraints: []string{
			"date范围: 2023-01-01 到 2024-12-31",
			"amount > 0",
		},
	}

	// 示例1: 基础数据查询
	fmt.Println("=== 示例1: 基础数据查询 ===")
	response1, err := agentSystem.ProcessQueryWithDataSchema(ctx,
		"查询2024年Q1北京地区产品A的销售数据，按月份统计销售金额",
		dataSchema)
	if err != nil {
		log.Printf("查询失败: %v", err)
	} else {
		fmt.Printf("响应: %s\n\n", response1.Content)
	}

	// 示例2: 趋势分析和预测
	fmt.Println("=== 示例2: 趋势分析和预测 ===")
	response2, err := agentSystem.ProcessQueryWithDataSchema(ctx,
		"分析2024年各地区的销售趋势，识别季节性模式，并预测下个季度的销售情况",
		dataSchema)
	if err != nil {
		log.Printf("分析失败: %v", err)
	} else {
		fmt.Printf("响应: %s\n\n", response2.Content)
	}

	// 示例3: 异动检测
	fmt.Println("=== 示例3: 异动检测 ===")
	response3, err := agentSystem.ProcessQueryWithDataSchema(ctx,
		"检测销售数据中的异常情况，识别可能的异动原因",
		dataSchema)
	if err != nil {
		log.Printf("检测失败: %v", err)
	} else {
		fmt.Printf("响应: %s\n\n", response3.Content)
	}

	// 示例4: 归因分析
	fmt.Println("=== 示例4: 归因分析 ===")
	response4, err := agentSystem.ProcessQueryWithDataSchema(ctx,
		"分析影响销售额的主要因素，计算各因素的贡献度",
		dataSchema)
	if err != nil {
		log.Printf("归因分析失败: %v", err)
	} else {
		fmt.Printf("响应: %s\n\n", response4.Content)
	}

	// 示例5: 流式执行复杂分析
	fmt.Println("=== 示例5: 流式执行复杂分析 ===")
	stream, err := agentSystem.StreamQueryWithDataSchema(ctx,
		"执行综合分析：1)查询最近6个月的销售数据 2)分析销售趋势和季节性 3)检测异常数据点 4)分析关键影响因素 5)预测未来3个月销售情况",
		dataSchema)
	if err != nil {
		log.Printf("流式分析失败: %v", err)
	} else {
		fmt.Println("流式执行进度:")
		for {
			response, err := stream.Recv()
			if err != nil {
				break
			}
			fmt.Printf(">> %s\n", response.Content)
		}
	}
}

// DirectMultiAgentExample 直接使用Multi-Agent管理器的示例
func DirectMultiAgentExample() {
	ctx := context.Background()

	// 1. 初始化组件
	llmManager := llm.NewLLMManager()
	chatModel, _ := llmManager.GetChatModel("openai")
	pythonSandbox, _ := sanbox.NewPythonSandbox()

	// 2. 创建配置
	config := &types.AgentConfig{
		ChatModel:     chatModel,
		PythonSandbox: pythonSandbox,
		MaxSteps:      20,
		EnableDebug:   true,
		Metadata:      make(map[string]interface{}),
	}

	// 3. 直接创建Multi-Agent管理器
	multiAgent, err := agents.NewMultiAgentManager(ctx, config)
	if err != nil {
		log.Fatalf("创建Multi-Agent管理器失败: %v", err)
	}

	defer multiAgent.Shutdown(ctx)

	// 4. 定义数据模式
	dataSchema := &types.DataSchema{
		TableName: "user_behavior",
		Columns: []types.ColumnInfo{
			{Name: "user_id", Type: "int", Description: "用户ID"},
			{Name: "action", Type: "string", Description: "用户行为", Values: []string{"浏览", "购买", "收藏", "分享"}},
			{Name: "timestamp", Type: "datetime", Description: "行为时间"},
			{Name: "page_url", Type: "string", Description: "页面URL"},
			{Name: "duration", Type: "int", Description: "停留时长(秒)"},
		},
	}

	// 5. 构建查询消息
	messages := []*schema.Message{
		{
			Role:    schema.User,
			Content: "分析用户行为模式，识别异常用户，预测用户流失风险",
		},
	}

	// 6. 执行分析
	fmt.Println("=== 直接使用Multi-Agent管理器 ===")
	response, err := multiAgent.Generate(ctx, messages, dataSchema)
	if err != nil {
		log.Printf("分析失败: %v", err)
	} else {
		fmt.Printf("分析结果: %s\n", response.Content)
	}
}

// ExpertAgentExample 展示专家Agent的直接使用
func ExpertAgentExample() {
	ctx := context.Background()

	// 1. 初始化组件
	llmManager := llm.NewLLMManager()
	chatModel, _ := llmManager.GetChatModel("openai")
	pythonSandbox, _ := sanbox.NewPythonSandbox()

	config := &types.AgentConfig{
		ChatModel:     chatModel,
		PythonSandbox: pythonSandbox,
		MaxSteps:      10,
		EnableDebug:   true,
	}

	// 2. 使用工厂创建专家Agent
	factory := agents.NewAgentFactory()

	// 创建数据查询专家
	dataQueryAgent, err := factory.CreateExpertAgent(ctx, types.AgentTypeDataQuery, config)
	if err != nil {
		log.Fatalf("创建数据查询专家失败: %v", err)
	}

	// 创建趋势预测专家
	trendAgent, err := factory.CreateExpertAgent(ctx, types.AgentTypeTrendForecast, config)
	if err != nil {
		log.Fatalf("创建趋势预测专家失败: %v", err)
	}

	// 3. 创建任务
	queryTask := &types.Task{
		ID:          "task_1",
		Type:        "data_query",
		Description: "查询销售数据",
		AgentType:   types.AgentTypeDataQuery,
		Input: map[string]interface{}{
			"events":     []string{"销售"},
			"dimensions": []string{"时间", "地区"},
			"metrics":    []string{"销售额", "销售量"},
			"time_range": map[string]string{
				"start_time": "2024-01-01",
				"end_time":   "2024-03-31",
			},
		},
		Status: types.TaskStatusPending,
	}

	forecastTask := &types.Task{
		ID:          "task_2",
		Type:        "trend_forecast",
		Description: "预测销售趋势",
		AgentType:   types.AgentTypeTrendForecast,
		Input: map[string]interface{}{
			"target_variable":  "销售额",
			"forecast_periods": 12,
			"method":           "auto",
			"seasonal":         true,
		},
		Status: types.TaskStatusPending,
	}

	// 4. 执行任务
	fmt.Println("=== 专家Agent任务执行 ===")

	// 执行数据查询任务
	fmt.Println("执行数据查询任务...")
	queryResult, err := dataQueryAgent.ExecuteTask(ctx, queryTask)
	if err != nil {
		log.Printf("查询任务失败: %v", err)
	} else {
		fmt.Printf("查询结果: %+v\n", queryResult)
	}

	// 执行趋势预测任务
	fmt.Println("执行趋势预测任务...")
	forecastResult, err := trendAgent.ExecuteTask(ctx, forecastTask)
	if err != nil {
		log.Printf("预测任务失败: %v", err)
	} else {
		fmt.Printf("预测结果: %+v\n", forecastResult)
	}

	// 5. 显示专家能力
	fmt.Println("\n=== 专家Agent能力展示 ===")
	expertTypes := factory.GetExpertAgentTypes()
	for _, agentType := range expertTypes {
		capabilities := factory.GetAgentCapabilities(agentType)
		fmt.Printf("%s: %v\n", agentType, capabilities)
	}
}

func main() {
	fmt.Println("Multi-Agent 架构使用示例")
	fmt.Println("============================")

	// 运行不同的示例
	fmt.Println("\n1. 完整Multi-Agent系统示例")
	MultiAgentExample()

	fmt.Println("\n2. 直接Multi-Agent管理器示例")
	DirectMultiAgentExample()

	fmt.Println("\n3. 专家Agent直接使用示例")
	ExpertAgentExample()

	fmt.Println("\n所有示例执行完成！")
}
