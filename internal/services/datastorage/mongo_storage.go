package datastorage

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"minds_iolite_backend/internal/models/datasource"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoStorage 提供MongoDB存储功能
type MongoStorage struct {
	client   *mongo.Client
	dbName   string
	collName string
}

// MongoDBConnectionInfo 表示MongoDB连接信息
type MongoDBConnectionInfo struct {
	Host        string                           `json:"host"`
	Port        int                              `json:"port"`
	Username    string                           `json:"username"`
	Password    string                           `json:"password"`
	Database    string                           `json:"database"`
	Collections map[string]CollectionInformation `json:"collections"`
}

// CollectionInformation 表示集合信息
type CollectionInformation struct {
	Fields     map[string]string `json:"fields"`
	SampleData string            `json:"sample_data"`
}

// NewMongoStorage 创建新的MongoDB存储服务
func NewMongoStorage(uri string) (*MongoStorage, error) {
	// 创建MongoDB客户端
	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return nil, fmt.Errorf("连接MongoDB失败: %w", err)
	}

	// 验证连接
	if err := client.Ping(context.Background(), nil); err != nil {
		return nil, fmt.Errorf("MongoDB ping失败: %w", err)
	}

	return &MongoStorage{
		client: client,
	}, nil
}

// Close 关闭MongoDB连接
func (s *MongoStorage) Close() error {
	return s.client.Disconnect(context.Background())
}

// ImportCSVToMongoDB 将CSV数据导入MongoDB
func (s *MongoStorage) ImportCSVToMongoDB(data *datasource.UnifiedDataModel, dbName, collName string) (*MongoDBConnectionInfo, error) {
	// 如果没有提供数据库名，使用CSV文件名
	if dbName == "" {
		fileName := filepath.Base(data.Metadata.SourcePath)
		fileNameWithoutExt := strings.TrimSuffix(fileName, filepath.Ext(fileName))
		dbName = "csv_" + fileNameWithoutExt
	}

	// 如果没有提供集合名，使用默认值
	if collName == "" {
		collName = "data"
	}

	// 保存数据库和集合名
	s.dbName = dbName
	s.collName = collName

	// 获取数据库和集合
	db := s.client.Database(dbName)
	collection := db.Collection(collName)

	// 准备要插入的文档
	var documents []interface{}
	for _, record := range data.Records {
		// 添加MongoDB特定的_id字段
		if _, has := record["_id"]; !has {
			record["_id"] = primitive.NewObjectID()
		}
		documents = append(documents, record)
	}

	// 清空集合（如果已存在）
	if err := collection.Drop(context.Background()); err != nil {
		return nil, fmt.Errorf("清空集合失败: %w", err)
	}

	// 批量插入文档
	_, err := collection.InsertMany(context.Background(), documents)
	if err != nil {
		return nil, fmt.Errorf("插入文档失败: %w", err)
	}

	// 生成连接信息
	connInfo, err := s.GenerateConnectionInfo()
	if err != nil {
		return nil, fmt.Errorf("生成连接信息失败: %w", err)
	}

	return connInfo, nil
}

// GenerateConnectionInfo 生成Agent所需的连接信息
func (s *MongoStorage) GenerateConnectionInfo() (*MongoDBConnectionInfo, error) {
	// 获取集合
	db := s.client.Database(s.dbName)
	collection := db.Collection(s.collName)

	// 查询集合中的一个样本文档
	var sampleDoc bson.M
	err := collection.FindOne(context.Background(), bson.M{}).Decode(&sampleDoc)
	if err != nil {
		return nil, fmt.Errorf("获取样本文档失败: %w", err)
	}

	// 获取字段类型信息
	fields := make(map[string]string)
	for key, value := range sampleDoc {
		fields[key] = getMongoType(value)
	}

	// 将样本文档转为JSON字符串
	sampleJSON, err := json.Marshal(sampleDoc)
	if err != nil {
		return nil, fmt.Errorf("转换样本数据失败: %w", err)
	}

	// 创建集合信息
	collectionInfo := CollectionInformation{
		Fields:     fields,
		SampleData: string(sampleJSON),
	}

	// 提取连接信息
	connInfo := &MongoDBConnectionInfo{
		Host:     "localhost",
		Port:     27017,
		Username: "",
		Password: "",
		Database: s.dbName,
		Collections: map[string]CollectionInformation{
			s.collName: collectionInfo,
		},
	}

	return connInfo, nil
}

// getMongoType 根据值确定MongoDB中的类型名称
func getMongoType(value interface{}) string {
	switch value.(type) {
	case primitive.ObjectID:
		return "ObjectId"
	case string:
		return "str"
	case int, int32, int64:
		return "int"
	case float32, float64:
		return "float"
	case bool:
		return "bool"
	case time.Time:
		return "date"
	case primitive.DateTime:
		return "date"
	case bson.A:
		return "array"
	case bson.D, bson.M:
		return "object"
	case primitive.Binary:
		return "binary"
	default:
		return "unknown"
	}
}
