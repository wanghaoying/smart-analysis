# Python代码执行沙箱工具

这是一个在Go项目中安全执行Python代码的统一工具，支持多种数据类型的返回和处理。

## 🚀 重构说明

**已优化**: 原来的 `python_executor.go` 和 `python_sandbox.go` 两个文件功能重复，现在已合并为一个统一的 `PythonSandbox` 工具。

## 功能特性

- ✅ **统一API**: 提供 `ExecuteCode()` 和 `ExecutePython()` 两个方法（兼容性）
- ✅ **安全沙箱执行**: 使用临时目录隔离执行环境
- ✅ **多种数据类型支持**: 文本、数字、字典、列表、DataFrame、图片等
- ✅ **图表生成**: 自动保存matplotlib图表到指定目录
- ✅ **错误处理**: 完善的异常捕获和错误信息返回
- ✅ **超时控制**: 防止代码无限执行，可自定义
- ✅ **输出捕获**: 捕获print语句和标准输出
- ✅ **配置灵活**: 可设置Python路径、超时时间等

## 使用方法

### 基本使用

```go
import "smart-analysis/internal/utils/sanbox"

// 创建沙箱
sandbox := sanbox.NewPythonSandbox("/path/to/uploads")

// 方法1: 使用ExecuteCode（推荐）
result, err := sandbox.ExecuteCode(`
x = 10
y = 20
result = x + y
print(f"计算结果: {result}")
result
`)

// 方法2: 使用ExecutePython（兼容API）
result, err := sandbox.ExecutePython(`
import pandas as pd
df = pd.DataFrame({'A': [1, 2, 3], 'B': [4, 5, 6]})
df
`)

if err != nil {
    log.Printf("执行失败: %v", err)
    return
}

if result.Success {
    fmt.Printf("输出类型: %s\n", result.OutputType)
    fmt.Printf("结果: %v\n", result.Output)
    fmt.Printf("打印输出: %s\n", result.Stdout)
    if result.ImagePath != "" {
        fmt.Printf("图片路径: %s\n", result.ImagePath)
    }
} else {
    fmt.Printf("执行错误: %s\n", result.Error)
}
```

### 配置选项

```go
// 设置超时时间
sandbox.SetTimeout(60 * time.Second)

// 设置自定义Python路径
sandbox.SetPythonPath("/custom/path/to/python")
```

## API参考

### 结构体

```go
type PythonSandbox struct {
    timeout    time.Duration
    uploadDir  string
    pythonPath string
}

type PythonExecutionResult struct {
    Success    bool        `json:"success"`        // 是否执行成功
    Output     interface{} `json:"output"`         // 主要返回值
    OutputType string      `json:"output_type"`    // 输出类型
    Error      string      `json:"error"`          // 错误信息
    ImagePath  string      `json:"image_path"`     // 图片文件路径
    Stdout     string      `json:"stdout"`         // 标准输出
    Stderr     string      `json:"stderr"`         // 标准错误输出
}
```

### 方法

| 方法 | 描述 |
|------|------|
| `NewPythonSandbox(uploadDir string)` | 创建新的沙箱实例 |
| `ExecuteCode(code string)` | 执行Python代码（主要API） |
| `ExecutePython(code string)` | 执行Python代码（兼容API） |
| `SetTimeout(timeout time.Duration)` | 设置超时时间 |
| `SetPythonPath(path string)` | 设置Python解释器路径 |
| `InstallRequiredPackages()` | 安装必需的Python包 |

## 支持的数据类型

| 类型 | 描述 | 示例 |
|------|------|------|
| `text` | 字符串文本 | `"Hello World"` |
| `number` | 数字(int/float) | `42`, `3.14` |
| `boolean` | 布尔值 | `True`, `False` |
| `dict` | 字典对象 | `{"key": "value"}` |
| `list` | 列表/数组 | `[1, 2, 3]` |
| `dataframe` | Pandas DataFrame | 包含列名、数据、形状等信息 |
| `image` | 图片文件 | matplotlib生成的图表 |
| `none` | 空值 | `None` |

## 使用示例

### 1. 基本计算

```python
x = 10
y = 20
result = x + y
print(f"计算结果: {result}")
result  # 返回30 (number类型)
```

### 2. 数据分析

```python
import pandas as pd
df = pd.DataFrame({
    'name': ['Alice', 'Bob', 'Charlie'],
    'age': [25, 30, 35],
    'salary': [50000, 60000, 70000]
})
print("数据创建完成")
df  # 返回DataFrame信息 (dataframe类型)
```

### 3. 数据可视化

```python
import matplotlib.pyplot as plt
import numpy as np

x = np.linspace(0, 2*np.pi, 100)
y = np.sin(x)

plt.figure(figsize=(8, 6))
plt.plot(x, y, 'b-', linewidth=2)
plt.title('正弦函数')
plt.xlabel('X')
plt.ylabel('sin(X)')
plt.grid(True)

# 图片会自动保存，ImagePath字段包含文件路径
```

### 4. 错误处理

```python
# 这会产生错误并被正确捕获
x = 1 / 0
```

## 测试

运行单元测试：

```bash
go test ./internal/utils/sanbox -v
```

运行示例：

```bash
go run cmd/python_sandbox_unified_example.go
```

## API集成示例

```go
func handlePythonExecution(c *gin.Context) {
    var req struct {
        Code string `json:"code"`
    }
    
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    
    sandbox := sanbox.NewPythonSandbox("./uploads")
    result, err := sandbox.ExecuteCode(req.Code)
    
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(200, result)
}
```

## 迁移指南

如果你之前使用了两个不同的工具：

```go
// 旧的用法
executor := utils.NewPythonExecutor("/uploads")
sandbox := utils.NewPythonSandbox("/uploads")

// 新的统一用法 
sandbox := sanbox.NewPythonSandbox("/uploads")

// 两个方法都可以使用
result1, _ := sandbox.ExecuteCode(code)      // 推荐
result2, _ := sandbox.ExecutePython(code)    // 兼容
```

## 注意事项

1. 确保系统已安装Python 3.x
2. 确保上传目录有写权限
3. 长时间运行的代码建议增加超时时间
4. 图片文件需要定期清理以避免占用过多磁盘空间
5. 生产环境建议添加资源限制和更严格的安全措施
