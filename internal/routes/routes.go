package routes

import (
	"minds_iolite_backend/internal/database"

	"fmt"

	"github.com/gin-gonic/gin"
)

// SetupRoutes 设置所有路由
func SetupRoutes(router *gin.Engine, db *database.MongoDB) error {
	// 添加调试输出
	fmt.Println("正在设置路由...")

	// 设置元数据管理路由
	if err := SetupMetadataRoutes(router, db); err != nil {
		return err
	}
	fmt.Println("元数据路由设置完成")

	// 设置动态API路由
	if err := SetupDynamicRoutes(router, db); err != nil {
		return err
	}
	fmt.Println("动态API路由设置完成")

	return nil
}
