package handler

import (
	"net/http"
	"smart-analysis/internal/config"
	"smart-analysis/internal/model"
	"smart-analysis/internal/service"
	"smart-analysis/internal/utils"

	"github.com/gin-gonic/gin"
	// UserHandler 用户相关接口处理器
	// @Description 用户相关接口
	// @Tags 用户
	// @Router /user [group]
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
// @Summary 用户注册
// @Description 注册新用户
// @Tags 用户
// @Accept json
// @Produce json
// @Param data body model.RegisterRequest true "注册请求体"
// @Success 201 {object} model.Response{data=model.LoginResponse}
// @Failure 400 {object} model.Response
// @Failure 500 {object} model.Response
// @Router /api/register [post]
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
// @Summary 用户登录
// @Description 用户登录获取 token
// @Tags 用户
// @Accept json
// @Produce json
// @Param data body model.LoginRequest true "登录请求体"
// @Success 200 {object} model.Response{data=model.LoginResponse}
// @Failure 400 {object} model.Response
// @Failure 401 {object} model.Response
// @Failure 500 {object} model.Response
// @Router /api/login [post]
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
// @Summary 获取用户资料
// @Description 获取当前登录用户的资料
// @Tags 用户
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} model.Response{data=model.User}
// @Failure 404 {object} model.Response
// @Router /api/profile [get]
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
// @Summary 更新用户资料
// @Description 更新当前登录用户的资料
// @Tags 用户
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param data body model.UpdateProfileRequest true "更新资料请求体"
// @Success 200 {object} model.Response{data=model.User}
// @Failure 400 {object} model.Response
// @Router /api/profile [put]
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
