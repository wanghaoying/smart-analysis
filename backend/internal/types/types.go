package types

import (
	"context"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"smart-analysis/internal/utils/sanbox"
)

// AgentType 智能体类型
type AgentType string

const (
	AgentTypeMaster              AgentType = "master"
	AgentTypePlanner             AgentType = "planner"
	AgentTypeDataQuery           AgentType = "data_query"
	AgentTypeDataAnalysis        AgentType = "data_analysis"
	AgentTypeTrendForecast       AgentType = "trend_forecast"
	AgentTypeAnomalyDetection    AgentType = "anomaly_detection"
	AgentTypeAttributionAnalysis AgentType = "attribution_analysis"
	AgentTypeReact               AgentType = "react"
	AgentTypeAnalysis            AgentType = "analysis"
	AgentTypeMulti               AgentType = "multi"
)

// FileData 文件数据结构
type FileData struct {
	ID       int                    `json:"id"`
	Name     string                 `json:"name"`
	Path     string                 `json:"path"`
	Size     int64                  `json:"size"`
	Type     string                 `json:"type"`
	Content  string                 `json:"content,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// AnalysisContext 分析上下文
type AnalysisContext struct {
	SessionID int                    `json:"session_id"`
	UserID    int                    `json:"user_id"`
	FileData  *FileData              `json:"file_data,omitempty"`
	Query     string                 `json:"query"`
	History   []*schema.Message      `json:"history,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// AnalysisResult 分析结果
type AnalysisResult struct {
	Type         string                 `json:"type"` // "text", "image", "table", "chart", "json"
	Content      interface{}            `json:"content"`
	Description  string                 `json:"description,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	ExecutionLog string                 `json:"execution_log,omitempty"`
}

// EChartsConfig ECharts图表配置
type EChartsConfig struct {
	Type    string                   `json:"type"` // "bar", "line", "pie", "scatter", "heatmap"
	Title   string                   `json:"title"`
	Data    []map[string]interface{} `json:"data"`
	XAxis   []string                 `json:"xAxis,omitempty"`
	Series  []EChartsSeries          `json:"series,omitempty"`
	Options map[string]interface{}   `json:"options,omitempty"`
}

// EChartsSeries ECharts系列数据
type EChartsSeries struct {
	Name string    `json:"name"`
	Type string    `json:"type"`
	Data []float64 `json:"data"`
}

// AgentConfig 智能体配置
type AgentConfig struct {
	ChatModel     model.BaseChatModel    `json:"-"`
	PythonSandbox *sanbox.PythonSandbox  `json:"-"`
	Tools         []tool.BaseTool        `json:"-"`
	MaxSteps      int                    `json:"max_steps"`
	EnableDebug   bool                   `json:"enable_debug"`
	Model         string                 `json:"model"`
	Temperature   float64                `json:"temperature"`
	MaxTokens     int                    `json:"max_tokens"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// Agent 智能体接口
type Agent interface {
	// GetType 获取智能体类型
	GetType() AgentType

	// Generate 生成响应
	Generate(ctx context.Context, messages []*schema.Message, opts ...interface{}) (*schema.Message, error)

	// Stream 流式生成响应
	Stream(ctx context.Context, messages []*schema.Message, opts ...interface{}) (*schema.StreamReader[*schema.Message], error)

	// Initialize 初始化智能体
	Initialize(ctx context.Context) error

	// Shutdown 关闭智能体
	Shutdown(ctx context.Context) error
}

// Task 任务定义
type Task struct {
	ID           string                 `json:"id"`
	Type         string                 `json:"type"`
	Description  string                 `json:"description"`
	AgentType    AgentType              `json:"agent_type"`
	Input        interface{}            `json:"input"`
	Dependencies []string               `json:"dependencies,omitempty"`
	Status       TaskStatus             `json:"status"`
	Result       *TaskResult            `json:"result,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// TaskStatus 任务状态
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
	TaskStatusCancelled TaskStatus = "cancelled"
)

// TaskResult 任务结果
type TaskResult struct {
	Success      bool                   `json:"success"`
	Output       interface{}            `json:"output"`
	Error        string                 `json:"error,omitempty"`
	ExecutedBy   AgentType              `json:"executed_by"`
	ExecutionLog string                 `json:"execution_log,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// QueryIntent 查询意图
type QueryIntent struct {
	IntentType   string                 `json:"intent_type"` // "data_query", "analysis", "visualization", etc.
	DataSchema   *DataSchema            `json:"data_schema"`
	QueryObject  *QueryObject           `json:"query_object"`
	Requirements []string               `json:"requirements"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// DataSchema 数据模式
type DataSchema struct {
	TableName   string                 `json:"table_name"`
	Columns     []ColumnInfo           `json:"columns"`
	Constraints []string               `json:"constraints,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// ColumnInfo 列信息
type ColumnInfo struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Description string   `json:"description,omitempty"`
	IsKey       bool     `json:"is_key,omitempty"`
	Values      []string `json:"values,omitempty"` // 对于枚举类型
}

// QueryObject 查询对象
type QueryObject struct {
	Events     []string               `json:"events,omitempty"`     // 事件
	Dimensions []string               `json:"dimensions,omitempty"` // 维度
	Metrics    []string               `json:"metrics,omitempty"`    // 度量
	Filters    []FilterCondition      `json:"filters,omitempty"`    // 过滤条件
	TimeRange  *TimeRange             `json:"time_range,omitempty"` // 时间范围
	GroupBy    []string               `json:"group_by,omitempty"`   // 分组
	OrderBy    []OrderCondition       `json:"order_by,omitempty"`   // 排序
	Limit      int                    `json:"limit,omitempty"`      // 限制数量
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// FilterCondition 过滤条件
type FilterCondition struct {
	Column   string      `json:"column"`
	Operator string      `json:"operator"` // "=", "!=", ">", "<", ">=", "<=", "IN", "NOT IN", "LIKE"
	Value    interface{} `json:"value"`
}

// TimeRange 时间范围
type TimeRange struct {
	StartTime   string `json:"start_time"`
	EndTime     string `json:"end_time"`
	Granularity string `json:"granularity,omitempty"` // "day", "hour", "month", etc.
}

// OrderCondition 排序条件
type OrderCondition struct {
	Column    string `json:"column"`
	Direction string `json:"direction"` // "ASC", "DESC"
}

// ExecutionPlan 执行计划
type ExecutionPlan struct {
	ID           string                 `json:"id"`
	QueryIntent  *QueryIntent           `json:"query_intent"`
	Tasks        []*Task                `json:"tasks"`
	Dependencies map[string][]string    `json:"dependencies"` // task_id -> dependency_task_ids
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// ExpertAgent 专家智能体接口
type ExpertAgent interface {
	Agent

	// GetCapabilities 获取能力描述
	GetCapabilities() []string

	// CanHandle 判断是否能处理特定任务
	CanHandle(task *Task) bool

	// ExecuteTask 执行任务
	ExecuteTask(ctx context.Context, task *Task) (*TaskResult, error)
}
