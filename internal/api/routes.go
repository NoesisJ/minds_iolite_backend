package api

import (
	"minds_iolite_backend/internal/api/handlers"

	"github.com/gin-gonic/gin"
)

// 创建处理器实例
var dataSourceHandler = handlers.NewDataSourceHandler()
var sessionHandler = handlers.NewSessionHandler()

func SetupRoutes(r *gin.Engine) {
	// 初始化会话管理器 - 放在最前面确保在使用前初始化
	handlers.InitSessionManager()

	// 添加CORS中间件
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// API路由
	apiGroup := r.Group("/api")
	{
		// 数据源API
		datasourceGroup := apiGroup.Group("/datasource")
		{
			// CSV相关API
			csvGroup := datasourceGroup.Group("/csv")
			{
				csvGroup.POST("/process", dataSourceHandler.ProcessCSVFile)
				csvGroup.POST("/column-types", dataSourceHandler.GetColumnTypes)
				csvGroup.POST("/upload", dataSourceHandler.UploadCSVFile)
				csvGroup.POST("/import-to-mongo", dataSourceHandler.ImportCSVToMongoDB)
			}

			// MongoDB相关API
			mongoGroup := datasourceGroup.Group("/mongodb")
			{
				mongoGroup.POST("/connect", dataSourceHandler.ConnectToMongoDB)
			}

			// MySQL相关API
			mysqlGroup := datasourceGroup.Group("/mysql")
			{
				mysqlGroup.POST("/connect", dataSourceHandler.ConnectToMySQL)
			}

			// SQLite相关API
			sqliteGroup := datasourceGroup.Group("/sqlite")
			{
				sqliteGroup.POST("/process", dataSourceHandler.ProcessSQLiteFile)
				sqliteGroup.POST("/import-to-mongo", dataSourceHandler.ImportSQLiteToMongoDB)
			}
		}

		// 持久会话API
		sessions := apiGroup.Group("/sessions")
		{
			sessions.POST("", sessionHandler.CreateSession)                    // 创建会话
			sessions.GET("", sessionHandler.GetAllSessions)                    // 获取所有会话
			sessions.GET("/:sessionId", sessionHandler.GetSession)             // 获取指定会话
			sessions.PUT("/:sessionId/refresh", sessionHandler.RefreshSession) // 刷新会话
			sessions.DELETE("/:sessionId", sessionHandler.CloseSession)        // 关闭会话
		}
	}
}
