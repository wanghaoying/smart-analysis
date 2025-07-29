package sanbox

import (
	"testing"
	"time"
)

func TestPythonSandbox_Basic(t *testing.T) {
	// 创建沙箱
	uploadDir := "/tmp/test_uploads"
	sandbox := NewPythonSandbox(uploadDir)
	sandbox.SetTimeout(10 * time.Second)

	// 测试基本文本输出
	code := `print("Hello, World!")`
	result, err := sandbox.ExecuteCode(code)

	if err != nil {
		t.Fatalf("执行代码失败: %v", err)
	}

	if !result.Success {
		t.Fatalf("代码执行失败: %s", result.Error)
	}

	if result.OutputType != "text" {
		t.Errorf("期望输出类型为 text，实际为 %s", result.OutputType)
	}
}

func TestPythonSandbox_DataFrame(t *testing.T) {
	uploadDir := "/tmp/test_uploads"
	sandbox := NewPythonSandbox(uploadDir)

	code := `
import pandas as pd
df = pd.DataFrame({'A': [1, 2, 3], 'B': [4, 5, 6]})
df
`
	result, err := sandbox.ExecuteCode(code)

	if err != nil {
		t.Fatalf("执行代码失败: %v", err)
	}

	if !result.Success {
		t.Fatalf("代码执行失败: %s", result.Error)
	}

	if result.OutputType != "dataframe" {
		t.Errorf("期望输出类型为 dataframe，实际为 %s", result.OutputType)
	}
}

func TestPythonSandbox_Plot(t *testing.T) {
	uploadDir := "/tmp/test_uploads"
	sandbox := NewPythonSandbox(uploadDir)

	code := `
import matplotlib.pyplot as plt
import numpy as np

x = np.linspace(0, 10, 100)
y = np.sin(x)

plt.figure(figsize=(8, 6))
plt.plot(x, y)
plt.title("Sine Wave")
plt.xlabel("X")
plt.ylabel("Y")
`
	result, err := sandbox.ExecuteCode(code)

	if err != nil {
		t.Fatalf("执行代码失败: %v", err)
	}

	if !result.Success {
		t.Fatalf("代码执行失败: %s", result.Error)
	}

	if result.ImagePath == "" {
		t.Error("期望生成图片，但没有图片路径")
	}
}

func TestPythonSandbox_Dictionary(t *testing.T) {
	uploadDir := "/tmp/test_uploads"
	sandbox := NewPythonSandbox(uploadDir)

	code := `
data = {
    'name': 'John',
    'age': 30,
    'city': 'New York',
    'scores': [85, 90, 78]
}
data
`
	result, err := sandbox.ExecuteCode(code)

	if err != nil {
		t.Fatalf("执行代码失败: %v", err)
	}

	if !result.Success {
		t.Fatalf("代码执行失败: %s", result.Error)
	}

	if result.OutputType != "dict" {
		t.Errorf("期望输出类型为 dict，实际为 %s", result.OutputType)
	}
}

func TestPythonSandbox_Error(t *testing.T) {
	uploadDir := "/tmp/test_uploads"
	sandbox := NewPythonSandbox(uploadDir)

	code := `
# 这段代码会出错
raise ValueError("This is a test error")
`
	result, err := sandbox.ExecuteCode(code)

	if err != nil {
		t.Fatalf("执行代码失败: %v", err)
	}

	if result.Success {
		t.Error("期望代码执行失败，但实际成功了")
	}

	if result.Error == "" {
		t.Error("期望有错误信息，但没有")
	}
}
