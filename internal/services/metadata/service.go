package metadata

import (
	"context"
	"fmt"
	"time"

	"minds_iolite_backend/internal/database"
	"minds_iolite_backend/internal/models/metadata"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	// ModelCollectionName 是存储模型定义的集合名称
	ModelCollectionName = "models"
)

// Service 提供元数据管理功能
// 这个服务负责管理模型定义及其相关数据库结构
// 它是低代码平台的核心服务之一，使平台能够动态创建和管理数据模型
type Service struct {
	db *database.MongoDB // MongoDB数据库连接
}

// NewService 创建一个新的元数据服务实例
// 初始化过程中会确保必要的系统集合存在
func NewService(db *database.MongoDB) (*Service, error) {
	service := &Service{
		db: db,
	}

	// 确保模型集合存在，这是存储所有模型定义的地方
	if err := service.ensureModelCollection(); err != nil {
		return nil, err
	}

	return service, nil
}

// ensureModelCollection 确保模型集合存在并有正确的索引
func (s *Service) ensureModelCollection() error {
	// 检查集合是否存在
	exists, err := s.db.CollectionExists(ModelCollectionName)
	if err != nil {
		return fmt.Errorf("检查模型集合失败: %w", err)
	}

	// 如果集合不存在，创建它
	if !exists {
		if err := s.db.EnsureCollection(ModelCollectionName, nil); err != nil {
			return fmt.Errorf("创建模型集合失败: %w", err)
		}
	}

	// 创建索引以确保模型名称唯一
	collection := s.db.Collection(ModelCollectionName)
	_, err = collection.Indexes().CreateOne(
		context.Background(),
		mongo.IndexModel{
			Keys:    bson.D{{Key: "name", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	)
	if err != nil {
		return fmt.Errorf("创建模型名称索引失败: %w", err)
	}

	return nil
}

// CreateModel 创建一个新的模型定义
// 此方法包含以下步骤:
// 1. 验证模型定义的完整性和正确性
// 2. 设置创建和更新时间戳
// 3. 保存模型定义到数据库
// 4. 创建对应的数据集合和索引
func (s *Service) CreateModel(model *metadata.ModelDefinition) error {
	// 验证模型
	if err := model.Validate(); err != nil {
		return err
	}

	// 设置创建和更新时间
	now := time.Now()
	model.CreatedAt = now
	model.UpdatedAt = now

	// 确保集合名称被设置，默认使用模型名称
	if model.Collection == "" {
		model.Collection = model.Name
	}

	// 保存到数据库的模型集合中
	collection := s.db.Collection(ModelCollectionName)
	_, err := collection.InsertOne(context.Background(), model)
	if err != nil {
		return fmt.Errorf("保存模型定义失败: %w", err)
	}

	// 如果不是系统模型，创建对应的集合用于存储实际数据
	if !model.IsSystem {
		if err := s.createModelCollection(model); err != nil {
			return fmt.Errorf("创建模型集合失败: %w", err)
		}
	}

	return nil
}

// createModelCollection 为模型创建对应的集合和索引
// 此方法根据模型定义在MongoDB中创建实际存储数据的集合和必要的索引
// 这使得每个模型有自己专用的数据存储结构
func (s *Service) createModelCollection(model *metadata.ModelDefinition) error {
	// 创建集合
	if err := s.db.EnsureCollection(model.Collection, nil); err != nil {
		return err
	}

	// 创建模型中定义的索引
	collection := s.db.Collection(model.Collection)
	for _, index := range model.Indexes {
		indexKeys := bson.D{}
		for _, field := range index.Fields {
			indexKeys = append(indexKeys, bson.E{Key: field, Value: 1})
		}

		_, err := collection.Indexes().CreateOne(
			context.Background(),
			mongo.IndexModel{
				Keys:    indexKeys,
				Options: options.Index().SetUnique(index.Unique).SetName(index.Name),
			},
		)
		if err != nil {
			return fmt.Errorf("创建索引 '%s' 失败: %w", index.Name, err)
		}
	}

	// 为所有标记为unique的字段自动创建唯一索引
	for _, field := range model.Fields {
		if field.Unique {
			_, err := collection.Indexes().CreateOne(
				context.Background(),
				mongo.IndexModel{
					Keys:    bson.D{{Key: field.Name, Value: 1}},
					Options: options.Index().SetUnique(true).SetName(fmt.Sprintf("%s_unique", field.Name)),
				},
			)
			if err != nil {
				return fmt.Errorf("为字段 '%s' 创建唯一索引失败: %w", field.Name, err)
			}
		}
	}

	return nil
}

// GetModelByID 根据ID获取模型定义
func (s *Service) GetModelByID(id string) (*metadata.ModelDefinition, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("无效的模型ID: %w", err)
	}

	collection := s.db.Collection(ModelCollectionName)
	var model metadata.ModelDefinition
	err = collection.FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&model)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("模型不存在")
		}
		return nil, fmt.Errorf("获取模型失败: %w", err)
	}

	return &model, nil
}

// GetModelByName 根据名称获取模型定义
func (s *Service) GetModelByName(name string) (*metadata.ModelDefinition, error) {
	collection := s.db.Collection(ModelCollectionName)
	var model metadata.ModelDefinition
	err := collection.FindOne(context.Background(), bson.M{"name": name}).Decode(&model)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("模型 '%s' 不存在", name)
		}
		return nil, fmt.Errorf("获取模型失败: %w", err)
	}

	return &model, nil
}

// ListModels 列出所有模型定义
func (s *Service) ListModels(filter bson.M) ([]*metadata.ModelDefinition, error) {
	collection := s.db.Collection(ModelCollectionName)

	// 如果没有提供过滤条件，使用空过滤器
	if filter == nil {
		filter = bson.M{}
	}

	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		return nil, fmt.Errorf("查询模型列表失败: %w", err)
	}
	defer cursor.Close(context.Background())

	var models []*metadata.ModelDefinition
	if err := cursor.All(context.Background(), &models); err != nil {
		return nil, fmt.Errorf("解析模型列表失败: %w", err)
	}

	return models, nil
}

// UpdateModel 更新模型定义
func (s *Service) UpdateModel(id string, updates *metadata.ModelDefinition) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("无效的模型ID: %w", err)
	}

	// 获取原始模型
	original, err := s.GetModelByID(id)
	if err != nil {
		return err
	}

	// 验证更新后的模型
	if err := updates.Validate(); err != nil {
		return err
	}

	// 保留原始的创建时间和ID
	updates.CreatedAt = original.CreatedAt
	updates.ID = objectID
	updates.UpdatedAt = time.Now()

	// 保存更新
	collection := s.db.Collection(ModelCollectionName)
	_, err = collection.ReplaceOne(
		context.Background(),
		bson.M{"_id": objectID},
		updates,
	)
	if err != nil {
		return fmt.Errorf("更新模型失败: %w", err)
	}

	// 如果集合名称发生变化，需要重命名集合
	if !updates.IsSystem && original.Collection != updates.Collection {
		// MongoDB没有直接的重命名集合API，所以需要创建新集合，复制数据，然后删除旧集合
		// 这里简化处理，直接创建新集合
		if err := s.createModelCollection(updates); err != nil {
			return fmt.Errorf("为更新后的模型创建集合失败: %w", err)
		}
	} else if !updates.IsSystem {
		// 更新集合索引
		if err := s.createModelCollection(updates); err != nil {
			return fmt.Errorf("更新模型集合索引失败: %w", err)
		}
	}

	return nil
}

// DeleteModel 删除模型定义
func (s *Service) DeleteModel(id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("无效的模型ID: %w", err)
	}

	// 获取原始模型
	model, err := s.GetModelByID(id)
	if err != nil {
		return err
	}

	// 禁止删除系统模型
	if model.IsSystem {
		return fmt.Errorf("不能删除系统模型")
	}

	// 删除模型定义
	collection := s.db.Collection(ModelCollectionName)
	_, err = collection.DeleteOne(context.Background(), bson.M{"_id": objectID})
	if err != nil {
		return fmt.Errorf("删除模型定义失败: %w", err)
	}

	// 删除对应的集合
	if err := s.db.DropCollection(model.Collection); err != nil {
		return fmt.Errorf("删除模型集合失败: %w", err)
	}

	return nil
}
