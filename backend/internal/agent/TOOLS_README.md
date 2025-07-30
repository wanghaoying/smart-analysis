# Tools.go 重构内容说明

由于文件创建时遇到技术问题，请手动创建 `tools.go` 文件，包含以下主要内容：

## 1. 类型定义

```go
// AgentType 智能体类型
type AgentType string

const (
    AgentTypeMain     AgentType = "main"
    AgentTypeReact    AgentType = "react"
    AgentTypeAnalysis AgentType = "analysis"
    AgentTypeMulti    AgentType = "multi"
)

// EChartsConfig ECharts图表配置
type EChartsConfig struct {
    Type    string                   `json:"type"`
    Title   string                   `json:"title"`
    Data    []map[string]interface{} `json:"data"`
    XAxis   []string                 `json:"xAxis,omitempty"`
    Series  []EChartsSeries          `json:"series,omitempty"`
    Options map[string]interface{}   `json:"options,omitempty"`
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
```

## 2. 主要工具

### PythonAnalysisTool
- 支持通用Python执行
- 集成统计分析功能（`analysis_type: "statistical"`）
- 数据清洗功能（`analysis_type: "cleaning"`）

### EChartsVisualizationTool  
- 生成ECharts格式配置
- 支持bar、line、pie、scatter、heatmap图表
- 返回JSON格式配置供前端渲染

### FileReaderTool
- 读取CSV、Excel、JSON文件
- 提供数据预览和基础信息

### DataQueryTool
- 支持pandas查询语法
- 数据筛选、分组、聚合操作

## 3. 关键实现要点

1. 所有工具继承 `tool.BaseTool` 接口
2. 实现 `Info()` 和 `InvokableRun()` 方法
3. ECharts工具返回标准JSON配置格式
4. Python工具支持预处理代码注入
5. 错误处理和结果格式化

完整的工具实现代码请参考前面创建的版本或从重构总结文档中获取详细信息。
