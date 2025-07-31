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
	AgentTypeMain     AgentType = "main"
	AgentTypeReact    AgentType = "react"
	AgentTypeAnalysis AgentType = "analysis"
	AgentTypeMulti    AgentType = "multi"
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
	ChatModel     model.BaseChatModel   `json:"-"`
	PythonSandbox *sanbox.PythonSandbox `json:"-"`
	Tools         []tool.BaseTool       `json:"-"`
	MaxSteps      int                   `json:"max_steps"`
	EnableDebug   bool                  `json:"enable_debug"`
	Model         string                `json:"model"`
	Temperature   float64               `json:"temperature"`
	MaxTokens     int                   `json:"max_tokens"`
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
