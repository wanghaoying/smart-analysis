# AI 数据分析智能体系统 - Eino 集成指南

本项目基于 CloudWeGo Eino 框架实现了一个完整的AI数据分析智能体系统，提供Python沙箱执行、数据可视化、统计分析等功能。

## 🏗️ 系统架构

### 核心组件

#### 1. Eino 智能体系统
- **EinoReactAgent**: 基于Eino React架构的核心智能体，具备工具调用和推理能力
- **EinoMainAgent**: 主智能体，负责意图识别和查询改写
- **EinoAnalysisAgent**: 数据分析专用智能体

#### 2. 工具系统 (Tools)
- **PythonAnalysisTool**: Python代码执行工具
- **DataVisualizationTool**: 数据可视化工具
- **StatisticalAnalysisTool**: 统计分析工具

#### 3. 管理系统
- **EinoAgentManager**: 智能体管理器，负责注册、协调和消息路由
- **EinoAgentSystemBuilder**: 构建器模式的系统初始化工具

#### 4. 集成适配
- **EinoLLMModelAdapter**: 将现有LLM客户端适配为Eino模型接口
- **EinoAgentIntegration**: 提供与传统系统的集成接口

## 🚀 快速开始

### 1. 依赖安装

```bash
go mod tidy
```

主要依赖：
- `github.com/cloudwego/eino v0.4.0` - 核心框架
- Python沙箱执行环境
- LLM客户端 (OpenAI/Hunyuan等)

### 2. 系统初始化

```go
// 创建LLM客户端
llmClient := // 你的LLM客户端实现

// 创建Python沙箱
pythonSandbox := &sanbox.PythonSandbox{}

// 使用Eino系统
einoSystem, err := agent.CreateEinoAgentSystemFromExisting(
    ctx,
    llmClient,
    pythonSandbox,
    true, // 启用调试模式
)
if err != nil {
    log.Fatal("Failed to create Eino system:", err)
}
defer einoSystem.Shutdown(ctx)
```

### 3. 使用示例

#### 基本查询
```go
response, err := einoSystem.ProcessQuery(ctx, "分析这个数据集的基本统计信息")
if err != nil {
    log.Fatal("Query failed:", err)
}
fmt.Println("Response:", response.Content)
```

#### 流式查询
```go
stream, err := einoSystem.StreamQuery(ctx, "创建一个散点图来展示数据关系")
if err != nil {
    log.Fatal("Stream query failed:", err)
}

for {
    chunk, err := stream.Recv()
    if err != nil {
        if err.Error() == "EOF" {
            break
        }
        log.Fatal("Stream error:", err)
    }
    
    if chunk != nil {
        fmt.Print(chunk.Content)
    }
}
stream.Close()
```

## 🔧 高级配置

### 自定义工具
```go
// 创建自定义工具
customTool := // 你的工具实现

// 使用构建器添加工具
manager, err := agent.NewEinoAgentSystemBuilder().
    WithChatModel(modelAdapter).
    WithPythonSandbox(pythonSandbox).
    WithTools([]tool.BaseTool{customTool}).
    WithMaxSteps(15).
    WithDebug(true).
    Build(ctx)
```

### 多智能体协作
```go
// 获取管理器
manager := einoSystem.GetManager()

// 获取特定智能体
reactAgent, exists := manager.GetAgent(agent.EinoAgentTypeReact)
if exists {
    // 直接与特定智能体交互
    response, err := reactAgent.Generate(ctx, messages)
}
```

## 🧪 测试

运行完整测试套件：
```bash
go test ./internal/service/agent -v
```

运行特定的Eino测试：
```bash
go test ./internal/service/agent -v -run TestEino
```

## 📊 功能特性

### 1. 智能工具调用
- 自动识别用户意图
- 智能选择合适的工具
- 支持复杂的多步推理

### 2. Python沙箱集成
- 安全的代码执行环境
- 支持数据分析库 (pandas, numpy, matplotlib等)
- 结果输出和错误处理

### 3. 流式响应
- 实时响应用户查询
- 支持长时间运行的分析任务
- 渐进式结果展示

### 4. 多模态输出
- 文本分析结果
- 数据可视化图表
- 统计报告

## 🔗 API 接口

### 核心接口

#### EinoAgent
```go
type EinoAgent interface {
    GetType() EinoAgentType
    Generate(ctx context.Context, messages []*schema.Message, opts ...interface{}) (*schema.Message, error)
    Stream(ctx context.Context, messages []*schema.Message, opts ...interface{}) (*schema.StreamReader[*schema.Message], error)
    Initialize(ctx context.Context) error
    Shutdown(ctx context.Context) error
}
```

#### 工具接口
```go
type PythonAnalysisTool struct {
    sandbox *sanbox.PythonSandbox
    name    string
    desc    string
}

// 实现 tool.InvokableTool 接口
func (t *PythonAnalysisTool) Info(ctx context.Context) (*schema.ToolInfo, error)
func (t *PythonAnalysisTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error)
```

## 🎯 最佳实践

### 1. 错误处理
```go
response, err := einoSystem.ProcessQuery(ctx, query)
if err != nil {
    // 记录错误并提供友好的用户反馈
    log.Printf("Query processing failed: %v", err)
    return handleError(err)
}
```

### 2. 资源管理
```go
// 确保正确关闭系统
defer func() {
    if err := einoSystem.Shutdown(ctx); err != nil {
        log.Printf("System shutdown error: %v", err)
    }
}()
```

### 3. 调试模式
```go
// 开发环境启用调试
einoSystem, err := agent.CreateEinoAgentSystemFromExisting(
    ctx,
    llmClient,
    pythonSandbox,
    os.Getenv("ENV") == "development", // 调试模式
)
```

## 🔧 故障排除

### 常见问题

1. **工具调用失败**
   - 检查Python沙箱环境
   - 验证工具参数格式
   - 查看调试日志

2. **LLM适配问题**
   - 确保实现了ToolCallingChatModel接口
   - 检查WithTools方法实现
   - 验证消息格式转换

3. **流式响应中断**
   - 检查网络连接
   - 验证流式客户端实现
   - 确保正确关闭流

## 📈 性能优化

### 1. 并发处理
- 工具调用支持并行执行
- 流式响应减少延迟
- 智能体池化管理

### 2. 缓存策略
- LLM响应缓存
- 工具执行结果缓存
- 智能体状态缓存

### 3. 资源限制
- Python代码执行超时
- 内存使用限制
- 并发请求控制

## 🎉 总结

基于CloudWeGo Eino框架的AI数据分析智能体系统提供了：

✅ **企业级框架**: 使用成熟的Eino框架，提供可靠的智能体编排能力  
✅ **完整工具生态**: 内置Python分析、数据可视化、统计分析工具  
✅ **流式响应**: 支持实时交互和长时间分析任务  
✅ **多智能体协作**: 主智能体+专用智能体的分层架构  
✅ **向后兼容**: 与现有系统无缝集成  
✅ **测试完备**: 全面的测试套件保证系统稳定性  

这个系统为数据分析场景提供了强大、灵活、可扩展的AI智能体解决方案！

## 📝 更新日志

### v2.0.0 - Eino集成版本
- 🎉 集成CloudWeGo Eino框架
- ⚡ 添加React智能体支持
- 🔧 重构工具系统
- 📊 改进流式响应
- 🧪 完善测试覆盖

### v1.0.0 - 基础版本
- 🚀 初始版本发布
- 🤖 基础智能体系统
- 🐍 Python沙箱集成
- 📈 数据分析功能
