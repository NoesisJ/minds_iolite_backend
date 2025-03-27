package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// 全局变量，用于在应用程序中访问MongoDB连接
var (
	// Client 是MongoDB的全局客户端连接
	Client *mongo.Client

	// Database 是默认的MongoDB数据库实例
	Database *mongo.Database
)

// MongoDB 结构体封装了MongoDB连接相关功能
type MongoDB struct {
	// Client 是MongoDB客户端连接
	Client *mongo.Client

	// Database 是当前使用的数据库
	Database *mongo.Database

	// URI 是MongoDB连接字符串
	URI string

	// DBName 是数据库名称
	DBName string
}

// Config 表示MongoDB配置选项
type Config struct {
	// URI 是MongoDB连接字符串，例如"mongodb://localhost:27017"
	URI string

	// DBName 是要使用的数据库名称
	DBName string

	// Timeout 是连接超时时间
	Timeout time.Duration

	// MaxPoolSize 是连接池的最大大小
	MaxPoolSize uint64
}

// NewMongoDB 创建并初始化一个新的MongoDB连接
// 参数:
//   - cfg: MongoDB连接配置
//
// 返回:
//   - MongoDB实例和可能的错误
func NewMongoDB(cfg Config) (*MongoDB, error) {
	// 如果没有设置超时，使用默认10秒
	if cfg.Timeout == 0 {
		cfg.Timeout = 10 * time.Second
	}

	// 创建上下文，用于控制连接超时
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	// 准备MongoDB客户端选项
	clientOptions := options.Client().ApplyURI(cfg.URI)

	// 如果设置了连接池大小，则应用
	if cfg.MaxPoolSize > 0 {
		clientOptions.SetMaxPoolSize(cfg.MaxPoolSize)
	}

	// 连接到MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("连接MongoDB失败: %w", err)
	}

	// 验证连接是否成功
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		return nil, fmt.Errorf("无法ping MongoDB: %w", err)
	}

	// 获取数据库实例
	database := client.Database(cfg.DBName)

	// 设置全局变量，方便其他包使用
	Client = client
	Database = database

	log.Printf("已成功连接到MongoDB数据库: %s", cfg.DBName)

	return &MongoDB{
		Client:   client,
		Database: database,
		URI:      cfg.URI,
		DBName:   cfg.DBName,
	}, nil
}

// Close 关闭MongoDB连接
// 应在应用程序结束时调用此方法以释放资源
func (m *MongoDB) Close() error {
	if m.Client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := m.Client.Disconnect(ctx); err != nil {
			return fmt.Errorf("关闭MongoDB连接失败: %w", err)
		}
		log.Print("MongoDB连接已关闭")
	}
	return nil
}

// Collection 获取指定名称的集合
// 参数:
//   - name: 集合名称
//
// 返回:
//   - 集合对象
func (m *MongoDB) Collection(name string) *mongo.Collection {
	return m.Database.Collection(name)
}

// EnsureCollection 确保集合存在，如不存在则创建
// 参数:
//   - name: 集合名称
//   - opts: 创建集合的选项
//
// 返回:
//   - 可能出现的错误
func (m *MongoDB) EnsureCollection(name string, opts *options.CreateCollectionOptions) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 获取所有集合名称
	collections, err := m.Database.ListCollectionNames(ctx, map[string]interface{}{"name": name})
	if err != nil {
		return fmt.Errorf("列出集合失败: %w", err)
	}

	// 检查集合是否已存在
	for _, coll := range collections {
		if coll == name {
			// 集合已存在，无需创建
			return nil
		}
	}

	// 创建新集合
	err = m.Database.CreateCollection(ctx, name, opts)
	if err != nil {
		return fmt.Errorf("创建集合 %s 失败: %w", name, err)
	}

	log.Printf("已成功创建集合: %s", name)
	return nil
}

// DropCollection 删除指定的集合
// 警告：此操作将永久删除集合中的所有数据
// 参数:
//   - name: 要删除的集合名称
//
// 返回:
//   - 可能出现的错误
func (m *MongoDB) DropCollection(name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 获取集合
	coll := m.Database.Collection(name)

	// 删除集合
	if err := coll.Drop(ctx); err != nil {
		return fmt.Errorf("删除集合 %s 失败: %w", name, err)
	}

	log.Printf("已成功删除集合: %s", name)
	return nil
}

// CollectionExists 检查集合是否存在
// 参数:
//   - name: 集合名称
//
// 返回:
//   - 存在返回true，否则返回false
//   - 可能出现的错误
func (m *MongoDB) CollectionExists(name string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collections, err := m.Database.ListCollectionNames(ctx, map[string]interface{}{"name": name})
	if err != nil {
		return false, fmt.Errorf("列出集合失败: %w", err)
	}

	for _, coll := range collections {
		if coll == name {
			return true, nil
		}
	}

	return false, nil
}
