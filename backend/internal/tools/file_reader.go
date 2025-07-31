package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"smart-analysis/internal/utils/sanbox"
)

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
