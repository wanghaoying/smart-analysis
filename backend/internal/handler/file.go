package handler

import (
	"net/http"
	"smart-analysis/internal/model"
	"smart-analysis/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

// FileHandler 文件相关接口处理器
// @Description 文件相关接口
// @Tags 文件
// @Router /file [group]
type FileHandler struct {
	fileService *service.FileService
}

func NewFileHandler(fileService *service.FileService) *FileHandler {
	return &FileHandler{
		fileService: fileService,
	}
}

// Upload 文件上传
// @Summary 文件上传
// @Description 上传文件
// @Tags 文件
// @Accept multipart/form-data
// @Produce json
// @Security ApiKeyAuth
// @Param file formData file true "文件"
// @Success 201 {object} model.Response{data=model.FileUploadResponse}
// @Failure 400 {object} model.Response
// @Router /api/file/upload [post]
func (h *FileHandler) Upload(c *gin.Context) {
	userID := c.GetInt("user_id")

	// 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    400,
			Message: "No file uploaded",
		})
		return
	}

	uploadedFile, err := h.fileService.Upload(userID, file)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, model.Response{
		Code:    201,
		Message: "File uploaded successfully",
		Data:    model.FileUploadResponse{File: *uploadedFile},
	})
}

// List 获取文件列表
// @Summary 获取文件列表
// @Description 获取当前用户的文件列表
// @Tags 文件
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} model.Response{data=[]model.File}
// @Router /api/file/list [get]
func (h *FileHandler) List(c *gin.Context) {
	userID := c.GetInt("user_id")

	files := h.fileService.GetFilesByUserID(userID)

	c.JSON(http.StatusOK, model.Response{
		Code:    200,
		Message: "Success",
		Data:    files,
	})
}

// Delete 删除文件
// @Summary 删除文件
// @Description 删除指定文件
// @Tags 文件
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "文件ID"
// @Success 200 {object} model.Response
// @Failure 400 {object} model.Response
// @Router /api/file/{id} [delete]
func (h *FileHandler) Delete(c *gin.Context) {
	userID := c.GetInt("user_id")

	fileIDStr := c.Param("id")
	fileID, err := strconv.Atoi(fileIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    400,
			Message: "Invalid file ID",
		})
		return
	}

	err = h.fileService.DeleteFile(userID, fileID)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Code:    200,
		Message: "File deleted successfully",
	})
}

// Preview 预览文件数据
// @Summary 预览文件数据
// @Description 预览指定文件的部分数据
// @Tags 文件
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "文件ID"
// @Param limit query int false "预览行数，默认50"
// @Success 200 {object} model.Response{data=interface{}}
// @Failure 400 {object} model.Response
// @Router /api/file/preview/{id} [get]
func (h *FileHandler) Preview(c *gin.Context) {
	userID := c.GetInt("user_id")

	fileIDStr := c.Param("id")
	fileID, err := strconv.Atoi(fileIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    400,
			Message: "Invalid file ID",
		})
		return
	}

	// 获取限制行数参数
	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 50
	}

	data, err := h.fileService.PreviewFile(userID, fileID, limit)
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
		Data:    data,
	})
}
