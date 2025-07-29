# Agent系统架构 - Eino实现

本目录包含基于CloudWeGo Eino框架的AI数据分析智能体系统实现。

## 架构概述

采用CloudWeGo Eino v0.4.0企业级LLM应用框架，提供高性能、可扩展的智能体系统。

### 核心组件

1. **EinoAgentManager** - 智能体管理器
   - 统一管理多种智能体
   - 提供生命周期管理
   - 支持动态配置

2. **智能体类型**
   - **EinoMainAgent** - 主智能体，负责任务分发和协调
   - **EinoReactAgent** - ReAct智能体，支持推理和行动循环
   - **EinoAnalysisAgent** - 数据分析专用智能体

3. **工具生态系统**
   - **PythonAnalysisTool** - Python代码执行和数据分析
   - **DataVisualizationTool** - 数据可视化图表生成
   - **StatisticalAnalysisTool** - 统计分析工具

4. **EinoAgentSystem** - 系统级协调器
   - 管理智能体间通信
   - 处理查询路由
   - 支持流式处理

## 主要特性

- ✅ **企业级框架**: 基于CloudWeGo Eino，生产就绪
- ✅ **ReAct架构**: 支持推理-行动循环的智能决策
- ✅ **工具调用**: 丰富的数据分析工具生态
- ✅ **流式处理**: 支持实时响应和大数据处理
- ✅ **多智能体协作**: 智能体间无缝协作
- ✅ **灵活配置**: 支持动态配置和扩展

## 快速开始

### 1. 创建智能体系统

```go
// 使用构建器模式创建系统
manager, err := NewEinoAgentSystemBuilder().
    WithChatModel(chatModel).
    WithPythonSandbox(sandbox).
    WithDebug(true).
    Build(ctx)

// 创建系统配置
config := &EinoAgentConfig{
    MaxSteps:    10,
    EnableDebug: true,
}

system := NewEinoAgentSystem(manager, config)
```

### 2. 处理查询

```go
// 同步查询
response, err := system.ProcessQuery(ctx, "分析销售数据的趋势")

// 流式查询
stream, err := system.StreamQuery(ctx, "生成数据报告")
for {
    msg, err := stream.Recv()
    if err != nil {
        break
    }
    fmt.Println(msg.Content)
}
```

## 文件说明

- `eino_types.go` - Eino系统类型定义和工具实现
- `eino_agents.go` - 智能体实现（Main、React、Analysis）
- `eino_manager.go` - 管理器和系统构建器
- `eino_test.go` - 完整测试套件
- `README_EINO.md` - 详细的Eino集成指南

## 测试

```bash
go test -v
```

当前测试覆盖：
- ✅ EinoAgentManager测试
- ✅ EinoAgentSystemBuilder测试  
- ✅ 所有测试通过

## 技术栈

- **Go 1.23.7** - 编程语言
- **CloudWeGo Eino v0.4.0** - LLM应用框架
- **ReAct架构** - 推理-行动循环
- **流式处理** - 实时响应支持

## 更多信息

详细的集成和使用指南请参考：
- [README_EINO.md](./README_EINO.md) - Eino框架集成指南
- [INTEGRATION.md](./INTEGRATION.md) - 系统集成文档
