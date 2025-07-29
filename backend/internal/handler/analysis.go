package handler

import (
	"net/http"
	"smart-analysis/internal/model"
	"smart-analysis/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AnalysisHandler struct {
	analysisService *service.AnalysisService
	fileService     *service.FileService
}

func NewAnalysisHandler(analysisService *service.AnalysisService) *AnalysisHandler {
	return &AnalysisHandler{
		analysisService: analysisService,
		fileService:     service.NewFileService(), // 这应该通过依赖注入
	}
}

// Query 处理分析查询
func (h *AnalysisHandler) Query(c *gin.Context) {
	userID := c.GetInt("user_id")

	var req model.QueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    400,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	response, err := h.analysisService.Query(userID, &req, h.fileService)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Code:    200,
		Message: "Query processed successfully",
		Data:    response,
	})
}

// Visualize 生成可视化图表
//func (h *AnalysisHandler) Visualize(c *gin.Context) {
//	userID := c.GetInt("user_id")
//
//	var req model.VisualizationRequest
//	if err := c.ShouldBindJSON(&req); err != nil {
//		c.JSON(http.StatusBadRequest, model.Response{
//			Code:    400,
//			Message: "Invalid request: " + err.Error(),
//		})
//		return
//	}
//
//	response, err := h.analysisService.Visualize(userID, &req, h.fileService)
//	if err != nil {
//		c.JSON(http.StatusBadRequest, model.Response{
//			Code:    400,
//			Message: err.Error(),
//		})
//		return
//	}
//
//	c.JSON(http.StatusOK, model.Response{
//		Code:    200,
//		Message: "Visualization generated successfully",
//		Data:    response,
//	})
//}
//
//// GenerateReport 生成分析报告
//func (h *AnalysisHandler) GenerateReport(c *gin.Context) {
//	userID := c.GetInt("user_id")
//
//	var req model.ReportRequest
//	if err := c.ShouldBindJSON(&req); err != nil {
//		c.JSON(http.StatusBadRequest, model.Response{
//			Code:    400,
//			Message: "Invalid request: " + err.Error(),
//		})
//		return
//	}
//
//	response, err := h.analysisService.GenerateReport(userID, &req, h.fileService)
//	if err != nil {
//		c.JSON(http.StatusBadRequest, model.Response{
//			Code:    400,
//			Message: err.Error(),
//		})
//		return
//	}
//
//	c.JSON(http.StatusOK, model.Response{
//		Code:    200,
//		Message: "Report generated successfully",
//		Data:    response,
//	})
//}

// GetHistory 获取查询历史
func (h *AnalysisHandler) GetHistory(c *gin.Context) {
	userID := c.GetInt("user_id")

	// 获取可选的session_id参数
	var sessionID *int
	if sessionIDStr := c.Query("session_id"); sessionIDStr != "" {
		if id, err := strconv.Atoi(sessionIDStr); err == nil {
			sessionID = &id
		}
	}

	history, err := h.analysisService.GetHistory(userID, sessionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Code:    200,
		Message: "Success",
		Data:    history,
	})
}

// CreateSession 创建会话
func (h *AnalysisHandler) CreateSession(c *gin.Context) {
	userID := c.GetInt("user_id")

	var req model.CreateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    400,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	session, err := h.analysisService.CreateSession(userID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, model.Response{
		Code:    201,
		Message: "Session created successfully",
		Data:    session,
	})
}

// GetSession 获取会话详情
func (h *AnalysisHandler) GetSession(c *gin.Context) {
	userID := c.GetInt("user_id")

	sessionIDStr := c.Param("id")
	sessionID, err := strconv.Atoi(sessionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    400,
			Message: "Invalid session ID",
		})
		return
	}

	session, err := h.analysisService.GetSession(userID, sessionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Code:    200,
		Message: "Success",
		Data:    session,
	})
}

// ConfigLLM 配置LLM
func (h *AnalysisHandler) ConfigLLM(c *gin.Context) {
	userID := c.GetInt("user_id")

	var req model.LLMConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    400,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	config, err := h.analysisService.ConfigLLM(userID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, model.Response{
		Code:    201,
		Message: "LLM configured successfully",
		Data:    config,
	})
}

// GetLLMConfig 获取LLM配置
func (h *AnalysisHandler) GetLLMConfig(c *gin.Context) {
	userID := c.GetInt("user_id")

	configs, err := h.analysisService.GetLLMConfig(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Code:    200,
		Message: "Success",
		Data:    configs,
	})
}

// GetUsage 获取使用量统计
func (h *AnalysisHandler) GetUsage(c *gin.Context) {
	userID := c.GetInt("user_id")

	usage, err := h.analysisService.GetUsage(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Code:    200,
		Message: "Success",
		Data:    usage,
	})
}
