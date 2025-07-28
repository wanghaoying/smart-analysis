package handler

import (
	"net/http"
	"smart-analysis/internal/model"
	"smart-analysis/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type FileHandler struct {
	fileService *service.FileService
}

func NewFileHandler(fileService *service.FileService) *FileHandler {
	return &FileHandler{
		fileService: fileService,
	}
}

// Upload 文件上传
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
