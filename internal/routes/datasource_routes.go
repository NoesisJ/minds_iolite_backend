package routes

import (
	"minds_iolite_backend/internal/api/handlers"

	"github.com/gin-gonic/gin"
)

// SetupDataSourceRoutes 设置数据源相关路由
func SetupDataSourceRoutes(router *gin.Engine) {
	// 创建数据源处理器
	dataSourceHandler := handlers.NewDataSourceHandler()

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
}
