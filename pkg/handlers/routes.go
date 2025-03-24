package handlers

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// SetupRoutes 配置所有API路由
func SetupRoutes(r *gin.Engine) {
	// 使用gin-contrib/cors库配置CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "Authorization", "Accept", "Cache-Control", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// API路由组
	api := r.Group("/api")

	// 数据相关路由
	api.GET("/data", GetAllData)

	// 财务相关路由
	api.GET("/financial", GetAllFinancial)
	api.POST("/financial", CreateFinancial)
	api.DELETE("/financial/batch", BatchDeleteFinancial)
	api.GET("/financial/export", ExportFinancialCSV)
	api.GET("/financial/:id", GetFinancialByID)
	api.PUT("/financial/:id", UpdateFinancial)
	api.DELETE("/financial/:id", DeleteFinancial)
	api.GET("/health", CheckFinancialAPIHealth)
}
