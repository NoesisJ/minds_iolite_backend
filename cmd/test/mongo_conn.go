package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func main() {
	// 尝试各种连接字符串
	uris := []string{
		"mongodb://localhost:27017",
		"mongodb://localhost:27017/?directConnection=true",
		"mongodb://127.0.0.1:27017",
		"mongodb://127.0.0.1:27017/?directConnection=true",
	}

	for _, uri := range uris {
		fmt.Printf("\n尝试连接: %s\n", uri)

		// 创建上下文，10秒超时
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// 创建连接选项
		clientOptions := options.Client().ApplyURI(uri)

		// 尝试连接
		client, err := mongo.Connect(ctx, clientOptions)
		if err != nil {
			log.Printf("连接失败: %v\n", err)
			continue
		}

		// 尝试ping
		err = client.Ping(ctx, readpref.Primary())
		if err != nil {
			log.Printf("Ping失败: %v\n", err)
		} else {
			log.Printf("连接成功！\n")
		}

		// 关闭连接
		if client != nil {
			client.Disconnect(ctx)
		}
	}
}
