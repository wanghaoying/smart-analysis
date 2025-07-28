package model

// 用户相关请求结构
type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type UpdateProfileRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
}

// 分析相关请求结构
type QueryRequest struct {
	SessionID int    `json:"session_id"`
	Question  string `json:"question" binding:"required"`
	FileID    *int   `json:"file_id"`
}

type VisualizationRequest struct {
	SessionID int    `json:"session_id"`
	Query     string `json:"query" binding:"required"`
	FileID    int    `json:"file_id" binding:"required"`
	ChartType string `json:"chart_type"` // bar, line, pie, scatter
}

type ReportRequest struct {
	SessionID   int      `json:"session_id"`
	FileID      int      `json:"file_id" binding:"required"`
	Dimensions  []string `json:"dimensions"`
	Description string   `json:"description"`
}

type CreateSessionRequest struct {
	Name   string `json:"name" binding:"required"`
	FileID *int   `json:"file_id"`
}

// LLM配置相关请求结构
type LLMConfigRequest struct {
	Provider  string `json:"provider" binding:"required"`
	APIKey    string `json:"api_key" binding:"required"`
	Model     string `json:"model" binding:"required"`
	IsDefault bool   `json:"is_default"`
}

// 通用响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type FileUploadResponse struct {
	File File `json:"file"`
}

type QueryResponse struct {
	Answer    string      `json:"answer"`
	Data      interface{} `json:"data,omitempty"`
	QueryType string      `json:"query_type"`
	Status    string      `json:"status"`
}

type VisualizationResponse struct {
	ChartData interface{} `json:"chart_data"`
	ChartType string      `json:"chart_type"`
	Title     string      `json:"title"`
}

type ReportResponse struct {
	Content   string        `json:"content"`
	Charts    []interface{} `json:"charts"`
	Summary   string        `json:"summary"`
	ExportURL string        `json:"export_url,omitempty"`
}

type UsageResponse struct {
	TotalTokens int      `json:"total_tokens"`
	TotalCost   float64  `json:"total_cost"`
	Usage       []*Usage `json:"usage"`
}
