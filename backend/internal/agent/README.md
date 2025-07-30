# Agent系统架构

本目录包含基于CloudWeGo Eino框架的AI数据分析智能体系统实现。

## 架构概述

采用CloudWeGo Eino v0.4.0企业级LLM应用框架，提供高性能、可扩展的智能体系统。

### 核心组件

1. **AgentManager** - 智能体管理器
   - 统一管理多种智能体
   - 提供生命周期管理
   - 支持动态配置

2. **智能体类型**
   - **MainAgent** - 主智能体，负责任务分发和协调
   - **ReactAgent** - ReAct智能体，支持推理和行动循环

3. **工具生态系统**
   - **PythonAnalysisTool** - Python代码执行、数据分析和统计计算
   - **EChartsVisualizationTool** - ECharts格式的交互式数据可视化
   - **FileReaderTool** - 多格式文件读取和预览
   - **DataQueryTool** - SQL样式的数据查询和筛选

## 主要特性

- ✅ **企业级框架**: 基于CloudWeGo Eino，生产就绪
- ✅ **ReAct架构**: 支持推理-行动循环的智能决策
- ✅ **丰富工具集**: 完整的数据分析工具生态系统
- ✅ **ECharts集成**: 生成交互式图表配置
- ✅ **流式处理**: 支持实时响应和大数据处理
- ✅ **多智能体协作**: 智能体间无缝协作
- ✅ **灵活配置**: 支持动态配置和扩展

## 快速开始

### 1. 创建智能体系统

```go
// 创建智能体配置
config := &AgentConfig{
    ChatModel:     chatModel,
    PythonSandbox: sandbox,
    MaxSteps:      10,
}

// 创建React智能体
agent, err := NewReactAgent(ctx, config)

// 创建管理器
manager := NewAgentManager(config)
manager.RegisterAgent(AgentTypeMain, agent)
```

### 2. 处理查询

```go
// 同步查询
response, err := manager.ProcessQuery(ctx, "分析销售数据趋势并生成图表")

// 流式查询
stream, err := manager.StreamQuery(ctx, "生成数据分析报告")
for {
    msg, err := stream.Recv()
    if err != nil {
        break
    }
    fmt.Println(msg.Content)
}
```

## 工具详解

### 1. PythonAnalysisTool
支持三种分析类型：
- `general`: 通用Python代码执行
- `statistical`: 统计分析（描述性统计、相关性分析、假设检验）
- `cleaning`: 数据清洗（异常值检测、缺失值处理）

### 2. EChartsVisualizationTool
生成ECharts格式配置：
```json
{
  "type": "bar|line|pie|scatter|heatmap",
  "title": "图表标题",
  "data": [{"name": "项目", "value": 100}],
  "xAxis": ["标签"],
  "series": [{"name": "数据", "type": "bar", "data": [100]}]
}
```

### 3. FileReaderTool
支持格式：
- CSV文件
- Excel文件（.xlsx, .xls）
- JSON文件

### 4. DataQueryTool
支持pandas查询语法：
- 条件筛选：`query('age > 25')`
- 分组聚合：`groupby('category').sum()`
- 统计计算：`describe()`

## 文件说明

- `tools.go` - 工具定义和实现
- `agents.go` - 智能体实现（Main、React）
- `manager.go` - 管理器和系统协调器
- `agent_test.go` - 测试套件

## 前端集成

前端支持ECharts渲染：

```tsx
// 使用ECharts组件
<EChartsDisplay config={chartConfig} />

// Markdown中自动渲染
<MarkdownRenderer content={responseText} />
```

## 测试

```bash
go test -v
```

## 技术栈

- **Go 1.18+** - 编程语言
- **CloudWeGo Eino v0.4.0** - LLM应用框架
- **ReAct架构** - 推理-行动循环
- **ECharts** - 交互式图表渲染
- **流式处理** - 实时响应支持
