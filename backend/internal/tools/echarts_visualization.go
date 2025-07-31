package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"smart-analysis/internal/types"
	"smart-analysis/internal/utils/sanbox"
)

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
		desc:    "创建ECharts格式的交互式数据可视化图表，返回可在前端直接渲染的图表配置。支持柱状图、折线图、饼图、散点图、热力图等多种图表类型。",
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
					Desc:     "图表类型: bar（柱状图）, line（折线图）, pie（饼图）, scatter（散点图）, heatmap（热力图）, area（面积图）, radar（雷达图）",
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
				"custom_options": {
					Type:     schema.String,
					Desc:     "自定义ECharts配置选项（JSON格式）",
					Required: false,
				},
				"fallback_to_image": {
					Type:     schema.Boolean,
					Desc:     "如果ECharts生成失败，是否回退到静态图片生成",
					Required: false,
				},
			}),
	}, nil
}

// InvokableRun 执行工具
func (t *EChartsVisualizationTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var args struct {
		ChartType       string   `json:"chart_type"`
		DataColumns     []string `json:"data_columns"`
		Title           string   `json:"title,omitempty"`
		FilePath        string   `json:"file_path,omitempty"`
		CustomOptions   string   `json:"custom_options,omitempty"`
		FallbackToImage bool     `json:"fallback_to_image,omitempty"`
	}

	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", err
	}

	// 生成ECharts配置代码
	code := t.generateEChartsCode(args.ChartType, args.DataColumns, args.Title, args.FilePath, args.CustomOptions)

	// 执行Python代码
	result, err := t.sandbox.ExecutePython(code)
	if err != nil {
		if args.FallbackToImage {
			// 回退到静态图片生成
			return t.generateStaticImageFallback(args.ChartType, args.DataColumns, args.Title, args.FilePath)
		}
		return "", err
	}

	if !result.Success {
		if args.FallbackToImage {
			// 回退到静态图片生成
			return t.generateStaticImageFallback(args.ChartType, args.DataColumns, args.Title, args.FilePath)
		}
		return "图表生成失败: " + result.Error, nil
	}

	// 格式化结果为ECharts配置
	return t.formatEChartsResult(result), nil
}

// generateEChartsCode 生成ECharts配置的Python代码
func (t *EChartsVisualizationTool) generateEChartsCode(chartType string, columns []string, title, filePath, customOptions string) string {
	if title == "" {
		title = "数据可视化图表"
	}

	code := fmt.Sprintf(`
import pandas as pd
import numpy as np
import json

# 读取数据
%s

# 数据验证
if 'df' not in locals():
    print("ERROR: 数据未加载成功")
    exit()

# 基础图表配置
chart_config = {
    "type": "echarts",
    "chartType": "%s",
    "title": "%s",
    "data": [],
    "xAxis": [],
    "yAxis": [],
    "series": [],
    "legend": {},
    "tooltip": {
        "trigger": "axis",
        "axisPointer": {
            "type": "shadow"
        }
    },
    "grid": {
        "left": "3%%",
        "right": "4%%",
        "bottom": "3%%",
        "containLabel": True
    }
}

`, t.getDataLoadCode(filePath), chartType, title)

	switch chartType {
	case "bar", "line":
		code += t.generateBarLineChartCode(columns, chartType)
	case "pie":
		code += t.generatePieChartCode(columns)
	case "scatter":
		code += t.generateScatterChartCode(columns)
	case "heatmap":
		code += t.generateHeatmapChartCode(columns)
	case "area":
		code += t.generateAreaChartCode(columns)
	case "radar":
		code += t.generateRadarChartCode(columns)
	default:
		code += t.generateBarLineChartCode(columns, "bar") // 默认柱状图
	}

	// 添加自定义选项
	if customOptions != "" {
		code += fmt.Sprintf(`
# 合并自定义配置
try:
    custom_opts = json.loads('''%s''')
    chart_config.update(custom_opts)
except Exception as e:
    print(f"自定义配置解析失败: {e}")

`, customOptions)
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
func (t *EChartsVisualizationTool) generateBarLineChartCode(columns []string, chartType string) string {
	if len(columns) < 2 {
		return `
# 数据列不足
chart_config["data"] = []
print("WARNING: 需要至少两列数据用于柱状图/折线图")
`
	}

	return fmt.Sprintf(`
# 柱状图/折线图配置
x_col = "%s"
y_col = "%s"

if x_col in df.columns and y_col in df.columns:
    # 数据预处理
    plot_data = df[[x_col, y_col]].dropna()
    
    chart_config["xAxis"] = {
        "type": "category",
        "data": plot_data[x_col].astype(str).tolist()
    }
    chart_config["yAxis"] = {
        "type": "value"
    }
    chart_config["series"] = [{
        "name": y_col,
        "type": "%s",
        "data": plot_data[y_col].tolist(),
        "itemStyle": {
            "color": "#5470c6"
        }
    }]
    chart_config["legend"] = {
        "data": [y_col]
    }
else:
    print(f"ERROR: 列 {x_col} 或 {y_col} 在数据中不存在")
`, columns[0], columns[1], chartType)
}

// generatePieChartCode 生成饼图代码
func (t *EChartsVisualizationTool) generatePieChartCode(columns []string) string {
	if len(columns) < 2 {
		return `
# 数据列不足
chart_config["data"] = []
print("WARNING: 需要至少两列数据用于饼图")
`
	}

	return fmt.Sprintf(`
# 饼图配置
name_col = "%s"
value_col = "%s"

if name_col in df.columns and value_col in df.columns:
    # 数据预处理和聚合
    pie_data = df.groupby(name_col)[value_col].sum().reset_index()
    
    chart_config["series"] = [{
        "name": "数据分布",
        "type": "pie",
        "radius": ["40%%", "70%%"],
        "avoidLabelOverlap": False,
        "data": [
            {"name": str(name), "value": float(value)} 
            for name, value in zip(pie_data[name_col], pie_data[value_col])
            if pd.notna(name) and pd.notna(value) and value > 0
        ],
        "emphasis": {
            "itemStyle": {
                "shadowBlur": 10,
                "shadowOffsetX": 0,
                "shadowColor": "rgba(0, 0, 0, 0.5)"
            }
        }
    }]
    chart_config["tooltip"] = {
        "trigger": "item",
        "formatter": "{a} <br/>{b}: {c} ({d}%%)"
    }
else:
    print(f"ERROR: 列 {name_col} 或 {value_col} 在数据中不存在")
`, columns[0], columns[1])
}

// generateScatterChartCode 生成散点图代码
func (t *EChartsVisualizationTool) generateScatterChartCode(columns []string) string {
	if len(columns) < 2 {
		return `
# 数据列不足
chart_config["data"] = []
print("WARNING: 需要至少两列数据用于散点图")
`
	}

	return fmt.Sprintf(`
# 散点图配置
x_col = "%s"
y_col = "%s"

if x_col in df.columns and y_col in df.columns:
    # 数据预处理
    scatter_data = df[[x_col, y_col]].dropna()
    
    chart_config["xAxis"] = {
        "type": "value",
        "name": x_col
    }
    chart_config["yAxis"] = {
        "type": "value", 
        "name": y_col
    }
    chart_config["series"] = [{
        "name": f"{x_col} vs {y_col}",
        "type": "scatter",
        "data": [[float(x), float(y)] for x, y in zip(scatter_data[x_col], scatter_data[y_col])],
        "symbolSize": 8,
        "itemStyle": {
            "color": "#ff7f50"
        }
    }]
    chart_config["tooltip"] = {
        "trigger": "item",
        "formatter": f"{x_col}: {{c[0]}}<br/>{y_col}: {{c[1]}}"
    }
else:
    print(f"ERROR: 列 {x_col} 或 {y_col} 在数据中不存在")
`, columns[0], columns[1])
}

// generateAreaChartCode 生成面积图代码
func (t *EChartsVisualizationTool) generateAreaChartCode(columns []string) string {
	if len(columns) < 2 {
		return `
# 数据列不足
chart_config["data"] = []
print("WARNING: 需要至少两列数据用于面积图")
`
	}

	return fmt.Sprintf(`
# 面积图配置
x_col = "%s"
y_col = "%s"

if x_col in df.columns and y_col in df.columns:
    # 数据预处理
    area_data = df[[x_col, y_col]].dropna().sort_values(x_col)
    
    chart_config["xAxis"] = {
        "type": "category",
        "data": area_data[x_col].astype(str).tolist()
    }
    chart_config["yAxis"] = {
        "type": "value"
    }
    chart_config["series"] = [{
        "name": y_col,
        "type": "line",
        "data": area_data[y_col].tolist(),
        "areaStyle": {},
        "smooth": True,
        "itemStyle": {
            "color": "#91cc75"
        }
    }]
else:
    print(f"ERROR: 列 {x_col} 或 {y_col} 在数据中不存在")
`, columns[0], columns[1])
}

// generateRadarChartCode 生成雷达图代码
func (t *EChartsVisualizationTool) generateRadarChartCode(columns []string) string {
	if len(columns) < 3 {
		return `
# 数据列不足
chart_config["data"] = []
print("WARNING: 雷达图需要至少3列数据")
`
	}

	return fmt.Sprintf(`
# 雷达图配置
indicator_cols = %s

# 验证列是否存在
missing_cols = [col for col in indicator_cols if col not in df.columns]
if missing_cols:
    print(f"ERROR: 列 {missing_cols} 在数据中不存在")
else:
    # 数据预处理
    radar_data = df[indicator_cols].dropna()
    
    # 计算每个指标的最大值用于雷达图范围
    max_values = radar_data.max()
    
    chart_config["radar"] = {
        "indicator": [
            {"name": col, "max": float(max_values[col]) * 1.2}
            for col in indicator_cols
        ]
    }
    
    chart_config["series"] = [{
        "name": "雷达图数据",
        "type": "radar",
        "data": [
            {
                "value": row.tolist(),
                "name": f"样本 {idx+1}"
            }
            for idx, (_, row) in enumerate(radar_data.head(5).iterrows())
        ]
    }]
`, fmt.Sprintf("%q", columns))
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
	var config types.EChartsConfig
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		return "图表配置JSON格式错误: " + err.Error()
	}

	return fmt.Sprintf("ECharts图表配置生成成功:\n```json\n%s\n```", configJSON)
}

// generateStaticImageFallback 生成静态图片作为回退方案
func (t *EChartsVisualizationTool) generateStaticImageFallback(chartType string, columns []string, title, filePath string) (string, error) {
	if title == "" {
		title = "数据可视化图表"
	}

	code := fmt.Sprintf(`
import pandas as pd
import numpy as np
import matplotlib.pyplot as plt
import seaborn as sns

# 设置中文字体
plt.rcParams['font.sans-serif'] = ['SimHei', 'Arial Unicode MS', 'DejaVu Sans']
plt.rcParams['axes.unicode_minus'] = False

# 读取数据
%s

if 'df' not in locals():
    print("ERROR: 数据未加载成功")
    exit()

# 创建图表
plt.figure(figsize=(10, 6))
plt.title("%s")

`, t.getDataLoadCode(filePath), title)

	switch chartType {
	case "bar":
		if len(columns) >= 2 {
			code += fmt.Sprintf(`
if "%s" in df.columns and "%s" in df.columns:
    plt.bar(df["%s"], df["%s"])
    plt.xlabel("%s")
    plt.ylabel("%s")
`, columns[0], columns[1], columns[0], columns[1], columns[0], columns[1])
		}
	case "line":
		if len(columns) >= 2 {
			code += fmt.Sprintf(`
if "%s" in df.columns and "%s" in df.columns:
    plt.plot(df["%s"], df["%s"], marker='o')
    plt.xlabel("%s")
    plt.ylabel("%s")
`, columns[0], columns[1], columns[0], columns[1], columns[0], columns[1])
		}
	case "scatter":
		if len(columns) >= 2 {
			code += fmt.Sprintf(`
if "%s" in df.columns and "%s" in df.columns:
    plt.scatter(df["%s"], df["%s"])
    plt.xlabel("%s")
    plt.ylabel("%s")
`, columns[0], columns[1], columns[0], columns[1], columns[0], columns[1])
		}
	case "pie":
		if len(columns) >= 2 {
			code += fmt.Sprintf(`
if "%s" in df.columns and "%s" in df.columns:
    pie_data = df.groupby("%s")["%s"].sum()
    plt.pie(pie_data.values, labels=pie_data.index, autopct='%%1.1f%%%%')
`, columns[0], columns[1], columns[0], columns[1])
		}
	}

	code += `
plt.tight_layout()
plt.savefig('output.png', dpi=300, bbox_inches='tight')
plt.close()
print("静态图片生成完成: output.png")
`

	result, err := t.sandbox.ExecutePython(code)
	if err != nil {
		return "", err
	}

	if !result.Success {
		return "静态图片生成失败: " + result.Error, nil
	}

	response := "ECharts生成失败，已生成静态图片作为替代:\n"
	if result.Stdout != "" {
		response += result.Stdout + "\n"
	}
	if result.ImagePath != "" {
		response += "生成的图片路径: " + result.ImagePath
	}

	return response, nil
}
