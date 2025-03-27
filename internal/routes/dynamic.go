package routes

import (
	"minds_iolite_backend/internal/database"
	"minds_iolite_backend/internal/services/dynamic"
	metadataService "minds_iolite_backend/internal/services/metadata"

	"github.com/gin-gonic/gin"
)

// SetupDynamicRoutes 设置动态API路由
func SetupDynamicRoutes(router *gin.Engine, db *database.MongoDB) error {
	// 创建元数据服务
	metaService, err := metadataService.NewService(db)
	if err != nil {
		return err
	}

	// 创建动态API生成器
	dynamicGenerator := dynamic.NewGenerator(db, metaService)

	// 定义API路由组
	apiGroup := router.Group("/api")

	// 注册所有模型的动态API
	if err := dynamicGenerator.RegisterAllModelRoutes(apiGroup); err != nil {
		return err
	}

	return nil
}
