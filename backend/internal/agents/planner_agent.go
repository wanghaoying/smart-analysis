package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"
	"smart-analysis/internal/types"
)

// PlannerAgent 任务规划和调度智能体
type PlannerAgent struct {
	chatModel    model.BaseChatModel
	config       *types.AgentConfig
	agentType    types.AgentType
	expertAgents map[types.AgentType]types.ExpertAgent
	agentMutex   sync.RWMutex
}

// NewPlannerAgent 创建新的规划智能体
func NewPlannerAgent(ctx context.Context, config *types.AgentConfig) (*PlannerAgent, error) {
	return &PlannerAgent{
		chatModel:    config.ChatModel,
		config:       config,
		agentType:    types.AgentTypePlanner,
		expertAgents: make(map[types.AgentType]types.ExpertAgent),
	}, nil
}

// GetType 获取智能体类型
func (a *PlannerAgent) GetType() types.AgentType {
	return a.agentType
}

// RegisterExpertAgent 注册专家智能体
func (a *PlannerAgent) RegisterExpertAgent(agent types.ExpertAgent) {
	a.agentMutex.Lock()
	defer a.agentMutex.Unlock()
	a.expertAgents[agent.GetType()] = agent
}

// Generate 生成响应 - 创建并执行执行计划
func (a *PlannerAgent) Generate(ctx context.Context, messages []*schema.Message, opts ...interface{}) (*schema.Message, error) {
	// 解析查询意图
	var queryIntent *types.QueryIntent
	if len(opts) > 0 {
		if intent, ok := opts[0].(*types.QueryIntent); ok {
			queryIntent = intent
		}
	}

	if queryIntent == nil {
		return &schema.Message{
			Role:    schema.Assistant,
			Content: "未提供查询意图信息",
		}, nil
	}

	// 创建执行计划
	plan, err := a.createExecutionPlan(ctx, queryIntent)
	if err != nil {
		return &schema.Message{
			Role:    schema.Assistant,
			Content: fmt.Sprintf("创建执行计划失败: %v", err),
		}, nil
	}

	// 执行计划
	results, err := a.executePlan(ctx, plan)
	if err != nil {
		return &schema.Message{
			Role:    schema.Assistant,
			Content: fmt.Sprintf("执行计划失败: %v", err),
		}, nil
	}

	// 整合结果
	finalResponse := a.consolidateResults(results)

	return &schema.Message{
		Role:    schema.Assistant,
		Content: finalResponse,
	}, nil
}

// Stream 流式生成响应
func (a *PlannerAgent) Stream(ctx context.Context, messages []*schema.Message, opts ...interface{}) (*schema.StreamReader[*schema.Message], error) {
	// 对于PlannerAgent，可以实现流式返回任务执行进度
	sr, sw := schema.Pipe[*schema.Message](10)

	go func() {
		defer sw.Close()

		// 解析查询意图
		var queryIntent *types.QueryIntent
		if len(opts) > 0 {
			if intent, ok := opts[0].(*types.QueryIntent); ok {
				queryIntent = intent
			}
		}

		if queryIntent == nil {
			sw.Send(&schema.Message{
				Role:    schema.Assistant,
				Content: "未提供查询意图信息",
			}, nil)
			return
		}

		// 发送开始消息
		sw.Send(&schema.Message{
			Role:    schema.Assistant,
			Content: "开始创建执行计划...",
		}, nil)

		// 创建执行计划
		plan, err := a.createExecutionPlan(ctx, queryIntent)
		if err != nil {
			sw.Send(&schema.Message{
				Role:    schema.Assistant,
				Content: fmt.Sprintf("创建执行计划失败: %v", err),
			}, nil)
			return
		}

		sw.Send(&schema.Message{
			Role:    schema.Assistant,
			Content: fmt.Sprintf("创建了包含 %d 个任务的执行计划", len(plan.Tasks)),
		}, nil)

		// 流式执行计划
		results, err := a.executeStreamPlan(ctx, plan, sw)
		if err != nil {
			sw.Send(&schema.Message{
				Role:    schema.Assistant,
				Content: fmt.Sprintf("执行计划失败: %v", err),
			}, nil)
			return
		}

		// 发送最终结果
		finalResponse := a.consolidateResults(results)
		sw.Send(&schema.Message{
			Role:    schema.Assistant,
			Content: finalResponse,
		}, nil)
	}()

	return sr, nil
}

// Initialize 初始化智能体
func (a *PlannerAgent) Initialize(ctx context.Context) error {
	// 初始化所有注册的专家智能体
	a.agentMutex.RLock()
	defer a.agentMutex.RUnlock()

	for _, agent := range a.expertAgents {
		if err := agent.Initialize(ctx); err != nil {
			return fmt.Errorf("初始化专家智能体 %s 失败: %w", agent.GetType(), err)
		}
	}

	return nil
}

// Shutdown 关闭智能体
func (a *PlannerAgent) Shutdown(ctx context.Context) error {
	// 关闭所有注册的专家智能体
	a.agentMutex.RLock()
	defer a.agentMutex.RUnlock()

	for _, agent := range a.expertAgents {
		if err := agent.Shutdown(ctx); err != nil {
			return fmt.Errorf("关闭专家智能体 %s 失败: %w", agent.GetType(), err)
		}
	}

	return nil
}

// createExecutionPlan 创建执行计划
func (a *PlannerAgent) createExecutionPlan(ctx context.Context, queryIntent *types.QueryIntent) (*types.ExecutionPlan, error) {
	// 构建任务规划的系统提示
	systemPrompt := a.buildPlanningPrompt(queryIntent)

	// 使用LLM生成执行计划
	planMessages := []*schema.Message{
		{
			Role:    schema.System,
			Content: systemPrompt,
		},
		{
			Role:    schema.User,
			Content: "请为这个查询意图创建详细的执行计划。",
		},
	}

	response, err := a.chatModel.Generate(ctx, planMessages)
	if err != nil {
		return nil, fmt.Errorf("生成执行计划失败: %w", err)
	}

	// 解析LLM返回的执行计划
	plan, err := a.parseExecutionPlan(response.Content, queryIntent)
	if err != nil {
		// 如果解析失败，创建一个简单的默认计划
		return a.createDefaultPlan(queryIntent), nil
	}

	return plan, nil
}

// buildPlanningPrompt 构建任务规划提示
func (a *PlannerAgent) buildPlanningPrompt(queryIntent *types.QueryIntent) string {
	// 获取可用的专家智能体信息
	a.agentMutex.RLock()
	availableAgents := make([]string, 0, len(a.expertAgents))
	for agentType, agent := range a.expertAgents {
		capabilities := agent.GetCapabilities()
		agentInfo := fmt.Sprintf("- %s: %s", string(agentType), strings.Join(capabilities, ", "))
		availableAgents = append(availableAgents, agentInfo)
	}
	a.agentMutex.RUnlock()

	intentJSON, _ := json.MarshalIndent(queryIntent, "", "  ")

	return fmt.Sprintf(`你是一个专业的任务规划专家。请根据查询意图创建详细的执行计划。

查询意图:
%s

可用的专家智能体:
%s

请创建一个执行计划，包含以下信息：
1. 将复杂任务分解为多个可执行的子任务
2. 确定任务之间的依赖关系
3. 为每个任务分配合适的专家智能体
4. 确定任务的执行顺序（串行或并行）

返回JSON格式的执行计划，包含：
- id: 计划ID
- tasks: 任务列表，每个任务包含:
  - id: 任务ID
  - type: 任务类型
  - description: 任务描述
  - agent_type: 负责执行的智能体类型
  - input: 任务输入
  - dependencies: 依赖的任务ID列表
- dependencies: 任务依赖关系映射

示例:
{
  "id": "plan_xxx",
  "tasks": [
    {
      "id": "task_1",
      "type": "data_query",
      "description": "查询基础数据",
      "agent_type": "data_query",
      "input": {"query": "..."},
      "dependencies": []
    },
    {
      "id": "task_2", 
      "type": "analysis",
      "description": "数据分析",
      "agent_type": "data_analysis",
      "input": {"data_source": "task_1"},
      "dependencies": ["task_1"]
    }
  ],
  "dependencies": {
    "task_2": ["task_1"]
  }
}`, intentJSON, strings.Join(availableAgents, "\n"))
}

// parseExecutionPlan 解析执行计划
func (a *PlannerAgent) parseExecutionPlan(content string, queryIntent *types.QueryIntent) (*types.ExecutionPlan, error) {
	// 提取JSON内容
	jsonStr := a.extractJSON(content)
	if jsonStr == "" {
		return nil, fmt.Errorf("未找到有效的JSON内容")
	}

	var planData struct {
		ID           string                   `json:"id"`
		Tasks        []map[string]interface{} `json:"tasks"`
		Dependencies map[string][]string      `json:"dependencies"`
	}

	err := json.Unmarshal([]byte(jsonStr), &planData)
	if err != nil {
		return nil, fmt.Errorf("JSON解析失败: %w", err)
	}

	// 构建执行计划
	plan := &types.ExecutionPlan{
		ID:           planData.ID,
		QueryIntent:  queryIntent,
		Tasks:        make([]*types.Task, 0, len(planData.Tasks)),
		Dependencies: planData.Dependencies,
	}

	// 转换任务
	for _, taskData := range planData.Tasks {
		task := &types.Task{
			ID:          taskData["id"].(string),
			Type:        taskData["type"].(string),
			Description: taskData["description"].(string),
			AgentType:   types.AgentType(taskData["agent_type"].(string)),
			Input:       taskData["input"],
			Status:      types.TaskStatusPending,
		}

		if deps, ok := taskData["dependencies"].([]interface{}); ok {
			task.Dependencies = make([]string, len(deps))
			for i, dep := range deps {
				task.Dependencies[i] = dep.(string)
			}
		}

		plan.Tasks = append(plan.Tasks, task)
	}

	return plan, nil
}

// createDefaultPlan 创建默认执行计划
func (a *PlannerAgent) createDefaultPlan(queryIntent *types.QueryIntent) *types.ExecutionPlan {
	planID := uuid.New().String()

	// 根据意图类型创建对应的任务
	var tasks []*types.Task

	switch queryIntent.IntentType {
	case "data_query":
		tasks = []*types.Task{
			{
				ID:          uuid.New().String(),
				Type:        "data_query",
				Description: "执行数据查询",
				AgentType:   types.AgentTypeDataQuery,
				Input:       queryIntent.QueryObject,
				Status:      types.TaskStatusPending,
			},
		}
	case "analysis":
		queryTaskID := uuid.New().String()
		analysisTaskID := uuid.New().String()

		tasks = []*types.Task{
			{
				ID:          queryTaskID,
				Type:        "data_query",
				Description: "查询分析数据",
				AgentType:   types.AgentTypeDataQuery,
				Input:       queryIntent.QueryObject,
				Status:      types.TaskStatusPending,
			},
			{
				ID:           analysisTaskID,
				Type:         "analysis",
				Description:  "执行数据分析",
				AgentType:    types.AgentTypeDataAnalysis,
				Input:        map[string]interface{}{"data_source": queryTaskID},
				Dependencies: []string{queryTaskID},
				Status:       types.TaskStatusPending,
			},
		}
	default:
		// 默认分析任务
		tasks = []*types.Task{
			{
				ID:          uuid.New().String(),
				Type:        "analysis",
				Description: "执行综合分析",
				AgentType:   types.AgentTypeDataAnalysis,
				Input:       queryIntent.QueryObject,
				Status:      types.TaskStatusPending,
			},
		}
	}

	return &types.ExecutionPlan{
		ID:           planID,
		QueryIntent:  queryIntent,
		Tasks:        tasks,
		Dependencies: map[string][]string{},
	}
}

// executePlan 执行执行计划
func (a *PlannerAgent) executePlan(ctx context.Context, plan *types.ExecutionPlan) (map[string]*types.TaskResult, error) {
	results := make(map[string]*types.TaskResult)
	completedTasks := make(map[string]bool)

	// 执行任务直到所有任务完成
	for len(completedTasks) < len(plan.Tasks) {
		// 找到可以执行的任务（依赖已完成）
		readyTasks := a.findReadyTasks(plan.Tasks, completedTasks)

		if len(readyTasks) == 0 {
			return nil, fmt.Errorf("没有可执行的任务，可能存在循环依赖")
		}

		// 并行执行就绪的任务
		taskResults, err := a.executeTasksBatch(ctx, readyTasks, results)
		if err != nil {
			return nil, err
		}

		// 更新结果和完成状态
		for taskID, result := range taskResults {
			results[taskID] = result
			completedTasks[taskID] = true
		}
	}

	return results, nil
}

// executeStreamPlan 流式执行执行计划
func (a *PlannerAgent) executeStreamPlan(ctx context.Context, plan *types.ExecutionPlan, sw *schema.StreamWriter[*schema.Message]) (map[string]*types.TaskResult, error) {
	results := make(map[string]*types.TaskResult)
	completedTasks := make(map[string]bool)

	for len(completedTasks) < len(plan.Tasks) {
		readyTasks := a.findReadyTasks(plan.Tasks, completedTasks)

		if len(readyTasks) == 0 {
			return nil, fmt.Errorf("没有可执行的任务，可能存在循环依赖")
		}

		// 发送任务开始消息
		for _, task := range readyTasks {
			sw.Send(&schema.Message{
				Role:    schema.Assistant,
				Content: fmt.Sprintf("开始执行任务: %s - %s", task.ID, task.Description),
			}, nil)
		}

		// 执行任务
		taskResults, err := a.executeTasksBatch(ctx, readyTasks, results)
		if err != nil {
			return nil, err
		}

		// 发送任务完成消息
		for taskID, result := range taskResults {
			status := "成功"
			if !result.Success {
				status = "失败: " + result.Error
			}
			sw.Send(&schema.Message{
				Role:    schema.Assistant,
				Content: fmt.Sprintf("任务 %s 执行%s", taskID, status),
			}, nil)

			results[taskID] = result
			completedTasks[taskID] = true
		}
	}

	return results, nil
}

// findReadyTasks 找到可以执行的任务
func (a *PlannerAgent) findReadyTasks(tasks []*types.Task, completed map[string]bool) []*types.Task {
	var readyTasks []*types.Task

	for _, task := range tasks {
		if completed[task.ID] {
			continue
		}

		// 检查依赖是否都已完成
		allDepsCompleted := true
		for _, depID := range task.Dependencies {
			if !completed[depID] {
				allDepsCompleted = false
				break
			}
		}

		if allDepsCompleted {
			readyTasks = append(readyTasks, task)
		}
	}

	return readyTasks
}

// executeTasksBatch 批量执行任务
func (a *PlannerAgent) executeTasksBatch(ctx context.Context, tasks []*types.Task, previousResults map[string]*types.TaskResult) (map[string]*types.TaskResult, error) {
	results := make(map[string]*types.TaskResult)

	// 使用goroutine并行执行任务
	var wg sync.WaitGroup
	var mutex sync.Mutex

	for _, task := range tasks {
		wg.Add(1)
		go func(t *types.Task) {
			defer wg.Done()

			result := a.executeTask(ctx, t, previousResults)

			mutex.Lock()
			results[t.ID] = result
			mutex.Unlock()
		}(task)
	}

	wg.Wait()
	return results, nil
}

// executeTask 执行单个任务
func (a *PlannerAgent) executeTask(ctx context.Context, task *types.Task, previousResults map[string]*types.TaskResult) *types.TaskResult {
	task.Status = types.TaskStatusRunning

	// 获取对应的专家智能体
	a.agentMutex.RLock()
	agent, exists := a.expertAgents[task.AgentType]
	a.agentMutex.RUnlock()

	if !exists {
		return &types.TaskResult{
			Success:    false,
			Error:      fmt.Sprintf("未找到类型为 %s 的专家智能体", task.AgentType),
			ExecutedBy: task.AgentType,
		}
	}

	// 检查智能体是否能处理此任务
	if !agent.CanHandle(task) {
		return &types.TaskResult{
			Success:    false,
			Error:      fmt.Sprintf("智能体 %s 无法处理此任务", task.AgentType),
			ExecutedBy: task.AgentType,
		}
	}

	// 执行任务
	result, err := agent.ExecuteTask(ctx, task)
	if err != nil {
		task.Status = types.TaskStatusFailed
		return &types.TaskResult{
			Success:    false,
			Error:      err.Error(),
			ExecutedBy: task.AgentType,
		}
	}

	task.Status = types.TaskStatusCompleted
	task.Result = result

	return result
}

// consolidateResults 整合结果
func (a *PlannerAgent) consolidateResults(results map[string]*types.TaskResult) string {
	var response strings.Builder
	response.WriteString("任务执行完成！\n\n")

	successCount := 0
	for taskID, result := range results {
		if result.Success {
			successCount++
			response.WriteString(fmt.Sprintf("✅ 任务 %s 执行成功\n", taskID))
			if result.Output != nil {
				if outputStr, ok := result.Output.(string); ok {
					response.WriteString(fmt.Sprintf("   结果: %s\n", outputStr))
				}
			}
		} else {
			response.WriteString(fmt.Sprintf("❌ 任务 %s 执行失败: %s\n", taskID, result.Error))
		}
	}

	response.WriteString(fmt.Sprintf("\n总计: %d/%d 任务成功完成", successCount, len(results)))

	return response.String()
}

// extractJSON 从文本中提取JSON内容
func (a *PlannerAgent) extractJSON(content string) string {
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
