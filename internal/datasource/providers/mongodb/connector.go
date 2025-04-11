package mongodb

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"minds_iolite_backend/internal/services/datastorage"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDBConnector MongoDB连接器
type MongoDBConnector struct {
	client *mongo.Client
	uri    string
}

// NewMongoDBConnector 创建MongoDB连接器
func NewMongoDBConnector(uri string) (*MongoDBConnector, error) {
	// 创建上下文，设置超时
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 创建连接选项
	clientOptions := options.Client().ApplyURI(uri)

	// 连接MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("连接MongoDB失败: %w", err)
	}

	// 测试连接
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("MongoDB服务器无响应: %w", err)
	}

	return &MongoDBConnector{
		client: client,
		uri:    uri,
	}, nil
}

// Close 关闭连接
func (c *MongoDBConnector) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return c.client.Disconnect(ctx)
}

// ExtractConnectionInfo 提取数据库连接信息
func (c *MongoDBConnector) ExtractConnectionInfo(dbName string) (*datastorage.MongoDBConnectionInfo, error) {
	// 创建上下文
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 获取数据库
	db := c.client.Database(dbName)

	// 创建连接信息
	connInfo := &datastorage.MongoDBConnectionInfo{
		Host:        "localhost", // 假设是本地MongoDB
		Port:        27017,       // 默认端口
		Username:    "",          // 本地通常无用户名
		Password:    "",          // 密码隐藏
		Database:    dbName,
		Collections: make(map[string]datastorage.CollectionInformation),
	}

	// 获取集合列表
	collections, err := db.ListCollectionNames(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("获取集合列表失败: %w", err)
	}

	// 若没有集合，返回错误
	if len(collections) == 0 {
		return nil, fmt.Errorf("数据库 %s 中没有集合", dbName)
	}

	// 处理每个集合
	for _, collName := range collections {
		// 获取集合
		coll := db.Collection(collName)

		// 获取样本文档
		var sampleDoc bson.M
		err := coll.FindOne(ctx, bson.M{}).Decode(&sampleDoc)
		if err != nil {
			// 如果集合为空，跳过
			if err == mongo.ErrNoDocuments {
				continue
			}
			return nil, fmt.Errorf("获取集合 %s 的样本文档失败: %w", collName, err)
		}

		// 获取字段类型信息
		fields := make(map[string]string)
		for key, value := range sampleDoc {
			fields[key] = getMongoType(value)
		}

		// 将样本文档转换为JSON字符串
		sampleJSON, err := json.Marshal(sampleDoc)
		if err != nil {
			return nil, fmt.Errorf("转换样本数据失败: %w", err)
		}

		// 添加集合信息
		connInfo.Collections[collName] = datastorage.CollectionInformation{
			Fields:     fields,
			SampleData: string(sampleJSON),
		}
	}

	return connInfo, nil
}

// getMongoType 获取MongoDB字段类型名称
func getMongoType(value interface{}) string {
	switch value.(type) {
	case nil:
		return "null"
	case bool:
		return "bool"
	case int, int32, int64:
		return "int"
	case float32, float64:
		return "float"
	case string:
		return "str"
	case time.Time:
		return "date"
	case bson.D:
		return "object"
	case bson.A:
		return "array"
	case bson.M:
		return "object"
	case primitive.ObjectID:
		return "ObjectId"
	default:
		return "unknown"
	}
}
