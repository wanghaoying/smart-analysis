package handler

import (
	"net/http"
	"smart-analysis/internal/config"
	"smart-analysis/internal/model"
	"smart-analysis/internal/service"
	"smart-analysis/internal/utils"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService *service.UserService
	config      *config.Config
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
		config:      config.Load(),
	}
}

// Register 用户注册
func (h *UserHandler) Register(c *gin.Context) {
	var req model.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    400,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	user, err := h.userService.Register(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	// 生成JWT token
	token, err := utils.GenerateToken(user.ID, user.Username, h.config.JWTSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Code:    500,
			Message: "Failed to generate token",
		})
		return
	}

	c.JSON(http.StatusCreated, model.Response{
		Code:    201,
		Message: "User registered successfully",
		Data: model.LoginResponse{
			Token: token,
			User:  *user,
		},
	})
}

// Login 用户登录
func (h *UserHandler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    400,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	user, err := h.userService.Login(&req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.Response{
			Code:    401,
			Message: err.Error(),
		})
		return
	}

	// 生成JWT token
	token, err := utils.GenerateToken(user.ID, user.Username, h.config.JWTSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Code:    500,
			Message: "Failed to generate token",
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Code:    200,
		Message: "Login successful",
		Data: model.LoginResponse{
			Token: token,
			User:  *user,
		},
	})
}

// GetProfile 获取用户资料
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID := c.GetInt("user_id")

	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, model.Response{
			Code:    404,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Code:    200,
		Message: "Success",
		Data:    user,
	})
}

// UpdateProfile 更新用户资料
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID := c.GetInt("user_id")

	var req model.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    400,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	user, err := h.userService.UpdateProfile(userID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Code:    200,
		Message: "Profile updated successfully",
		Data:    user,
	})
}
