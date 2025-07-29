package main

import (
	"log"
	"smart-analysis/internal/config"
	"smart-analysis/internal/handler"
	"smart-analysis/internal/middleware"
	"smart-analysis/internal/service"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	cfg := config.Load()

	// 初始化服务
	analysisService := service.NewAnalysisService()
	userService := service.NewUserService()
	fileService := service.NewFileService()

	// 初始化处理器
	analysisHandler := handler.NewAnalysisHandler(analysisService)
	userHandler := handler.NewUserHandler(userService)
	fileHandler := handler.NewFileHandler(fileService)

	// 创建Gin路由
	r := gin.Default()

	// 配置CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://127.0.0.1:3000", "http://localhost:3001", "http://127.0.0.1:3001"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// 静态文件服务
	r.Static("/uploads", "./uploads")

	// API路由组
	api := r.Group("/api/v1")
	{
		// 用户相关路由
		user := api.Group("/user")
		{
			user.POST("/register", userHandler.Register)
			user.POST("/login", userHandler.Login)
			user.GET("/profile", middleware.AuthMiddleware(), userHandler.GetProfile)
			user.PUT("/profile", middleware.AuthMiddleware(), userHandler.UpdateProfile)
		}

		// 文件上传相关路由
		file := api.Group("/file")
		file.Use(middleware.AuthMiddleware())
		{
			file.POST("/upload", fileHandler.Upload)
			file.GET("/list", fileHandler.List)
			file.DELETE("/:id", fileHandler.Delete)
			file.GET("/:id/preview", fileHandler.Preview)
		}

		// 数据分析相关路由
		analysis := api.Group("/analysis")
		analysis.Use(middleware.AuthMiddleware())
		{
			analysis.POST("/query", analysisHandler.Query)
			//analysis.POST("/visualize", analysisHandler.Visualize)
			//analysis.POST("/report", analysisHandler.GenerateReport)
			analysis.GET("/history", analysisHandler.GetHistory)
			analysis.POST("/session", analysisHandler.CreateSession)
			analysis.GET("/session/:id", analysisHandler.GetSession)
		}

		// LLM配置相关路由
		llm := api.Group("/llm")
		llm.Use(middleware.AuthMiddleware())
		{
			llm.POST("/config", analysisHandler.ConfigLLM)
			llm.GET("/config", analysisHandler.GetLLMConfig)
			llm.GET("/usage", analysisHandler.GetUsage)
		}
	}

	// 启动服务器
	log.Printf("Server starting on port %s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
