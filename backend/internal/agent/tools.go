package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

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

// PythonAnalysisTool Python分析工具（包含统计分析功能）
type PythonAnalysisTool struct {
	sandbox *sanbox.PythonSandbox
	name    string
	desc    string
}

// NewPythonAnalysisTool 创建Python分析工具
func NewPythonAnalysisTool(sandbox *sanbox.PythonSandbox) *PythonAnalysisTool {
	return &PythonAnalysisTool{
		sandbox: sandbox,
		name:    "python_analysis",
		desc:    "执行Python代码进行数据分析、统计计算和数据处理。支持pandas、numpy、scipy等数据科学库。",
	}
}

// Info 返回工具信息
func (t *PythonAnalysisTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: t.name,
		Desc: t.desc,
		ParamsOneOf: schema.NewParamsOneOfByParams(
			map[string]*schema.ParameterInfo{
				"code": {
					Type:     schema.String,
					Desc:     "要执行的Python代码",
					Required: true,
				},
				"analysis_type": {
					Type:     schema.String,
					Desc:     "分析类型：general（通用）、statistical（统计分析）、cleaning（数据清洗）",
					Required: false,
				},
			}),
	}, nil
}

// InvokableRun 执行工具
func (t *PythonAnalysisTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var args struct {
		Code         string `json:"code"`
		AnalysisType string `json:"analysis_type,omitempty"`
	}

	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", err
	}

	// 根据分析类型添加预处理代码
	finalCode := t.preprocessCode(args.Code, args.AnalysisType)

	// 执行Python代码
	result, err := t.sandbox.ExecutePython(finalCode)
	if err != nil {
		return "", err
	}

	if !result.Success {
		return "执行失败: " + result.Error, nil
	}

	// 格式化结果
	return t.formatResult(result), nil
}

// preprocessCode 预处理代码，添加常用的统计分析功能
func (t *PythonAnalysisTool) preprocessCode(code, analysisType string) string {
	prelude := `
import pandas as pd
import numpy as np
import matplotlib.pyplot as plt
import seaborn as sns
from scipy import stats
import warnings
warnings.filterwarnings('ignore')

# 设置中文字体
plt.rcParams['font.sans-serif'] = ['SimHei', 'Arial Unicode MS', 'DejaVu Sans']
plt.rcParams['axes.unicode_minus'] = False

`

	switch analysisType {
	case "statistical":
		prelude += `
# 统计分析辅助函数
def describe_data(df):
    """数据描述性统计"""
    return {
        'shape': df.shape,
        'dtypes': df.dtypes.to_dict(),
        'null_counts': df.isnull().sum().to_dict(),
        'describe': df.describe().to_dict()
    }

def correlation_analysis(df, method='pearson'):
    """相关性分析"""
    numeric_cols = df.select_dtypes(include=[np.number]).columns
    if len(numeric_cols) > 1:
        return df[numeric_cols].corr(method=method).to_dict()
    return {}

def statistical_tests(df, col1, col2=None):
    """统计检验"""
    results = {}
    if col2 is None:
        # 单样本检验
        if df[col1].dtype in ['int64', 'float64']:
            stat, p_value = stats.normaltest(df[col1].dropna())
            results['normality_test'] = {'statistic': stat, 'p_value': p_value}
    else:
        # 双样本检验
        if df[col1].dtype in ['int64', 'float64'] and df[col2].dtype in ['int64', 'float64']:
            stat, p_value = stats.pearsonr(df[col1].dropna(), df[col2].dropna())
            results['correlation_test'] = {'statistic': stat, 'p_value': p_value}
    return results

`
	case "cleaning":
		prelude += `
# 数据清洗辅助函数
def clean_data(df):
    """基本数据清洗"""
    cleaned_df = df.copy()
    
    # 移除完全重复的行
    cleaned_df = cleaned_df.drop_duplicates()
    
    # 填充数值型列的缺失值（使用中位数）
    numeric_cols = cleaned_df.select_dtypes(include=[np.number]).columns
    for col in numeric_cols:
        cleaned_df[col].fillna(cleaned_df[col].median(), inplace=True)
    
    # 填充类别型列的缺失值（使用众数）
    categorical_cols = cleaned_df.select_dtypes(include=['object']).columns
    for col in categorical_cols:
        mode_val = cleaned_df[col].mode()
        if len(mode_val) > 0:
            cleaned_df[col].fillna(mode_val[0], inplace=True)
    
    return cleaned_df

def detect_outliers(df, column, method='iqr'):
    """异常值检测"""
    if method == 'iqr':
        Q1 = df[column].quantile(0.25)
        Q3 = df[column].quantile(0.75)
        IQR = Q3 - Q1
        lower_bound = Q1 - 1.5 * IQR
        upper_bound = Q3 + 1.5 * IQR
        return df[(df[column] < lower_bound) | (df[column] > upper_bound)]
    return pd.DataFrame()

`
	}

	return prelude + "\n" + code
}

// formatResult 格式化执行结果
func (t *PythonAnalysisTool) formatResult(result *sanbox.PythonExecutionResult) string {
	resultStr := "执行成功:\n"

	if result.Stdout != "" {
		resultStr += "输出:\n" + result.Stdout + "\n"
	}

	if result.Output != nil {
		outputJSON, _ := json.Marshal(result.Output)
		resultStr += "结果数据:\n" + string(outputJSON) + "\n"
	}

	if result.ImagePath != "" {
		resultStr += "生成图片: " + result.ImagePath + "\n"
	}

	return resultStr
}

// EChartsVisualizationTool ECharts数据可视化工具
type EChartsVisualizationTool struct {
	sandbox *sanbox.PythonSandbox
	name    string
	desc    string
}

// NewEChartsVisualizationTool 创建ECharts数据可视化工具
func NewEChartsVisualizationTool(sandbox *sanbox.PythonSandbox) *EChartsVisualizationTool {
	return &EChartsVisualizationTool{
		sandbox: sandbox,
		name:    "echarts_visualization",
		desc:    "创建ECharts格式的数据可视化图表，返回可在前端渲染的图表配置。",
	}
}

// Info 返回工具信息
func (t *EChartsVisualizationTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: t.name,
		Desc: t.desc,
		ParamsOneOf: schema.NewParamsOneOfByParams(
			map[string]*schema.ParameterInfo{
				"chart_type": {
					Type:     schema.String,
					Desc:     "图表类型: bar（柱状图）, line（折线图）, pie（饼图）, scatter（散点图）, heatmap（热力图）",
					Required: true,
				},
				"data_columns": {
					Type: schema.Array,
					Desc: "用于可视化的数据列名",
					ElemInfo: &schema.ParameterInfo{
						Type: schema.String,
					},
					Required: true,
				},
				"title": {
					Type:     schema.String,
					Desc:     "图表标题",
					Required: false,
				},
				"file_path": {
					Type:     schema.String,
					Desc:     "数据文件路径",
					Required: false,
				},
			}),
	}, nil
}

// InvokableRun 执行工具
func (t *EChartsVisualizationTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var args struct {
		ChartType   string   `json:"chart_type"`
		DataColumns []string `json:"data_columns"`
		Title       string   `json:"title,omitempty"`
		FilePath    string   `json:"file_path,omitempty"`
	}

	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", err
	}

	// 生成ECharts配置代码
	code := t.generateEChartsCode(args.ChartType, args.DataColumns, args.Title, args.FilePath)

	// 执行Python代码
	result, err := t.sandbox.ExecutePython(code)
	if err != nil {
		return "", err
	}

	if !result.Success {
		return "图表生成失败: " + result.Error, nil
	}

	// 格式化结果为ECharts配置
	return t.formatEChartsResult(result), nil
}

// generateEChartsCode 生成ECharts配置的Python代码
func (t *EChartsVisualizationTool) generateEChartsCode(chartType string, columns []string, title, filePath string) string {
	if title == "" {
		title = "数据可视化图表"
	}

	code := fmt.Sprintf(`
import pandas as pd
import numpy as np
import json

# 读取数据
%s

# 数据处理
chart_config = {
    "type": "%s",
    "title": "%s",
    "data": [],
    "xAxis": [],
    "series": [],
    "options": {}
}

`, t.getDataLoadCode(filePath), chartType, title)

	switch chartType {
	case "bar", "line":
		code += t.generateBarLineChartCode(columns)
	case "pie":
		code += t.generatePieChartCode(columns)
	case "scatter":
		code += t.generateScatterChartCode(columns)
	case "heatmap":
		code += t.generateHeatmapChartCode(columns)
	default:
		code += t.generateBarLineChartCode(columns) // 默认柱状图
	}

	code += `
# 输出结果
print("ECHARTS_CONFIG_START")
print(json.dumps(chart_config, ensure_ascii=False, indent=2))
print("ECHARTS_CONFIG_END")
`

	return code
}

// getDataLoadCode 获取数据加载代码
func (t *EChartsVisualizationTool) getDataLoadCode(filePath string) string {
	if filePath != "" {
		return fmt.Sprintf("df = pd.read_csv('%s')", filePath)
	}
	return "# 假设数据已经加载到df变量中"
}

// generateBarLineChartCode 生成柱状图/折线图代码
func (t *EChartsVisualizationTool) generateBarLineChartCode(columns []string) string {
	if len(columns) < 2 {
		return `
# 数据列不足
chart_config["data"] = []
`
	}

	return fmt.Sprintf(`
# 柱状图/折线图配置
x_col = "%s"
y_col = "%s"

if x_col in df.columns and y_col in df.columns:
    chart_config["xAxis"] = df[x_col].astype(str).tolist()
    chart_config["data"] = [
        {"name": str(x), "value": float(y)} 
        for x, y in zip(df[x_col], df[y_col]) 
        if pd.notna(x) and pd.notna(y)
    ]
    chart_config["series"] = [{
        "name": y_col,
        "type": chart_config["type"],
        "data": df[y_col].dropna().tolist()
    }]
`, columns[0], columns[1])
}

// generatePieChartCode 生成饼图代码
func (t *EChartsVisualizationTool) generatePieChartCode(columns []string) string {
	if len(columns) < 2 {
		return `
# 数据列不足
chart_config["data"] = []
`
	}

	return fmt.Sprintf(`
# 饼图配置
name_col = "%s"
value_col = "%s"

if name_col in df.columns and value_col in df.columns:
    chart_config["data"] = [
        {"name": str(name), "value": float(value)} 
        for name, value in zip(df[name_col], df[value_col]) 
        if pd.notna(name) and pd.notna(value)
    ]
`, columns[0], columns[1])
}

// generateScatterChartCode 生成散点图代码
func (t *EChartsVisualizationTool) generateScatterChartCode(columns []string) string {
	if len(columns) < 2 {
		return `
# 数据列不足
chart_config["data"] = []
`
	}

	return fmt.Sprintf(`
# 散点图配置
x_col = "%s"
y_col = "%s"

if x_col in df.columns and y_col in df.columns:
    chart_config["data"] = [
        {"name": f"({x}, {y})", "value": [float(x), float(y)]} 
        for x, y in zip(df[x_col], df[y_col]) 
        if pd.notna(x) and pd.notna(y)
    ]
    chart_config["series"] = [{
        "name": "散点数据",
        "type": "scatter",
        "data": [[float(x), float(y)] for x, y in zip(df[x_col], df[y_col]) if pd.notna(x) and pd.notna(y)]
    }]
`, columns[0], columns[1])
}

// generateHeatmapChartCode 生成热力图代码
func (t *EChartsVisualizationTool) generateHeatmapChartCode(columns []string) string {
	return `
# 热力图配置（使用数值列的相关性矩阵）
numeric_cols = df.select_dtypes(include=[np.number]).columns.tolist()
if len(numeric_cols) >= 2:
    corr_matrix = df[numeric_cols].corr()
    
    data = []
    for i, row_name in enumerate(corr_matrix.index):
        for j, col_name in enumerate(corr_matrix.columns):
            data.append({"name": f"{row_name}-{col_name}", "value": [i, j, float(corr_matrix.iloc[i, j])]})
    
    chart_config["data"] = data
    chart_config["xAxis"] = corr_matrix.columns.tolist()
    chart_config["options"]["yAxis"] = corr_matrix.index.tolist()
`
}

// formatEChartsResult 格式化ECharts结果
func (t *EChartsVisualizationTool) formatEChartsResult(result *sanbox.PythonExecutionResult) string {
	if result.Stdout == "" {
		return "未生成图表配置"
	}

	// 提取ECharts配置
	output := result.Stdout
	startIdx := strings.Index(output, "ECHARTS_CONFIG_START")
	endIdx := strings.Index(output, "ECHARTS_CONFIG_END")

	if startIdx == -1 || endIdx == -1 {
		return "图表配置解析失败: " + output
	}

	configJSON := output[startIdx+len("ECHARTS_CONFIG_START") : endIdx]
	configJSON = strings.TrimSpace(configJSON)

	// 验证JSON格式
	var config EChartsConfig
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		return "图表配置JSON格式错误: " + err.Error()
	}

	return fmt.Sprintf("ECharts图表配置生成成功:\n```json\n%s\n```", configJSON)
}

// FileReaderTool 文件读取工具
type FileReaderTool struct {
	sandbox *sanbox.PythonSandbox
	name    string
	desc    string
}

// NewFileReaderTool 创建文件读取工具
func NewFileReaderTool(sandbox *sanbox.PythonSandbox) *FileReaderTool {
	return &FileReaderTool{
		sandbox: sandbox,
		name:    "file_reader",
		desc:    "读取和预览各种格式的数据文件（CSV、Excel、JSON等），提供数据基本信息。",
	}
}

// Info 返回工具信息
func (t *FileReaderTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: t.name,
		Desc: t.desc,
		ParamsOneOf: schema.NewParamsOneOfByParams(
			map[string]*schema.ParameterInfo{
				"file_path": {
					Type:     schema.String,
					Desc:     "文件路径",
					Required: true,
				},
				"preview_rows": {
					Type:     schema.Number,
					Desc:     "预览行数（默认5行）",
					Required: false,
				},
			}),
	}, nil
}

// InvokableRun 执行工具
func (t *FileReaderTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var args struct {
		FilePath    string `json:"file_path"`
		PreviewRows int    `json:"preview_rows,omitempty"`
	}

	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", err
	}

	if args.PreviewRows <= 0 {
		args.PreviewRows = 5
	}

	code := fmt.Sprintf(`
import pandas as pd
import json
import os
from pathlib import Path

file_path = "%s"
preview_rows = %d

try:
    # 检查文件是否存在
    if not os.path.exists(file_path):
        result = {"error": "文件不存在: " + file_path}
    else:
        # 获取文件信息
        file_info = {
            "file_path": file_path,
            "file_size": os.path.getsize(file_path),
            "file_extension": Path(file_path).suffix.lower()
        }
        
        # 根据文件类型读取
        if file_info["file_extension"] == ".csv":
            df = pd.read_csv(file_path)
        elif file_info["file_extension"] in [".xlsx", ".xls"]:
            df = pd.read_excel(file_path)
        elif file_info["file_extension"] == ".json":
            df = pd.read_json(file_path)
        else:
            result = {"error": "不支持的文件格式: " + file_info["file_extension"]}
            df = None
        
        if df is not None:
            result = {
                "file_info": file_info,
                "data_info": {
                    "shape": df.shape,
                    "columns": df.columns.tolist(),
                    "dtypes": df.dtypes.astype(str).to_dict(),
                    "null_counts": df.isnull().sum().to_dict(),
                    "memory_usage": df.memory_usage(deep=True).sum()
                },
                "preview": df.head(preview_rows).to_dict("records")
            }

except Exception as e:
    result = {"error": f"读取文件时出错: {str(e)}"}

print(json.dumps(result, ensure_ascii=False, indent=2))
`, args.FilePath, args.PreviewRows)

	result, err := t.sandbox.ExecutePython(code)
	if err != nil {
		return "", err
	}

	if !result.Success {
		return "文件读取失败: " + result.Error, nil
	}

	return result.Stdout, nil
}

// DataQueryTool 数据查询工具
type DataQueryTool struct {
	sandbox *sanbox.PythonSandbox
	name    string
	desc    string
}

// NewDataQueryTool 创建数据查询工具
func NewDataQueryTool(sandbox *sanbox.PythonSandbox) *DataQueryTool {
	return &DataQueryTool{
		sandbox: sandbox,
		name:    "data_query",
		desc:    "使用SQL样式的查询语法对数据进行筛选、聚合和分析。",
	}
}

// Info 返回工具信息
func (t *DataQueryTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: t.name,
		Desc: t.desc,
		ParamsOneOf: schema.NewParamsOneOfByParams(
			map[string]*schema.ParameterInfo{
				"query": {
					Type:     schema.String,
					Desc:     "查询语句，支持pandas查询语法",
					Required: true,
				},
				"file_path": {
					Type:     schema.String,
					Desc:     "数据文件路径",
					Required: false,
				},
			}),
	}, nil
}

// InvokableRun 执行工具
func (t *DataQueryTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var args struct {
		Query    string `json:"query"`
		FilePath string `json:"file_path,omitempty"`
	}

	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", err
	}

	code := fmt.Sprintf(`
import pandas as pd
import numpy as np
import json

%s

# 执行查询
try:
    # 安全的查询执行
    query = '''%s'''
    
    # 支持的查询类型
    if 'groupby(' in query.lower():
        # 分组查询
        result_df = eval(f"df.{query}")
    elif 'query(' in query.lower():
        # 条件查询
        result_df = eval(f"df.{query}")
    elif any(agg in query.lower() for agg in ['sum()', 'mean()', 'count()', 'max()', 'min()']):
        # 聚合查询
        result_df = eval(f"df.{query}")
    else:
        # 其他pandas操作
        result_df = eval(f"df.{query}")
    
    # 转换结果
    if isinstance(result_df, pd.DataFrame):
        result = {
            "type": "dataframe",
            "shape": result_df.shape,
            "data": result_df.head(20).to_dict("records"),
            "columns": result_df.columns.tolist()
        }
    elif isinstance(result_df, pd.Series):
        result = {
            "type": "series",
            "data": result_df.head(20).to_dict(),
            "name": result_df.name
        }
    else:
        result = {
            "type": "value",
            "data": str(result_df)
        }
    
    print("查询执行成功:")
    print(json.dumps(result, ensure_ascii=False, indent=2))
    
except Exception as e:
    print(f"查询执行失败: {str(e)}")
`, t.getDataLoadCode(args.FilePath), args.Query)

	result, err := t.sandbox.ExecutePython(code)
	if err != nil {
		return "", err
	}

	if !result.Success {
		return "查询执行失败: " + result.Error, nil
	}

	return result.Stdout, nil
}

// getDataLoadCode 获取数据加载代码
func (t *DataQueryTool) getDataLoadCode(filePath string) string {
	if filePath != "" {
		return fmt.Sprintf("df = pd.read_csv('%s')", filePath)
	}
	return "# 假设数据已经加载到df变量中"
}
