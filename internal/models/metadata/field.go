package metadata

import (
	"fmt"
	"strings"
)

// FieldType 表示字段的数据类型
// 这决定了字段存储的数据类型和验证规则
type FieldType string

// 支持的字段类型
// 每种类型有特定的存储和验证逻辑
const (
	FieldTypeString    FieldType = "string"    // 普通文本
	FieldTypeNumber    FieldType = "number"    // 浮点数
	FieldTypeInteger   FieldType = "integer"   // 整数
	FieldTypeBoolean   FieldType = "boolean"   // 布尔值
	FieldTypeDate      FieldType = "date"      // 日期（不包含时间）
	FieldTypeDateTime  FieldType = "datetime"  // 日期时间
	FieldTypeObject    FieldType = "object"    // 嵌套对象
	FieldTypeArray     FieldType = "array"     // 数组
	FieldTypeReference FieldType = "reference" // 引用其他模型的ID
	FieldTypeFile      FieldType = "file"      // 文件
	FieldTypeImage     FieldType = "image"     // 图片
	FieldTypeEnum      FieldType = "enum"      // 枚举选项
)

// FieldDefinition 表示模型中的一个字段定义
// 描述数据中单个属性的所有元信息
type FieldDefinition struct {
	Name         string                 `bson:"name" json:"name"`                  // 字段编程名称
	DisplayName  string                 `bson:"display_name" json:"displayName"`   // UI显示名称
	Type         FieldType              `bson:"type" json:"type"`                  // 字段数据类型
	Required     bool                   `bson:"required" json:"required"`          // 是否必填
	Unique       bool                   `bson:"unique" json:"unique"`              // 是否唯一
	DefaultValue interface{}            `bson:"default_value" json:"defaultValue"` // 默认值
	Validators   []Validator            `bson:"validators" json:"validators"`      // 验证器列表
	Properties   map[string]interface{} `bson:"properties" json:"properties"`      // 类型特定的附加属性
}

// Validator 表示一个字段验证器
// 用于对字段值进行额外的验证检查
type Validator struct {
	Type   string                 `bson:"type" json:"type"`     // 验证器类型，如"email"、"regex"等
	Params map[string]interface{} `bson:"params" json:"params"` // 验证器参数
}

// Validate 验证字段定义是否有效
func (f *FieldDefinition) Validate() error {
	// 检查字段名称
	if f.Name == "" {
		return fmt.Errorf("字段名称不能为空")
	}

	// 字段名称只能包含字母、数字和下划线，且必须以字母开头
	if !isValidFieldName(f.Name) {
		return fmt.Errorf("字段名称 '%s' 无效，必须以字母开头且只能包含字母、数字和下划线", f.Name)
	}

	// 验证字段类型
	switch f.Type {
	case FieldTypeString, FieldTypeNumber, FieldTypeInteger, FieldTypeBoolean,
		FieldTypeDate, FieldTypeDateTime, FieldTypeObject, FieldTypeArray,
		FieldTypeReference, FieldTypeFile, FieldTypeImage, FieldTypeEnum:
		// 有效的字段类型
	default:
		return fmt.Errorf("字段类型 '%s' 无效", f.Type)
	}

	// 针对不同类型的特殊验证
	switch f.Type {
	case FieldTypeReference:
		// 检查引用模型是否指定
		if model, ok := f.Properties["refModel"]; !ok || model == "" {
			return fmt.Errorf("引用类型字段必须指定引用模型")
		}
	case FieldTypeEnum:
		// 检查枚举值是否指定
		if options, ok := f.Properties["options"]; !ok || options == nil {
			return fmt.Errorf("枚举类型字段必须指定选项列表")
		}
	}

	return nil
}

// isValidFieldName 验证字段名称格式
func isValidFieldName(name string) bool {
	if len(name) == 0 || (name[0] < 'a' || name[0] > 'z') && (name[0] < 'A' || name[0] > 'Z') {
		return false
	}

	for _, c := range name {
		if (c < 'a' || c > 'z') && (c < 'A' || c > 'Z') && (c < '0' || c > '9') && c != '_' {
			return false
		}
	}

	// 检查是否是保留字
	reservedWords := []string{"id", "_id", "createdAt", "updatedAt"}
	for _, word := range reservedWords {
		if strings.ToLower(name) == word {
			return false
		}
	}

	return true
}
