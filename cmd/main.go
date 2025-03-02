package main

import (
	"fmt"
	"log"

	"github.com/NoesisJ/minds_iolite/backend/pkg/config"
	"github.com/NoesisJ/minds_iolite/backend/pkg/database"
	"github.com/NoesisJ/minds_iolite/backend/pkg/models"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化数据库
	db, err := database.InitDB(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// 自动迁移模型
	if err := models.AutoMigrate(db); err != nil {
		log.Fatalf("Failed to auto migrate: %v", err)
	}

	// 初始化Gin
	r := gin.Default()

	// 添加路由
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	// 启动服务
	fmt.Println("Server running on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
