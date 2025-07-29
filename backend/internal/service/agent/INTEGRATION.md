# 智能体系统集成指南

## 项目集成

### 1. 在现有的分析服务中集成智能体系统

修改 `backend/internal/service/analysis.go` 文件，集成多智能体系统：

```go
package service

import (
    "context"
    "smart-analysis/internal/service/agent"
    "smart-analysis/internal/utils/llm"
    "smart-analysis/internal/utils/sanbox"
)

type AnalysisService struct {
    agentSystem *agent.AgentSystem
    // ... 其他字段
}

func NewAnalysisService(llmClient llm.LLMClient, uploadDir string) *AnalysisService {
    // 创建Python沙箱
    pythonSandbox := sanbox.NewPythonSandbox(uploadDir)
    
    // 创建智能体系统
    agentSystem, err := agent.NewAgentSystemBuilder().
        WithLLMClient(llmClient).
        WithPythonSandbox(pythonSandbox).
        WithModel("gpt-3.5-turbo").
        WithTemperature(0.7).
        WithMaxTokens(2048).
        Build()
    
    if err != nil {
        // 处理错误
        panic(err)
    }
    
    return &AnalysisService{
        agentSystem: agentSystem,
    }
}

func (s *AnalysisService) ProcessQuery(ctx context.Context, sessionID, userID int, query string, fileData *agent.FileData) (*agent.AnalysisResult, error) {
    analysisCtx := &agent.AnalysisContext{
        SessionID: sessionID,
        UserID:    userID,
        Query:     query,
        FileData:  fileData,
    }
    
    return s.agentSystem.ProcessQuery(ctx, analysisCtx)
}
```

### 2. 在HTTP处理器中使用

修改 `backend/internal/handler/analysis.go`：

```go
func (h *AnalysisHandler) StreamAnalysis(c *gin.Context) {
    var req struct {
        SessionID int    `json:"session_id"`
        Query     string `json:"query"`
        FileID    *int   `json:"file_id,omitempty"`
    }
    
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    
    // 获取文件数据
    var fileData *agent.FileData
    if req.FileID != nil {
        // 从数据库获取文件信息并转换为FileData
        // ...
    }
    
    // 使用智能体系统处理查询
    result, err := h.analysisService.ProcessQuery(
        c.Request.Context(),
        req.SessionID,
        getUserID(c),
        req.Query,
        fileData,
    )
    
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(200, gin.H{
        "result": result,
    })
}
```

### 3. 配置管理

在 `backend/internal/config/config.go` 中添加智能体相关配置：

```go
type Config struct {
    // ... 现有配置
    
    Agent struct {
        Model       string  `yaml:"model" default:"gpt-3.5-turbo"`
        Temperature float64 `yaml:"temperature" default:"0.7"`
        MaxTokens   int     `yaml:"max_tokens" default:"2048"`
        Timeout     int     `yaml:"timeout" default:"30"`
        MaxRetries  int     `yaml:"max_retries" default:"3"`
        EnableDebug bool    `yaml:"enable_debug" default:"false"`
    } `yaml:"agent"`
}
```

### 4. 数据库模型扩展

可以考虑在数据库中添加智能体执行记录表：

```sql
CREATE TABLE agent_executions (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    session_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    query_text TEXT NOT NULL,
    agent_type VARCHAR(50) NOT NULL,
    execution_plan JSON,
    result JSON,
    status VARCHAR(20) NOT NULL,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_session_id (session_id),
    INDEX idx_user_id (user_id),
    INDEX idx_status (status)
);
```

### 5. 监控和日志

添加智能体系统的监控和日志：

```go
import (
    "log/slog"
    "time"
)

func (s *AnalysisService) ProcessQueryWithMonitoring(ctx context.Context, analysisCtx *agent.AnalysisContext) (*agent.AnalysisResult, error) {
    start := time.Now()
    
    slog.Info("Starting agent query processing", 
        "session_id", analysisCtx.SessionID,
        "user_id", analysisCtx.UserID,
        "query", analysisCtx.Query)
    
    result, err := s.agentSystem.ProcessQuery(ctx, analysisCtx)
    
    duration := time.Since(start)
    
    if err != nil {
        slog.Error("Agent query processing failed",
            "session_id", analysisCtx.SessionID,
            "error", err,
            "duration", duration)
        return nil, err
    }
    
    slog.Info("Agent query processing completed",
        "session_id", analysisCtx.SessionID,
        "result_type", result.Type,
        "duration", duration)
    
    return result, nil
}
```

## 扩展智能体

### 添加新的智能体类型

1. 定义新的智能体类型：

```go
const (
    AgentTypeReportGeneration AgentType = "report_generation"
)
```

2. 实现智能体接口：

```go
type ReportGenerationAgent struct {
    config *AgentConfig
}

func (a *ReportGenerationAgent) GetType() AgentType {
    return AgentTypeReportGeneration
}

func (a *ReportGenerationAgent) Process(ctx context.Context, msg *AgentMessage) (*AgentMessage, error) {
    // 实现报告生成逻辑
}

func (a *ReportGenerationAgent) Initialize(ctx context.Context) error {
    return nil
}

func (a *ReportGenerationAgent) Shutdown(ctx context.Context) error {
    return nil
}
```

3. 在系统中注册新智能体：

```go
func (s *AgentSystem) RegisterReportAgent() error {
    reportAgent := NewReportGenerationAgent(s.config)
    return s.manager.RegisterAgent(reportAgent)
}
```

### 自定义任务类型

在DataAnalysisAgent中添加新的任务类型处理：

```go
func (a *DataAnalysisAgent) executeTask(ctx context.Context, task *Task) (map[string]interface{}, error) {
    switch task.Type {
    case "custom_analysis":
        return a.executeCustomAnalysis(ctx, task)
    // ... 其他案例
    }
}

func (a *DataAnalysisAgent) executeCustomAnalysis(ctx context.Context, task *Task) (map[string]interface{}, error) {
    // 实现自定义分析逻辑
}
```

## 最佳实践

### 1. 错误处理

- 实现重试机制
- 记录详细的错误日志
- 提供用户友好的错误消息

### 2. 性能优化

- 使用连接池管理LLM客户端
- 实现结果缓存
- 限制并发执行数量

### 3. 安全性

- 验证用户输入
- 限制Python代码执行权限
- 过滤敏感信息

### 4. 可观测性

- 添加指标收集
- 实现分布式链路追踪
- 监控资源使用情况

## 测试策略

### 单元测试

```go
func TestDataAnalysisAgent(t *testing.T) {
    mockLLM := NewMockLLMClient()
    sandbox := sanbox.NewPythonSandbox("/tmp")
    
    config := &agent.AgentConfig{
        LLMClient: mockLLM,
        PythonSandbox: sandbox,
    }
    
    agent := agent.NewDataAnalysisAgent(config)
    
    // 测试各种场景
}
```

### 集成测试

```go
func TestAgentSystemIntegration(t *testing.T) {
    // 测试完整的智能体系统流程
}
```

### 端到端测试

```go
func TestEndToEndAnalysis(t *testing.T) {
    // 测试从HTTP请求到返回结果的完整流程
}
```
