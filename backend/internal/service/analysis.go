package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"smart-analysis/internal/model"
	"strings"
	"time"
)

type AnalysisService struct {
	sessions      map[int]*model.Session
	queries       map[int]*model.Query
	llmConfigs    map[int][]*model.LLMConfig
	usage         map[int][]*model.Usage
	nextSessionID int
	nextQueryID   int
	nextConfigID  int
	nextUsageID   int
}

func NewAnalysisService() *AnalysisService {
	return &AnalysisService{
		sessions:      make(map[int]*model.Session),
		queries:       make(map[int]*model.Query),
		llmConfigs:    make(map[int][]*model.LLMConfig),
		usage:         make(map[int][]*model.Usage),
		nextSessionID: 1,
		nextQueryID:   1,
		nextConfigID:  1,
		nextUsageID:   1,
	}
}

// CreateSession 创建会话
func (s *AnalysisService) CreateSession(userID int, req *model.CreateSessionRequest) (*model.Session, error) {
	session := &model.Session{
		ID:        s.nextSessionID,
		UserID:    userID,
		Name:      req.Name,
		FileID:    req.FileID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	s.sessions[s.nextSessionID] = session
	s.nextSessionID++

	return session, nil
}

// GetSession 获取会话
func (s *AnalysisService) GetSession(userID, sessionID int) (*model.Session, error) {
	session, exists := s.sessions[sessionID]
	if !exists {
		return nil, errors.New("session not found")
	}

	if session.UserID != userID {
		return nil, errors.New("permission denied")
	}

	return session, nil
}

// Query 处理查询请求
func (s *AnalysisService) Query(userID int, req *model.QueryRequest, fileService *FileService) (*model.QueryResponse, error) {
	// 获取会话
	session, err := s.GetSession(userID, req.SessionID)
	if err != nil {
		return nil, err
	}

	// 获取文件数据（如果指定了文件）
	var fileData interface{}
	if req.FileID != nil {
		file, err := fileService.GetFileByID(*req.FileID)
		if err != nil {
			return nil, err
		}

		if file.UserID != userID {
			return nil, errors.New("permission denied")
		}

		fileData, err = fileService.PreviewFile(userID, *req.FileID, 100) // 获取前100行
		if err != nil {
			return nil, err
		}
	}

	// 创建查询记录
	query := &model.Query{
		ID:        s.nextQueryID,
		SessionID: session.ID,
		UserID:    userID,
		Question:  req.Question,
		QueryType: "analysis",
		Status:    "processing",
		CreatedAt: time.Now(),
	}

	s.queries[s.nextQueryID] = query
	s.nextQueryID++

	// 调用LLM进行分析
	answer, err := s.callLLM(userID, req.Question, fileData)
	if err != nil {
		query.Status = "error"
		return nil, err
	}

	query.Answer = answer
	query.Status = "completed"

	return &model.QueryResponse{
		Answer:    answer,
		Data:      fileData,
		QueryType: "analysis",
		Status:    "completed",
	}, nil
}

// Visualize 生成可视化图表
func (s *AnalysisService) Visualize(userID int, req *model.VisualizationRequest, fileService *FileService) (*model.VisualizationResponse, error) {
	// 获取文件数据
	file, err := fileService.GetFileByID(req.FileID)
	if err != nil {
		return nil, err
	}

	if file.UserID != userID {
		return nil, errors.New("permission denied")
	}

	fileData, err := fileService.PreviewFile(userID, req.FileID, -1) // 获取所有数据
	if err != nil {
		return nil, err
	}

	// 调用LLM生成图表代码
	prompt := fmt.Sprintf("基于以下数据生成%s图表的配置: %s\n数据: %v",
		req.ChartType, req.Query, fileData)

	chartConfig, err := s.callLLM(userID, prompt, fileData)
	if err != nil {
		return nil, err
	}

	// 创建查询记录
	query := &model.Query{
		ID:        s.nextQueryID,
		SessionID: req.SessionID,
		UserID:    userID,
		Question:  req.Query,
		Answer:    chartConfig,
		QueryType: "visualization",
		Status:    "completed",
		CreatedAt: time.Now(),
	}

	s.queries[s.nextQueryID] = query
	s.nextQueryID++

	return &model.VisualizationResponse{
		ChartData: chartConfig,
		ChartType: req.ChartType,
		Title:     "数据可视化图表",
	}, nil
}

// GenerateReport 生成报告
func (s *AnalysisService) GenerateReport(userID int, req *model.ReportRequest, fileService *FileService) (*model.ReportResponse, error) {
	// 获取文件数据
	file, err := fileService.GetFileByID(req.FileID)
	if err != nil {
		return nil, err
	}

	if file.UserID != userID {
		return nil, errors.New("permission denied")
	}

	fileData, err := fileService.PreviewFile(userID, req.FileID, -1)
	if err != nil {
		return nil, err
	}

	// 构建报告提示
	prompt := fmt.Sprintf(`
请基于以下数据生成一份详细的分析报告：
数据描述: %s
分析维度: %s
数据内容: %v

请生成包含以下部分的报告：
1. 数据概览
2. 关键指标分析
3. 趋势分析
4. 结论和建议
`, req.Description, strings.Join(req.Dimensions, ", "), fileData)

	// 调用LLM生成报告
	reportContent, err := s.callLLM(userID, prompt, fileData)
	if err != nil {
		return nil, err
	}

	// 创建查询记录
	query := &model.Query{
		ID:        s.nextQueryID,
		SessionID: req.SessionID,
		UserID:    userID,
		Question:  prompt,
		Answer:    reportContent,
		QueryType: "report",
		Status:    "completed",
		CreatedAt: time.Now(),
	}

	s.queries[s.nextQueryID] = query
	s.nextQueryID++

	return &model.ReportResponse{
		Content: reportContent,
		Charts:  []interface{}{}, // 这里可以扩展包含图表
		Summary: "报告生成完成",
	}, nil
}

// GetHistory 获取查询历史
func (s *AnalysisService) GetHistory(userID int, sessionID *int) ([]*model.Query, error) {
	var history []*model.Query
	for _, query := range s.queries {
		if query.UserID == userID {
			if sessionID == nil || query.SessionID == *sessionID {
				history = append(history, query)
			}
		}
	}
	return history, nil
}

// ConfigLLM 配置LLM
func (s *AnalysisService) ConfigLLM(userID int, req *model.LLMConfigRequest) (*model.LLMConfig, error) {
	// 如果设置为默认，先取消其他默认配置
	if req.IsDefault {
		for _, config := range s.llmConfigs[userID] {
			config.IsDefault = false
		}
	}

	config := &model.LLMConfig{
		ID:        s.nextConfigID,
		UserID:    userID,
		Provider:  req.Provider,
		APIKey:    req.APIKey,
		Model:     req.Model,
		IsDefault: req.IsDefault,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if s.llmConfigs[userID] == nil {
		s.llmConfigs[userID] = make([]*model.LLMConfig, 0)
	}
	s.llmConfigs[userID] = append(s.llmConfigs[userID], config)
	s.nextConfigID++

	return config, nil
}

// GetLLMConfig 获取LLM配置
func (s *AnalysisService) GetLLMConfig(userID int) ([]*model.LLMConfig, error) {
	return s.llmConfigs[userID], nil
}

// GetUsage 获取使用量统计
func (s *AnalysisService) GetUsage(userID int) (*model.UsageResponse, error) {
	userUsage := s.usage[userID]
	if userUsage == nil {
		return &model.UsageResponse{
			TotalTokens: 0,
			TotalCost:   0,
			Usage:       []*model.Usage{},
		}, nil
	}

	totalTokens := 0
	totalCost := 0.0
	for _, usage := range userUsage {
		totalTokens += usage.Tokens
		totalCost += usage.Cost
	}

	return &model.UsageResponse{
		TotalTokens: totalTokens,
		TotalCost:   totalCost,
		Usage:       userUsage,
	}, nil
}

// callLLM 调用LLM API（模拟实现）
func (s *AnalysisService) callLLM(userID int, prompt string, data interface{}) (string, error) {
	// 获取用户的LLM配置
	var config *model.LLMConfig
	for _, cfg := range s.llmConfigs[userID] {
		if cfg.IsDefault {
			config = cfg
			break
		}
	}

	if config == nil && len(s.llmConfigs[userID]) > 0 {
		config = s.llmConfigs[userID][0] // 使用第一个配置
	}

	if config == nil {
		return "", errors.New("no LLM configuration found")
	}

	// 根据提供商调用不同的API
	switch config.Provider {
	case "openai":
		return s.callOpenAI(config, prompt)
	case "hunyuan":
		return s.callHunyuan(config, prompt)
	default:
		// 默认返回模拟响应
		return s.getMockResponse(prompt, data), nil
	}
}

// callOpenAI 调用OpenAI API
func (s *AnalysisService) callOpenAI(config *model.LLMConfig, prompt string) (string, error) {
	requestBody := map[string]interface{}{
		"model": config.Model,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"max_tokens": 1000,
	}

	jsonData, _ := json.Marshal(requestBody)

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+config.APIKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return "", err
	}

	// 解析响应
	choices, ok := response["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return "", errors.New("invalid response format")
	}

	message := choices[0].(map[string]interface{})["message"].(map[string]interface{})
	content := message["content"].(string)

	return content, nil
}

// callHunyuan 调用腾讯混元API（示例实现）
func (s *AnalysisService) callHunyuan(config *model.LLMConfig, prompt string) (string, error) {
	// 这里应该实现腾讯混元API的调用
	// 目前返回模拟响应
	return s.getMockResponse(prompt, nil), nil
}

// getMockResponse 获取模拟响应
func (s *AnalysisService) getMockResponse(prompt string, data interface{}) string {
	if strings.Contains(prompt, "销量") || strings.Contains(prompt, "销售") {
		return "根据数据分析，销量最高的产品是产品A，销售额为100万元。"
	} else if strings.Contains(prompt, "图表") || strings.Contains(prompt, "可视化") {
		return `{
			"type": "bar",
			"data": {
				"labels": ["产品A", "产品B", "产品C"],
				"datasets": [{
					"label": "销售额",
					"data": [100, 80, 60],
					"backgroundColor": ["#FF6384", "#36A2EB", "#FFCE56"]
				}]
			},
			"options": {
				"responsive": true,
				"title": {
					"display": true,
					"text": "产品销售额对比"
				}
			}
		}`
	} else if strings.Contains(prompt, "报告") {
		return `# 数据分析报告

## 1. 数据概览
本次分析涉及的数据包含多个维度的信息，数据质量良好。

## 2. 关键指标分析
- 总销售额：500万元
- 平均客单价：1000元
- 用户转化率：15%

## 3. 趋势分析
从时间趋势来看，销售额呈现稳步增长态势，特别是在第三季度表现突出。

## 4. 结论和建议
建议继续加强产品推广，特别是针对高价值用户群体。`
	}

	return "根据您的数据，我为您提供以下分析结果和建议..."
}
