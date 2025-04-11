package api

import (
	"minds_iolite_backend/internal/api/handlers"

	"github.com/gin-gonic/gin"
)

// 创建数据源处理器实例
var dataSourceHandler = handlers.NewDataSourceHandler()

func SetupRoutes(r *gin.Engine) {
	v1 := r.Group("/v1")
	{
		datasourceRoutes := v1.Group("/datasource")
		{
			datasourceRoutes.POST("/import-csv", dataSourceHandler.ImportCSVToMongoDB)
			datasourceRoutes.POST("/connect-mongodb", dataSourceHandler.ConnectToMongoDB)
		}
	}
}
