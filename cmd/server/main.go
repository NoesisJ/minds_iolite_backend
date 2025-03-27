package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"minds_iolite_backend/config"
	"minds_iolite_backend/internal/database"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 尝试直接使用已知可工作的连接方式
	mongoConfig := database.Config{
		URI:         "mongodb://localhost:27017/?directConnection=true", // 使用测试程序中成功的连接字符串
		DBName:      cfg.MongoDB.Database,
		Timeout:     20 * time.Second, // 增加超时时间
		MaxPoolSize: cfg.MongoDB.MaxPoolSize,
	}

	log.Printf("尝试连接到 MongoDB: %s", mongoConfig.URI)

	mongoDB, err := database.NewMongoDB(mongoConfig)
	if err != nil {
		log.Fatalf("连接MongoDB失败: %v", err)
	}

	log.Printf("MongoDB 连接成功!")

	// 确保在程序结束时关闭MongoDB连接
	defer mongoDB.Close()

	// 创建Gin路由引擎
	router := gin.Default()

	// 设置API路由 - 这部分会在后续实现
	// routes.SetupRoutes(router, mongoDB)

	// 启动HTTP服务器
	go func() {
		if err := router.Run(cfg.Server.Address); err != nil {
			log.Fatalf("启动服务器失败: %v", err)
		}
	}()

	log.Printf("服务器已启动，监听地址: %s", cfg.Server.Address)

	// 等待中断信号优雅地关闭服务器
	quit := make(chan os.Signal, 1)
	// kill (无参数) 默认发送 syscall.SIGTERM
	// kill -2 发送 syscall.SIGINT
	// kill -9 发送 syscall.SIGKILL，但无法被捕获，所以不需要添加
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("正在关闭服务器...")
}
