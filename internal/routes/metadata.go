package routes

import (
	"fmt"
	"net/http"

	"minds_iolite_backend/internal/database"
	"minds_iolite_backend/internal/models/metadata"
	metadataService "minds_iolite_backend/internal/services/metadata"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

// SetupMetadataRoutes 设置元数据管理路由
func SetupMetadataRoutes(router *gin.Engine, db *database.MongoDB) error {
	fmt.Println("设置元数据路由...")

	// 创建元数据服务
	metaService, err := metadataService.NewService(db)
	if err != nil {
		return err
	}

	// 元数据路由组
	metaGroup := router.Group("/metadata")
	fmt.Println("已注册 /metadata 路由组")
	{
		// 模型相关API
		modelsGroup := metaGroup.Group("/models")
		{
			// 创建模型
			modelsGroup.POST("", func(c *gin.Context) {
				var model metadata.ModelDefinition
				if err := c.ShouldBindJSON(&model); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "无效的模型定义"})
					return
				}

				if err := metaService.CreateModel(&model); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}

				c.JSON(http.StatusCreated, model)
			})

			// 获取模型列表
			modelsGroup.GET("", func(c *gin.Context) {
				// 默认不包含系统模型
				includeSystem := false
				if c.Query("includeSystem") == "true" {
					includeSystem = true
				}

				var filter bson.M
				if !includeSystem {
					filter = bson.M{"is_system": false}
				}

				models, err := metaService.ListModels(filter)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "获取模型列表失败"})
					return
				}

				c.JSON(http.StatusOK, models)
			})

			// 获取单个模型
			modelsGroup.GET("/:id", func(c *gin.Context) {
				id := c.Param("id")
				model, err := metaService.GetModelByID(id)
				if err != nil {
					c.JSON(http.StatusNotFound, gin.H{"error": "模型不存在"})
					return
				}

				c.JSON(http.StatusOK, model)
			})

			// 更新模型
			modelsGroup.PUT("/:id", func(c *gin.Context) {
				id := c.Param("id")
				var updates metadata.ModelDefinition
				if err := c.ShouldBindJSON(&updates); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "无效的模型定义"})
					return
				}

				if err := metaService.UpdateModel(id, &updates); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}

				// 获取更新后的模型
				model, err := metaService.GetModelByID(id)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "获取更新后的模型失败"})
					return
				}

				c.JSON(http.StatusOK, model)
			})

			// 删除模型
			modelsGroup.DELETE("/:id", func(c *gin.Context) {
				id := c.Param("id")
				if err := metaService.DeleteModel(id); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}

				c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
			})
		}
	}

	return nil
}
