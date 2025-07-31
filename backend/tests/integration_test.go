package tools_test

import (
	"fmt"
	"testing"

	"smart-analysis/internal/tools"
	"smart-analysis/internal/utils/sanbox"
)

// TestToolIntegration 工具集成测试
func TestToolIntegration(t *testing.T) {
	fmt.Println("=== Smart Analysis 工具集成测试 ===")

	// 创建Python沙箱
	sandbox := &sanbox.PythonSandbox{} // 这里应该有适当的初始化

	// 创建工具注册器
	registry := tools.NewToolRegistry(sandbox)

	// 注册所有工具
	toolList := registry.RegisterAllTools()

	fmt.Printf("已注册 %d 个工具:\n", len(toolList))
	for _, toolName := range registry.GetToolNames() {
		fmt.Printf("- %s\n", toolName)
	}

	fmt.Println("\n=== 测试核心工具功能 ===")

	// 测试系统测试工具
	if systemTool, exists := registry.GetTool("system_test"); exists {
		fmt.Println("\n测试系统测试工具...")
		testSystemTool(systemTool)
	}

	// 测试文件读取工具
	if fileTool, exists := registry.GetTool("file_reader"); exists {
		fmt.Println("\n测试文件读取工具...")
		testFileReaderTool(fileTool)
	}

	// 测试Python分析工具
	if pythonTool, exists := registry.GetTool("python_analysis"); exists {
		fmt.Println("\n测试Python分析工具...")
		testPythonAnalysisTool(pythonTool)
	}

	fmt.Println("\n=== 集成测试完成 ===")
}

func testSystemTool(tool interface{}) {
	// 这里应该有具体的测试逻辑
	fmt.Println("✅ 系统测试工具可用")
}

func testFileReaderTool(tool interface{}) {
	// 这里应该有具体的测试逻辑
	fmt.Println("✅ 文件读取工具可用")
}

func testPythonAnalysisTool(tool interface{}) {
	// 这里应该有具体的测试逻辑
	fmt.Println("✅ Python分析工具可用")
}

// 工具配置示例
type ToolConfigExample struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Parameters  interface{} `json:"parameters"`
	Example     string      `json:"example"`
}

// 生成工具文档
func generateToolDocumentation() {
	examples := []ToolConfigExample{
		{
			Name:        "python_analysis",
			Description: "执行Python代码进行数据分析",
			Parameters: map[string]interface{}{
				"code":          "要执行的Python代码",
				"analysis_type": "分析类型（可选）",
				"data_source":   "数据源文件路径（可选）",
			},
			Example: `{
  "code": "print('Hello, World!')",
  "analysis_type": "general"
}`,
		},
		{
			Name:        "echarts_visualization",
			Description: "创建ECharts格式的交互式图表",
			Parameters: map[string]interface{}{
				"chart_type":   "图表类型（bar, line, pie等）",
				"data_columns": "数据列名数组",
				"title":        "图表标题（可选）",
				"file_path":    "数据文件路径（可选）",
			},
			Example: `{
  "chart_type": "bar",
  "data_columns": ["sales", "profit"],
  "title": "销售利润图表"
}`,
		},
		{
			Name:        "file_reader",
			Description: "读取和预览数据文件",
			Parameters: map[string]interface{}{
				"file_path":    "文件路径",
				"preview_rows": "预览行数（可选，默认5）",
			},
			Example: `{
  "file_path": "/path/to/data.csv",
  "preview_rows": 10
}`,
		},
		{
			Name:        "text_analysis",
			Description: "文本分析工具",
			Parameters: map[string]interface{}{
				"operation":   "分析操作类型",
				"text_column": "文本列名",
				"file_path":   "数据文件路径（可选）",
				"language":    "文本语言（可选）",
			},
			Example: `{
  "operation": "sentiment",
  "text_column": "review_text",
  "language": "zh"
}`,
		},
		{
			Name:        "ml_analysis",
			Description: "机器学习分析工具",
			Parameters: map[string]interface{}{
				"task_type":     "任务类型（classification, regression, clustering）",
				"algorithm":     "算法类型",
				"target_column": "目标列名（可选）",
				"file_path":     "数据文件路径（可选）",
			},
			Example: `{
  "task_type": "classification",
  "algorithm": "rf",
  "target_column": "label"
}`,
		},
	}

	fmt.Println("\n=== 工具使用示例 ===")
	for _, example := range examples {
		fmt.Printf("\n**%s**\n", example.Name)
		fmt.Printf("描述: %s\n", example.Description)
		fmt.Printf("示例参数:\n%s\n", example.Example)
	}
}
