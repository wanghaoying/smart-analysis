package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"smart-analysis/internal/utils/sanbox"
)

// ReportGeneratorTool 自动报告生成工具
type ReportGeneratorTool struct {
	sandbox *sanbox.PythonSandbox
	name    string
	desc    string
}

// NewReportGeneratorTool 创建报告生成工具
func NewReportGeneratorTool(sandbox *sanbox.PythonSandbox) *ReportGeneratorTool {
	return &ReportGeneratorTool{
		sandbox: sandbox,
		name:    "report_generator",
		desc:    "自动数据分析报告生成工具，可以生成包含数据概览、统计分析、可视化图表的完整分析报告。",
	}
}

// Info 返回工具信息
func (t *ReportGeneratorTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: t.name,
		Desc: t.desc,
		ParamsOneOf: schema.NewParamsOneOfByParams(
			map[string]*schema.ParameterInfo{
				"report_type": {
					Type:     schema.String,
					Desc:     "报告类型: overview（数据概览）, statistical（统计分析）, comprehensive（综合分析）",
					Required: true,
				},
				"file_path": {
					Type:     schema.String,
					Desc:     "数据文件路径",
					Required: true,
				},
				"target_column": {
					Type:     schema.String,
					Desc:     "目标分析列（可选）",
					Required: false,
				},
				"include_charts": {
					Type:     schema.Boolean,
					Desc:     "是否包含图表（默认true）",
					Required: false,
				},
				"output_format": {
					Type:     schema.String,
					Desc:     "输出格式: markdown（默认）, html, json",
					Required: false,
				},
			}),
	}, nil
}

// InvokableRun 执行工具
func (t *ReportGeneratorTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var args struct {
		ReportType    string `json:"report_type"`
		FilePath      string `json:"file_path"`
		TargetColumn  string `json:"target_column,omitempty"`
		IncludeCharts bool   `json:"include_charts,omitempty"`
		OutputFormat  string `json:"output_format,omitempty"`
	}

	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", err
	}

	// 设置默认值
	if args.OutputFormat == "" {
		args.OutputFormat = "markdown"
	}
	args.IncludeCharts = true // 默认包含图表

	code := t.generateReportCode(args.ReportType, args.FilePath, args.TargetColumn, args.IncludeCharts, args.OutputFormat)

	result, err := t.sandbox.ExecutePython(code)
	if err != nil {
		return "", err
	}

	if !result.Success {
		return "报告生成失败: " + result.Error, nil
	}

	return result.Stdout, nil
}

// generateReportCode 生成报告生成代码
func (t *ReportGeneratorTool) generateReportCode(reportType, filePath, targetColumn string, includeCharts bool, outputFormat string) string {
	currentTime := time.Now().Format("2006-01-02 15:04:05")

	code := fmt.Sprintf(`
import pandas as pd
import numpy as np
import json
from datetime import datetime

# 数据加载
try:
    if '%s'.endswith('.csv'):
        df = pd.read_csv('%s')
    elif '%s'.endswith(('.xlsx', '.xls')):
        df = pd.read_excel('%s')
    elif '%s'.endswith('.json'):
        df = pd.read_json('%s')
    else:
        print("不支持的文件格式")
        exit()
except Exception as e:
    print(f"数据加载失败: {e}")
    exit()

# 生成报告
report_data = {
    "title": "数据分析报告",
    "generated_at": "%s",
    "file_path": "%s",
    "report_type": "%s",
    "data_summary": {},
    "analysis_results": {},
    "charts": [] if %t else None
}

`, filePath, filePath, filePath, filePath, filePath, filePath, currentTime, filePath, reportType, includeCharts)

	// 添加基础数据分析
	code += `
# 基础数据概览
report_data["data_summary"] = {
    "shape": df.shape,
    "columns": df.columns.tolist(),
    "dtypes": df.dtypes.astype(str).to_dict(),
    "missing_values": df.isnull().sum().to_dict(),
    "memory_usage": f"{df.memory_usage(deep=True).sum() / 1024 / 1024:.2f} MB"
}

# 数值型列统计
numeric_cols = df.select_dtypes(include=[np.number]).columns.tolist()
if numeric_cols:
    report_data["analysis_results"]["numeric_summary"] = df[numeric_cols].describe().to_dict()

# 类别型列统计
categorical_cols = df.select_dtypes(include=['object']).columns.tolist()
if categorical_cols:
    categorical_summary = {}
    for col in categorical_cols[:5]:  # 限制前5个类别列
        categorical_summary[col] = {
            "unique_count": df[col].nunique(),
            "top_values": df[col].value_counts().head(10).to_dict()
        }
    report_data["analysis_results"]["categorical_summary"] = categorical_summary

`

	// 根据报告类型添加特定分析
	switch reportType {
	case "statistical":
		code += t.generateStatisticalAnalysisCode(targetColumn)
	case "comprehensive":
		code += t.generateComprehensiveAnalysisCode(targetColumn)
	default: // overview
		code += t.generateOverviewAnalysisCode()
	}

	// 添加图表生成（如果需要）
	if includeCharts {
		code += t.generateChartCode()
	}

	// 根据输出格式生成最终报告
	switch outputFormat {
	case "html":
		code += t.generateHTMLOutput()
	case "json":
		code += t.generateJSONOutput()
	default: // markdown
		code += t.generateMarkdownOutput()
	}

	return code
}

// generateStatisticalAnalysisCode 生成统计分析代码
func (t *ReportGeneratorTool) generateStatisticalAnalysisCode(targetColumn string) string {
	code := `
# 统计分析
from scipy import stats

# 相关性分析
if len(numeric_cols) > 1:
    correlation_matrix = df[numeric_cols].corr()
    report_data["analysis_results"]["correlation_analysis"] = correlation_matrix.to_dict()

# 分布分析
distribution_analysis = {}
for col in numeric_cols[:5]:  # 限制前5个数值列
    data = df[col].dropna()
    if len(data) > 10:
        # 正态性检验
        statistic, p_value = stats.normaltest(data)
        distribution_analysis[col] = {
            "skewness": float(stats.skew(data)),
            "kurtosis": float(stats.kurtosis(data)),
            "normality_test": {
                "statistic": float(statistic),
                "p_value": float(p_value),
                "is_normal": p_value > 0.05
            }
        }

report_data["analysis_results"]["distribution_analysis"] = distribution_analysis

`

	if targetColumn != "" {
		code += fmt.Sprintf(`
# 目标变量分析
if '%s' in df.columns:
    target_data = df['%s'].dropna()
    target_analysis = {
        "column": '%s',
        "data_type": str(df['%s'].dtype),
        "unique_values": int(df['%s'].nunique()),
    }
    
    if df['%s'].dtype in ['int64', 'float64']:
        target_analysis.update({
            "mean": float(target_data.mean()),
            "median": float(target_data.median()),
            "std": float(target_data.std()),
            "min": float(target_data.min()),
            "max": float(target_data.max())
        })
    else:
        target_analysis.update({
            "value_counts": target_data.value_counts().head(10).to_dict()
        })
    
    report_data["analysis_results"]["target_analysis"] = target_analysis

`, targetColumn, targetColumn, targetColumn, targetColumn, targetColumn, targetColumn)
	}

	return code
}

// generateComprehensiveAnalysisCode 生成综合分析代码
func (t *ReportGeneratorTool) generateComprehensiveAnalysisCode(targetColumn string) string {
	code := t.generateStatisticalAnalysisCode(targetColumn)

	code += `
# 异常值检测
outlier_analysis = {}
for col in numeric_cols[:3]:  # 限制前3个数值列
    Q1 = df[col].quantile(0.25)
    Q3 = df[col].quantile(0.75)
    IQR = Q3 - Q1
    lower_bound = Q1 - 1.5 * IQR
    upper_bound = Q3 + 1.5 * IQR
    
    outliers = df[(df[col] < lower_bound) | (df[col] > upper_bound)]
    outlier_analysis[col] = {
        "outlier_count": len(outliers),
        "outlier_percentage": round(len(outliers) / len(df) * 100, 2),
        "bounds": {"lower": float(lower_bound), "upper": float(upper_bound)}
    }

report_data["analysis_results"]["outlier_analysis"] = outlier_analysis

# 数据质量评估
data_quality = {
    "completeness": round((1 - df.isnull().sum().sum() / (df.shape[0] * df.shape[1])) * 100, 2),
    "duplicate_rows": int(df.duplicated().sum()),
    "duplicate_percentage": round(df.duplicated().sum() / len(df) * 100, 2)
}

report_data["analysis_results"]["data_quality"] = data_quality

`

	return code
}

// generateOverviewAnalysisCode 生成概览分析代码
func (t *ReportGeneratorTool) generateOverviewAnalysisCode() string {
	return `
# 基础概览分析已在前面完成
# 添加简单的数据样本
report_data["analysis_results"]["data_sample"] = df.head(5).to_dict("records")

`
}

// generateChartCode 生成图表代码
func (t *ReportGeneratorTool) generateChartCode() string {
	return `
# 生成图表配置
charts = []

# 数值型变量的分布图
for col in numeric_cols[:3]:
    chart_config = {
        "type": "histogram",
        "title": f"{col} 分布图",
        "column": col,
        "data": df[col].dropna().tolist()
    }
    charts.append(chart_config)

# 类别型变量的柱状图
for col in categorical_cols[:2]:
    top_values = df[col].value_counts().head(10)
    chart_config = {
        "type": "bar",
        "title": f"{col} 分布图",
        "column": col,
        "data": {"categories": top_values.index.tolist(), "values": top_values.values.tolist()}
    }
    charts.append(chart_config)

# 相关性热力图（如果有足够的数值列）
if len(numeric_cols) > 2:
    corr_matrix = df[numeric_cols].corr()
    chart_config = {
        "type": "heatmap",
        "title": "相关性热力图",
        "data": corr_matrix.to_dict()
    }
    charts.append(chart_config)

report_data["charts"] = charts

`
}

// generateMarkdownOutput 生成Markdown格式输出
func (t *ReportGeneratorTool) generateMarkdownOutput() string {
	return `
# 生成Markdown报告
def generate_markdown_report(data):
    md_content = f"""# {data['title']}

**生成时间**: {data['generated_at']}  
**数据文件**: {data['file_path']}  
**报告类型**: {data['report_type']}

## 数据概览

- **数据形状**: {data['data_summary']['shape'][0]} 行 × {data['data_summary']['shape'][1]} 列
- **内存使用**: {data['data_summary']['memory_usage']}
- **列类型**: {len(data['data_summary']['columns'])} 个列

### 列信息
"""

    # 添加列信息
    for col, dtype in data['data_summary']['dtypes'].items():
        missing = data['data_summary']['missing_values'][col]
        md_content += f"- **{col}** ({dtype}): {missing} 个缺失值\\n"

    # 添加分析结果
    if 'numeric_summary' in data['analysis_results']:
        md_content += "\\n## 数值型变量统计\\n\\n"
        for col, stats in data['analysis_results']['numeric_summary'].items():
            md_content += f"### {col}\\n"
            md_content += f"- 平均值: {stats['mean']:.2f}\\n"
            md_content += f"- 中位数: {stats['50%']:.2f}\\n"
            md_content += f"- 标准差: {stats['std']:.2f}\\n"
            md_content += f"- 最小值: {stats['min']:.2f}\\n"
            md_content += f"- 最大值: {stats['max']:.2f}\\n\\n"

    if 'categorical_summary' in data['analysis_results']:
        md_content += "\\n## 类别型变量统计\\n\\n"
        for col, stats in data['analysis_results']['categorical_summary'].items():
            md_content += f"### {col}\\n"
            md_content += f"- 唯一值数量: {stats['unique_count']}\\n"
            md_content += "- 频率分布:\\n"
            for value, count in stats['top_values'].items():
                md_content += f"  - {value}: {count}\\n"

    # 添加图表信息
    if data.get('charts'):
        md_content += "\\n## 可视化图表\\n\\n"
        for i, chart in enumerate(data['charts']):
            md_content += f"{i+1}. **{chart['title']}** ({chart['type']})\\n"

    return md_content

# 输出Markdown报告
markdown_report = generate_markdown_report(report_data)
print(markdown_report)
`
}

// generateHTMLOutput 生成HTML格式输出
func (t *ReportGeneratorTool) generateHTMLOutput() string {
	return `
# 生成HTML报告
def generate_html_report(data):
    html_content = f"""
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{data['title']}</title>
    <style>
        body {{ font-family: Arial, sans-serif; margin: 20px; }}
        .header {{ background: #f0f0f0; padding: 15px; border-radius: 5px; }}
        .section {{ margin: 20px 0; }}
        table {{ border-collapse: collapse; width: 100%; }}
        th, td {{ border: 1px solid #ddd; padding: 8px; text-align: left; }}
        th {{ background-color: #f2f2f2; }}
    </style>
</head>
<body>
    <div class="header">
        <h1>{data['title']}</h1>
        <p><strong>生成时间:</strong> {data['generated_at']}</p>
        <p><strong>数据文件:</strong> {data['file_path']}</p>
        <p><strong>报告类型:</strong> {data['report_type']}</p>
    </div>
    
    <div class="section">
        <h2>数据概览</h2>
        <ul>
            <li>数据形状: {data['data_summary']['shape'][0]} 行 × {data['data_summary']['shape'][1]} 列</li>
            <li>内存使用: {data['data_summary']['memory_usage']}</li>
        </ul>
    </div>
"""

    # 添加更多HTML内容...
    html_content += "</body></html>"
    return html_content

# 输出HTML报告
html_report = generate_html_report(report_data)
print(html_report)
`
}

// generateJSONOutput 生成JSON格式输出
func (t *ReportGeneratorTool) generateJSONOutput() string {
	return `
# 输出JSON格式报告
print(json.dumps(report_data, ensure_ascii=False, indent=2))
`
}
