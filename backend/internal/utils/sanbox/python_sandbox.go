// Package sanbox provides a sandbox environment for executing Python code safely.
package sanbox

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// PythonExecutionResult Python执行结果
type PythonExecutionResult struct {
	Success    bool        `json:"success"`
	Output     interface{} `json:"output"`      // 可能是字符串、数字、字典等
	OutputType string      `json:"output_type"` // "text", "dataframe", "image", "number", "dict", "list"
	Error      string      `json:"error,omitempty"`
	ImagePath  string      `json:"image_path,omitempty"` // 如果输出是图片，返回图片路径
	Stdout     string      `json:"stdout,omitempty"`     // 标准输出
	Stderr     string      `json:"stderr,omitempty"`     // 标准错误输出
}

// PythonSandbox Python代码执行沙箱
type PythonSandbox struct {
	timeout    time.Duration
	uploadDir  string // 文件上传目录，用于保存图片等文件
	pythonPath string // Python解释器路径
}

// NewPythonSandbox 创建新的Python沙箱
func NewPythonSandbox(uploadDir string) *PythonSandbox {
	return &PythonSandbox{
		timeout:    30 * time.Second, // 默认30秒超时
		uploadDir:  uploadDir,
		pythonPath: "/Users/wanghao/Desktop/github/go/smart-analysis/.venv/bin/python",
	}
}

// SetTimeout 设置执行超时时间
func (ps *PythonSandbox) SetTimeout(timeout time.Duration) {
	ps.timeout = timeout
}

// SetPythonPath 设置Python解释器路径
func (ps *PythonSandbox) SetPythonPath(path string) {
	ps.pythonPath = path
}

// ExecuteCode 执行Python代码（主要API）
func (ps *PythonSandbox) ExecuteCode(code string) (*PythonExecutionResult, error) {
	return ps.execute(code)
}

// ExecutePython 执行Python代码（别名，为了兼容性）
func (ps *PythonSandbox) ExecutePython(code string) (*PythonExecutionResult, error) {
	return ps.execute(code)
}

// execute 内部执行方法
func (ps *PythonSandbox) execute(code string) (*PythonExecutionResult, error) {
	// 确保上传目录存在
	if ps.uploadDir != "" {
		os.MkdirAll(ps.uploadDir, 0755)
	}

	// 创建临时目录
	tempDir, err := ioutil.TempDir("", "python_sandbox_*")
	if err != nil {
		return nil, fmt.Errorf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建Python执行脚本
	pythonScript := ps.createExecutionScript(code, tempDir)

	// 写入Python脚本文件
	scriptPath := filepath.Join(tempDir, "execute.py")
	err = ioutil.WriteFile(scriptPath, []byte(pythonScript), 0644)
	if err != nil {
		return nil, fmt.Errorf("写入Python脚本失败: %v", err)
	}

	// 执行Python脚本
	ctx, cancel := context.WithTimeout(context.Background(), ps.timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, ps.pythonPath, scriptPath)
	cmd.Dir = tempDir

	stdout, stderr, err := ps.runCommand(cmd)

	// 解析结果
	result := &PythonExecutionResult{}

	if err != nil {
		result.Success = false
		result.Error = err.Error()
		result.Stderr = stderr
		result.Stdout = stdout
		return result, nil
	}

	// 尝试解析输出结果
	err = ps.parseResult(stdout, stderr, tempDir, result)
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("解析结果失败: %v", err)
		result.Stdout = stdout
		result.Stderr = stderr
	}

	return result, nil
}

// createExecutionScript 创建Python执行脚本
func (ps *PythonSandbox) createExecutionScript(userCode, tempDir string) string {
	return fmt.Sprintf(`
import sys
import json
import traceback
import os
from io import StringIO
import pandas as pd
import matplotlib
matplotlib.use('Agg')
import matplotlib.pyplot as plt
import numpy as np

def safe_serialize(obj):
    if obj is None:
        return {"type": "none", "value": None}
    elif isinstance(obj, str):
        return {"type": "text", "value": obj}
    elif isinstance(obj, (int, float)):
        return {"type": "number", "value": obj}
    elif isinstance(obj, bool):
        return {"type": "boolean", "value": obj}
    elif isinstance(obj, dict):
        try:
            json.dumps(obj)
            return {"type": "dict", "value": obj}
        except:
            return {"type": "text", "value": str(obj)}
    elif isinstance(obj, (list, tuple)):
        try:
            clean_list = []
            for item in obj:
                if isinstance(item, (str, int, float, bool, type(None))):
                    clean_list.append(item)
                else:
                    clean_list.append(str(item))
            return {"type": "list", "value": clean_list}
        except:
            return {"type": "text", "value": str(obj)}
    elif isinstance(obj, pd.DataFrame):
        try:
            return {
                "type": "dataframe", 
                "value": {
                    "columns": obj.columns.tolist(),
                    "data": obj.values.tolist(),
                    "shape": obj.shape,
                    "index": obj.index.tolist(),
                    "dtypes": {col: str(dtype) for col, dtype in obj.dtypes.items()}
                }
            }
        except:
            return {"type": "text", "value": str(obj)}
    else:
        return {"type": "text", "value": str(obj)}

def check_for_plots():
    try:
        if plt.get_fignums():
            image_path = os.path.join("%s", "output_plot.png")
            plt.savefig(image_path, dpi=150, bbox_inches='tight')
            plt.close('all')
            return image_path
    except:
        pass
    return None

old_stdout = sys.stdout
sys.stdout = captured_output = StringIO()

result_obj = None

try:
    exec_globals = {"__name__": "__main__"}
    exec_locals = {}
    
    code_lines = '''%s'''.strip().split('\n')
    
    if code_lines:
        last_line = code_lines[-1].strip()
        other_lines = code_lines[:-1]
        
        if other_lines:
            exec('\n'.join(other_lines), exec_globals, exec_locals)
        
        try:
            if not (last_line.startswith(('print', 'plt.', 'import', 'from', 'raise')) or 
                   last_line.endswith((':')) or 
                   any(keyword in last_line for keyword in ['=', 'if ', 'for ', 'while ', 'def ', 'class '])):
                result_obj = eval(last_line, exec_globals, exec_locals)
            else:
                exec(last_line, exec_globals, exec_locals)
        except Exception as e:
            if isinstance(e, (ValueError, NameError, TypeError, AttributeError, ImportError, ZeroDivisionError, SyntaxError)):
                raise e
            try:
                exec(last_line, exec_globals, exec_locals)
            except:
                pass
    else:
        exec('''%s''', exec_globals, exec_locals)
    
    print_output = captured_output.getvalue()
    image_path = check_for_plots()
    
    if result_obj is not None:
        serialized = safe_serialize(result_obj)
    elif print_output.strip():
        serialized = {"type": "text", "value": print_output.strip()}
    else:
        serialized = {"type": "none", "value": None}
    
    output = {
        "success": True,
        "result": serialized,
        "stdout": print_output,
        "image_path": image_path
    }
    
    print(json.dumps(output), file=sys.__stdout__)
    
except Exception as e:
    sys.stdout = old_stdout
    
    error_output = {
        "success": False,
        "error": str(e),
        "traceback": traceback.format_exc(),
        "stdout": captured_output.getvalue()
    }
    
    print(json.dumps(error_output), file=sys.__stdout__)
finally:
    sys.stdout = old_stdout
`, tempDir, userCode, userCode)
}

// runCommand 运行命令并获取输出
func (ps *PythonSandbox) runCommand(cmd *exec.Cmd) (string, string, error) {
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	return stdout.String(), stderr.String(), err
}

// parseResult 解析Python执行结果
func (ps *PythonSandbox) parseResult(stdout, stderr, tempDir string, result *PythonExecutionResult) error {
	var pythonResult map[string]interface{}

	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	var jsonLine string

	for i := len(lines) - 1; i >= 0; i-- {
		if strings.HasPrefix(lines[i], "{") {
			jsonLine = lines[i]
			break
		}
	}

	if jsonLine == "" {
		return fmt.Errorf("未找到JSON输出")
	}

	err := json.Unmarshal([]byte(jsonLine), &pythonResult)
	if err != nil {
		return fmt.Errorf("解析JSON输出失败: %v", err)
	}

	if success, ok := pythonResult["success"].(bool); ok {
		result.Success = success
	}

	if !result.Success {
		if errorMsg, ok := pythonResult["error"].(string); ok {
			result.Error = errorMsg
		}
		if traceback, ok := pythonResult["traceback"].(string); ok {
			result.Error += "\n" + traceback
		}
		if stdoutVal, ok := pythonResult["stdout"].(string); ok {
			result.Stdout = stdoutVal
		}
		return nil
	}

	if resultData, ok := pythonResult["result"].(map[string]interface{}); ok {
		if resultType, ok := resultData["type"].(string); ok {
			result.OutputType = resultType
			result.Output = resultData["value"]
		}
	}

	if stdoutVal, ok := pythonResult["stdout"].(string); ok {
		result.Stdout = stdoutVal
	}

	if imagePath, ok := pythonResult["image_path"].(string); ok && imagePath != "" {
		if ps.uploadDir != "" {
			destPath := filepath.Join(ps.uploadDir, fmt.Sprintf("python_plot_%d.png", time.Now().Unix()))
			err = ps.moveFile(imagePath, destPath)
			if err == nil {
				result.ImagePath = destPath
				if result.OutputType == "none" || result.OutputType == "" {
					result.OutputType = "image"
				}
			}
		}
	}

	result.Success = true
	return nil
}

// moveFile 移动文件
func (ps *PythonSandbox) moveFile(src, dst string) error {
	err := os.MkdirAll(filepath.Dir(dst), 0755)
	if err != nil {
		return err
	}

	data, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(dst, data, 0644)
}

// InstallRequiredPackages 安装必需的Python包
func (ps *PythonSandbox) InstallRequiredPackages() error {
	packages := []string{
		"pandas",
		"matplotlib",
		"numpy",
		"seaborn",
		"scipy",
	}

	for _, pkg := range packages {
		cmd := exec.Command(ps.pythonPath, "-m", "pip", "install", pkg)
		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("安装包 %s 失败: %v", pkg, err)
		}
	}

	return nil
}
