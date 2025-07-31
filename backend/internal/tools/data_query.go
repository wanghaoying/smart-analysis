package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"smart-analysis/internal/utils/sanbox"
)

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
