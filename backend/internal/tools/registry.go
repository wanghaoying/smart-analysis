package tools

import (
	"github.com/cloudwego/eino/components/tool"
	"smart-analysis/internal/utils/sanbox"
)

// ToolRegistry 工具注册器
type ToolRegistry struct {
	sandbox *sanbox.PythonSandbox
	tools   map[string]tool.BaseTool
}

// ToolConfig 工具配置
type ToolConfig struct {
	EnableCoreTools     bool // 启用核心工具（Python分析、ECharts可视化等）
	EnableAdvancedTools bool // 启用高级工具（ML、预处理等）
	EnableOptionalTools bool // 启用可选工具（文本分析、数据库等）
	EnableTestingTools  bool // 启用测试工具
}

// DefaultToolConfig 默认工具配置
func DefaultToolConfig() *ToolConfig {
	return &ToolConfig{
		EnableCoreTools:     true,
		EnableAdvancedTools: true,
		EnableOptionalTools: true,
		EnableTestingTools:  false, // 默认不启用测试工具
	}
}

// NewToolRegistry 创建新的工具注册器
func NewToolRegistry(sandbox *sanbox.PythonSandbox) *ToolRegistry {
	return &ToolRegistry{
		sandbox: sandbox,
		tools:   make(map[string]tool.BaseTool),
	}
}

// RegisterAllTools 注册所有工具
func (tr *ToolRegistry) RegisterAllTools() []tool.BaseTool {
	return tr.RegisterToolsWithConfig(DefaultToolConfig())
}

// RegisterToolsWithConfig 根据配置注册工具
func (tr *ToolRegistry) RegisterToolsWithConfig(config *ToolConfig) []tool.BaseTool {
	// 注册核心工具
	if config.EnableCoreTools {
		tr.tools["python_analysis"] = NewPythonAnalysisTool(tr.sandbox)
		tr.tools["echarts_visualization"] = NewEChartsVisualizationTool(tr.sandbox)
		tr.tools["file_reader"] = NewFileReaderTool(tr.sandbox)
		tr.tools["data_query"] = NewDataQueryTool(tr.sandbox)
	}

	// 注册高级工具
	if config.EnableAdvancedTools {
		tr.tools["data_preprocessing"] = NewDataPreprocessingTool(tr.sandbox)
		tr.tools["ml_analysis"] = NewMLAnalysisTool(tr.sandbox)
	}

	// 注册可选工具
	if config.EnableOptionalTools {
		tr.tools["text_analysis"] = NewTextAnalysisTool(tr.sandbox)
		tr.tools["report_generator"] = NewReportGeneratorTool(tr.sandbox)
		tr.tools["database_tool"] = NewDatabaseTool(tr.sandbox)
	}

	// 注册测试工具
	if config.EnableTestingTools {
		// tr.tools["system_test"] = NewSystemTestTool(tr.sandbox)
		// 暂时注释掉测试工具，等其他工具稳定后再启用
	}

	// 转换为切片返回
	var toolList []tool.BaseTool
	for _, t := range tr.tools {
		toolList = append(toolList, t)
	}

	return toolList
}

// RegisterCoreToolsOnly 只注册核心工具
func (tr *ToolRegistry) RegisterCoreToolsOnly() []tool.BaseTool {
	config := &ToolConfig{
		EnableCoreTools:     true,
		EnableAdvancedTools: false,
		EnableOptionalTools: false,
		EnableTestingTools:  false,
	}
	return tr.RegisterToolsWithConfig(config)
}

// GetTool 获取指定工具
func (tr *ToolRegistry) GetTool(name string) (tool.BaseTool, bool) {
	t, exists := tr.tools[name]
	return t, exists
}

// GetAllTools 获取所有工具
func (tr *ToolRegistry) GetAllTools() map[string]tool.BaseTool {
	return tr.tools
}

// GetToolNames 获取所有工具名称
func (tr *ToolRegistry) GetToolNames() []string {
	names := make([]string, 0, len(tr.tools))
	for name := range tr.tools {
		names = append(names, name)
	}
	return names
}

// GetToolCount 获取工具数量
func (tr *ToolRegistry) GetToolCount() int {
	return len(tr.tools)
}
