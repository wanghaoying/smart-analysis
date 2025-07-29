package sanbox

import (
	"fmt"
	"log"
	"time"
)

func example() {
	// 创建Python沙箱
	sandbox := NewPythonSandbox("/tmp/python_outputs")

	fmt.Println("=== 统一的Python代码执行沙箱示例 ===")

	// 示例1: 基本计算 - 使用ExecuteCode
	fmt.Println("1. 基本计算 (ExecuteCode):")
	result1, err := sandbox.ExecuteCode(`
x = 10
y = 20
result = x + y
print(f"计算: {x} + {y} = {result}")
result
`)
	if err != nil {
		log.Printf("执行失败: %v", err)
	} else if result1.Success {
		fmt.Printf("类型: %s, 结果: %v\n", result1.OutputType, result1.Output)
		fmt.Printf("输出: %s\n", result1.Stdout)
	} else {
		fmt.Printf("错误: %s\n", result1.Error)
	}

	// 示例2: DataFrame - 使用ExecutePython（兼容API）
	fmt.Println("\n2. Pandas DataFrame (ExecutePython):")
	result2, err := sandbox.ExecutePython(`
import pandas as pd
df = pd.DataFrame({
    'name': ['Alice', 'Bob', 'Charlie'],
    'age': [25, 30, 35],
    'salary': [50000, 60000, 70000]
})
print("DataFrame创建完成")
df
`)
	if err != nil {
		log.Printf("执行失败: %v", err)
	} else if result2.Success {
		fmt.Printf("类型: %s\n", result2.OutputType)
		if result2.OutputType == "dataframe" {
			if data, ok := result2.Output.(map[string]interface{}); ok {
				fmt.Printf("形状: %v\n", data["shape"])
				fmt.Printf("列名: %v\n", data["columns"])
			}
		}
		fmt.Printf("输出: %s\n", result2.Stdout)
	} else {
		fmt.Printf("错误: %s\n", result2.Error)
	}

	// 示例3: 图表生成
	fmt.Println("\n3. 生成图表:")
	result3, err := sandbox.ExecuteCode(`
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

print("图表已生成")
`)
	if err != nil {
		log.Printf("执行失败: %v", err)
	} else if result3.Success {
		fmt.Printf("类型: %s\n", result3.OutputType)
		if result3.ImagePath != "" {
			fmt.Printf("图片路径: %s\n", result3.ImagePath)
		}
		fmt.Printf("输出: %s\n", result3.Stdout)
	} else {
		fmt.Printf("错误: %s\n", result3.Error)
	}

	// 示例4: 设置自定义超时
	fmt.Println("\n4. 自定义超时测试:")
	sandbox.SetTimeout(5 * time.Second) // 设置5秒超时
	result4, err := sandbox.ExecuteCode(`
import time
print("开始睡眠...")
time.sleep(2)  # 睡眠2秒，应该成功
print("睡眠结束")
"完成"
`)
	if err != nil {
		log.Printf("执行失败: %v", err)
	} else if result4.Success {
		fmt.Printf("类型: %s, 结果: %v\n", result4.OutputType, result4.Output)
		fmt.Printf("输出: %s\n", result4.Stdout)
	} else {
		fmt.Printf("错误: %s\n", result4.Error)
	}

	// 示例5: 错误处理
	fmt.Println("\n5. 错误处理:")
	result5, err := sandbox.ExecuteCode(`
# 产生除零错误
x = 1 / 0
`)
	if err != nil {
		log.Printf("执行失败: %v", err)
	} else if result5.Success {
		fmt.Printf("意外成功: %v\n", result5.Output)
	} else {
		fmt.Printf("正确捕获错误: %s\n", result5.Error)
	}

	fmt.Println("\n=== 示例完成 ===")
}
