package agent

import (
	"context"
	"encoding/json"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"smart-analysis/internal/utils/sanbox"
)

// EinoAgentType Eino智能体类型
type EinoAgentType string

const (
	EinoAgentTypeMain     EinoAgentType = "main"
	EinoAgentTypeReact    EinoAgentType = "react"
	EinoAgentTypeAnalysis EinoAgentType = "analysis"
	EinoAgentTypeMulti    EinoAgentType = "multi"
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

// EinoAnalysisContext Eino分析上下文
type EinoAnalysisContext struct {
	SessionID int                    `json:"session_id"`
	UserID    int                    `json:"user_id"`
	FileData  *FileData              `json:"file_data,omitempty"`
	Query     string                 `json:"query"`
	History   []*schema.Message      `json:"history,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// EinoAnalysisResult Eino分析结果
type EinoAnalysisResult struct {
	Type         string                 `json:"type"` // "text", "image", "table", "chart", "json"
	Content      interface{}            `json:"content"`
	Description  string                 `json:"description,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	ExecutionLog string                 `json:"execution_log,omitempty"`
	Messages     []*schema.Message      `json:"messages,omitempty"`
}

// EinoAgentConfig Eino智能体配置
type EinoAgentConfig struct {
	ChatModel     model.BaseChatModel   `json:"-"`
	PythonSandbox *sanbox.PythonSandbox `json:"-"`
	Tools         []tool.BaseTool       `json:"-"`
	MaxSteps      int                   `json:"max_steps"`
	EnableDebug   bool                  `json:"enable_debug"`
	Model         string                `json:"model"`
	Temperature   float64               `json:"temperature"`
	MaxTokens     int                   `json:"max_tokens"`
}

// EinoAgent Eino智能体接口
type EinoAgent interface {
	// GetType 获取智能体类型
	GetType() EinoAgentType

	// Generate 生成响应
	Generate(ctx context.Context, messages []*schema.Message, opts ...interface{}) (*schema.Message, error)

	// Stream 流式生成响应
	Stream(ctx context.Context, messages []*schema.Message, opts ...interface{}) (*schema.StreamReader[*schema.Message], error)

	// Initialize 初始化智能体
	Initialize(ctx context.Context) error

	// Shutdown 关闭智能体
	Shutdown(ctx context.Context) error
}

// PythonAnalysisTool Python分析工具
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
		desc:    "Execute Python code for data analysis and return results",
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
			}),
	}, nil
} // InvokableRun 执行工具
func (t *PythonAnalysisTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	// 解析参数
	var args struct {
		Code     string `json:"code"`
		FilePath string `json:"file_path,omitempty"`
	}

	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", err
	}

	// 执行Python代码
	result, err := t.sandbox.ExecutePython(args.Code)
	if err != nil {
		return "", err
	}

	if !result.Success {
		return result.Error, nil
	}

	// 格式化结果
	resultStr := "执行成功:\n"
	if result.Stdout != "" {
		resultStr += "输出: " + result.Stdout + "\n"
	}
	if result.OutputType != "" {
		resultStr += "结果类型: " + result.OutputType + "\n"
	}
	if result.Output != nil {
		outputJSON, _ := json.Marshal(result.Output)
		resultStr += "结果内容: " + string(outputJSON) + "\n"
	}
	if result.ImagePath != "" {
		resultStr += "图片路径: " + result.ImagePath + "\n"
	}

	return resultStr, nil
}

// DataVisualizationTool 数据可视化工具
type DataVisualizationTool struct {
	sandbox *sanbox.PythonSandbox
	name    string
	desc    string
}

// NewDataVisualizationTool 创建数据可视化工具
func NewDataVisualizationTool(sandbox *sanbox.PythonSandbox) *DataVisualizationTool {
	return &DataVisualizationTool{
		sandbox: sandbox,
		name:    "data_visualization",
		desc:    "Create data visualizations and charts using Python",
	}
}

// Info 返回工具信息
func (t *DataVisualizationTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: t.name,
		Desc: t.desc,
		ParamsOneOf: schema.NewParamsOneOfByParams(
			map[string]*schema.ParameterInfo{
				"chart_type": {
					Type:     schema.String,
					Desc:     "Type of chart to create (histogram, scatter, line, bar, etc.)",
					Required: true,
				},
				"data_columns": {
					Type: schema.Array,
					Desc: "Columns to use for visualization",
					ElemInfo: &schema.ParameterInfo{
						Type: schema.String,
					},
					Required: true,
				},
				"file_path": {
					Type:     schema.String,
					Desc:     "Path to the data file",
					Required: false,
				},
			}),
	}, nil
} // InvokableRun 执行工具
func (t *DataVisualizationTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	// 解析参数
	var args struct {
		ChartType   string   `json:"chart_type"`
		DataColumns []string `json:"data_columns"`
		FilePath    string   `json:"file_path"`
	}

	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", err
	}

	// 生成可视化代码
	code := t.generateVisualizationCode(args.ChartType, args.DataColumns, args.FilePath)

	// 执行Python代码
	result, err := t.sandbox.ExecutePython(code)
	if err != nil {
		return "", err
	}

	if !result.Success {
		return "可视化失败: " + result.Error, nil
	}

	return "可视化图表已生成，保存路径: " + result.ImagePath, nil
}

// generateVisualizationCode 生成可视化代码
func (t *DataVisualizationTool) generateVisualizationCode(chartType string, columns []string, filePath string) string {
	code := `
import pandas as pd
import matplotlib.pyplot as plt
import seaborn as sns
import numpy as np

# 读取数据
df = pd.read_csv('` + filePath + `')

# 设置图形大小
plt.figure(figsize=(10, 6))

`

	switch chartType {
	case "histogram":
		if len(columns) > 0 {
			code += `plt.hist(df['` + columns[0] + `'], bins=30, alpha=0.7)
plt.title('` + columns[0] + ` Distribution')
plt.xlabel('` + columns[0] + `')
plt.ylabel('Frequency')`
		}
	case "scatter":
		if len(columns) >= 2 {
			code += `plt.scatter(df['` + columns[0] + `'], df['` + columns[1] + `'], alpha=0.6)
plt.title('` + columns[0] + ` vs ` + columns[1] + `')
plt.xlabel('` + columns[0] + `')
plt.ylabel('` + columns[1] + `')`
		}
	case "line":
		if len(columns) > 0 {
			code += `plt.plot(df['` + columns[0] + `'])
plt.title('` + columns[0] + ` Trend')
plt.xlabel('Index')
plt.ylabel('` + columns[0] + `')`
		}
	case "bar":
		if len(columns) > 0 {
			code += `value_counts = df['` + columns[0] + `'].value_counts()
plt.bar(value_counts.index, value_counts.values)
plt.title('` + columns[0] + ` Distribution')
plt.xlabel('` + columns[0] + `')
plt.ylabel('Count')`
		}
	default:
		code += `# 默认创建数据概览图
df.hist(figsize=(12, 8))
plt.suptitle('Data Overview')`
	}

	code += `

# 保存图片
plt.tight_layout()
plt.savefig('output.png', dpi=300, bbox_inches='tight')
plt.close()

result = {'chart_type': '` + chartType + `', 'status': 'success'}
print('Chart saved successfully')
`

	return code
}

// StatisticalAnalysisTool 统计分析工具
type StatisticalAnalysisTool struct {
	sandbox *sanbox.PythonSandbox
	name    string
	desc    string
}

// NewStatisticalAnalysisTool 创建统计分析工具
func NewStatisticalAnalysisTool(sandbox *sanbox.PythonSandbox) *StatisticalAnalysisTool {
	return &StatisticalAnalysisTool{
		sandbox: sandbox,
		name:    "statistical_analysis",
		desc:    "Perform statistical analysis on data including descriptive statistics, correlation analysis, etc.",
	}
}

// Info 获取工具信息
func (t *StatisticalAnalysisTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: t.name,
		Desc: t.desc,
		ParamsOneOf: schema.NewParamsOneOfByParams(
			map[string]*schema.ParameterInfo{
				"analysis_type": {
					Type:     schema.String,
					Desc:     "Type of statistical analysis (descriptive, correlation, regression, etc.)",
					Required: true,
				},
				"columns": {
					Type: schema.Array,
					Desc: "Columns to analyze",
					ElemInfo: &schema.ParameterInfo{
						Type: schema.String,
					},
					Required: false,
				},
				"file_path": {
					Type:     schema.String,
					Desc:     "Path to the data file",
					Required: true,
				},
			}),
	}, nil
}

// InvokableRun 执行工具
func (t *StatisticalAnalysisTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	// 解析参数
	var args struct {
		AnalysisType string   `json:"analysis_type"`
		Columns      []string `json:"columns"`
		FilePath     string   `json:"file_path"`
	}

	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", err
	}

	// 生成统计分析代码
	code := t.generateStatisticalCode(args.AnalysisType, args.Columns, args.FilePath)

	// 执行Python代码
	result, err := t.sandbox.ExecutePython(code)
	if err != nil {
		return "", err
	}

	if !result.Success {
		return "统计分析失败: " + result.Error, nil
	}

	return "统计分析完成:\n" + result.Stdout, nil
}

// generateStatisticalCode 生成统计分析代码
func (t *StatisticalAnalysisTool) generateStatisticalCode(analysisType string, columns []string, filePath string) string {
	code := `
import pandas as pd
import numpy as np
from scipy import stats

# 读取数据
df = pd.read_csv('` + filePath + `')

print("数据基本信息:")
print(f"数据形状: {df.shape}")
print(f"列名: {list(df.columns)}")
print()

`

	switch analysisType {
	case "descriptive":
		code += `# 描述性统计
print("描述性统计:")
print(df.describe())
print()

print("数据类型:")
print(df.dtypes)
print()

print("缺失值统计:")
print(df.isnull().sum())
print()`

	case "correlation":
		code += `# 相关性分析
numeric_df = df.select_dtypes(include=[np.number])
if not numeric_df.empty:
    print("相关性矩阵:")
    correlation_matrix = numeric_df.corr()
    print(correlation_matrix)
    print()
    
    # 找出强相关性
    strong_corr = []
    n = len(correlation_matrix.columns)
    for i in range(n):
        for j in range(i+1, n):
            corr_value = correlation_matrix.iloc[i, j]
            if abs(corr_value) > 0.7:
                strong_corr.append({
                    'var1': correlation_matrix.columns[i],
                    'var2': correlation_matrix.columns[j],
                    'correlation': corr_value
                })
    
    if strong_corr:
        print("强相关性 (|r| > 0.7):")
        for item in strong_corr:
            print(f"{item['var1']} vs {item['var2']}: {item['correlation']:.3f}")
    else:
        print("未发现强相关性")
else:
    print("数据中没有数值型列，无法进行相关性分析")`

	case "regression":
		if len(columns) >= 2 {
			code += `# 简单线性回归分析
from sklearn.linear_model import LinearRegression
from sklearn.metrics import r2_score

X = df[['` + columns[0] + `']].dropna()
y = df['` + columns[1] + `'].dropna()

# 确保X和y长度一致
min_len = min(len(X), len(y))
X = X[:min_len]
y = y[:min_len]

if len(X) > 0:
    model = LinearRegression()
    model.fit(X, y)
    y_pred = model.predict(X)
    r2 = r2_score(y, y_pred)
    
    print(f"回归分析结果 ({columns[0]} -> {columns[1]}):")
    print(f"R² 分数: {r2:.4f}")
    print(f"回归系数: {model.coef_[0]:.4f}")
    print(f"截距: {model.intercept_:.4f}")
else:
    print("数据不足，无法进行回归分析")`
		} else {
			code += `print("回归分析需要至少两个列")`
		}

	default:
		code += `# 默认进行基础统计分析
print("基础统计信息:")
print(df.describe())
print()

print("数据类型:")
print(df.dtypes)
print()`
	}

	code += `
result = {'analysis_type': '` + analysisType + `', 'status': 'completed'}
`

	return code
}
