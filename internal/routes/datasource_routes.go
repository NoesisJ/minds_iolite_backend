package routes

import (
	"minds_iolite_backend/internal/api/handlers"
	sessionHandlers "minds_iolite_backend/internal/handlers"

	"github.com/gin-gonic/gin"
)

// SetupDataSourceRoutes 设置数据源相关路由
func SetupDataSourceRoutes(router *gin.Engine) {
	// 创建数据源处理器
	dataSourceHandler := handlers.NewDataSourceHandler()

	// 创建会话处理器并初始化会话管理器
	sessionHandler := sessionHandlers.NewSessionHandler()
	sessionHandlers.InitSessionManager()

	// 数据源API路由组
	dataSourceGroup := router.Group("/api/datasource")
	{
		// CSV相关API
		csvGroup := dataSourceGroup.Group("/csv")
		{
			// 处理CSV文件
			csvGroup.POST("/process", dataSourceHandler.ProcessCSVFile)

			// 获取CSV列类型
			csvGroup.POST("/column-types", dataSourceHandler.GetColumnTypes)

			// 上传CSV文件
			csvGroup.POST("/upload", dataSourceHandler.UploadCSVFile)

			// 导入CSV到MongoDB
			csvGroup.POST("/import-to-mongo", dataSourceHandler.ImportCSVToMongoDB)
		}

		// TODO: 添加MongoDB数据源相关路由
		mongoGroup := dataSourceGroup.Group("/mongodb")
		{
			// 连接到MongoDB
			mongoGroup.POST("/connect", dataSourceHandler.ConnectToMongoDB)
		}

		// TODO: 添加MySQL数据源相关路由
		mysqlGroup := dataSourceGroup.Group("/mysql")
		{
			// 连接到MySQL
			mysqlGroup.POST("/connect", dataSourceHandler.ConnectToMySQL)
		}

		// SQLite数据源相关路由
		sqliteGroup := dataSourceGroup.Group("/sqlite")
		{
			// 处理SQLite文件
			sqliteGroup.POST("/process", dataSourceHandler.ProcessSQLiteFile)

			// 导入SQLite到MongoDB
			sqliteGroup.POST("/import-to-mongo", dataSourceHandler.ImportSQLiteToMongoDB)
		}
	}

	// 持久会话API路由组
	sessionsGroup := router.Group("/api/sessions")
	{
		// 创建会话
		sessionsGroup.POST("", sessionHandler.CreateSession)

		// 获取所有会话
		sessionsGroup.GET("", sessionHandler.GetAllSessions)

		// 获取特定会话
		sessionsGroup.GET("/:sessionId", sessionHandler.GetSession)

		// 刷新会话
		sessionsGroup.PUT("/:sessionId/refresh", sessionHandler.RefreshSession)

		// 关闭会话
		sessionsGroup.DELETE("/:sessionId", sessionHandler.CloseSession)
	}
}
