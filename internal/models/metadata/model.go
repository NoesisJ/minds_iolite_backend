package metadata

import (
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ModelDefinition 表示一个数据模型的定义
// 这是元数据的核心结构，描述了一个业务对象（如"客户"、"订单"等）在系统中如何表示
// 每个ModelDefinition在MongoDB中会对应创建一个实际的集合来存储该类型的数据
type ModelDefinition struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"` // MongoDB的唯一标识符
	Name        string             `bson:"name" json:"name"`                  // 模型唯一名称，同时也用作API路径
	DisplayName string             `bson:"display_name" json:"displayName"`   // 用于UI显示的友好名称
	Description string             `bson:"description" json:"description"`    // 模型的详细描述
	Collection  string             `bson:"collection" json:"collection"`      // 数据存储的MongoDB集合名
	Fields      []FieldDefinition  `bson:"fields" json:"fields"`              // 模型包含的所有字段定义
	Indexes     []IndexDefinition  `bson:"indexes" json:"indexes,omitempty"`  // 需要创建的索引定义
	CreatedAt   time.Time          `bson:"created_at" json:"createdAt"`       // 模型创建时间
	UpdatedAt   time.Time          `bson:"updated_at" json:"updatedAt"`       // 模型最后更新时间
	IsSystem    bool               `bson:"is_system" json:"isSystem"`         // 是否为系统内置模型，系统模型不可删除
}

// IndexDefinition 定义模型的索引
// 索引用于优化查询性能和保证数据唯一性
type IndexDefinition struct {
	Fields []string `bson:"fields" json:"fields"` // 构成索引的字段列表
	Unique bool     `bson:"unique" json:"unique"` // 是否是唯一索引
	Name   string   `bson:"name" json:"name"`     // 索引名称
}

// Validate 验证模型定义是否有效
// 在保存模型前会调用此方法进行验证
func (m *ModelDefinition) Validate() error {
	// 检查模型名称是否为空
	if m.Name == "" {
		return fmt.Errorf("模型名称不能为空")
	}

	// 检查是否有字段定义
	if len(m.Fields) == 0 {
		return fmt.Errorf("模型必须至少有一个字段")
	}

	// 检查字段名称是否重复
	fieldNames := make(map[string]bool)
	for _, field := range m.Fields {
		if _, exists := fieldNames[field.Name]; exists {
			return fmt.Errorf("字段名称 '%s' 重复", field.Name)
		}
		fieldNames[field.Name] = true

		// 验证字段定义
		if err := field.Validate(); err != nil {
			return fmt.Errorf("字段 '%s' 无效: %w", field.Name, err)
		}
	}

	return nil
}
