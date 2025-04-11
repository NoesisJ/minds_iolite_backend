package datasource

import (
	"time"
)

// ColumnType 数据列类型
type ColumnType string

const (
	ColumnTypeString    ColumnType = "string"
	ColumnTypeInteger   ColumnType = "integer"
	ColumnTypeFloat     ColumnType = "float"
	ColumnTypeBoolean   ColumnType = "boolean"
	ColumnTypeDateTime  ColumnType = "datetime"
	ColumnTypeDate      ColumnType = "date"
	ColumnTypeTimestamp ColumnType = "timestamp"
	ColumnTypeArray     ColumnType = "array"
	ColumnTypeObject    ColumnType = "object"
)

// Column 定义数据列的元信息
type Column struct {
	Name        string     `json:"name"`        // 列名
	DisplayName string     `json:"displayName"` // 显示名称
	Type        ColumnType `json:"type"`        // 数据类型
	Required    bool       `json:"required"`    // 是否必填
	Description string     `json:"description"` // 列描述
}

// DataMetadata 数据集元数据
type DataMetadata struct {
	SourceType   string    `json:"sourceType"`   // 数据源类型 (csv, mongodb, mysql)
	SourcePath   string    `json:"sourcePath"`   // 数据源路径
	RowCount     int       `json:"rowCount"`     // 数据行数
	ColumnCount  int       `json:"columnCount"`  // 列数量
	CreatedAt    time.Time `json:"createdAt"`    // 创建时间
	HasHeader    bool      `json:"hasHeader"`    // 是否有表头
	PreviewCount int       `json:"previewCount"` // 预览数据行数
}

// ValidationError 数据验证错误
type ValidationError struct {
	Row     int    `json:"row"`     // 行号
	Column  string `json:"column"`  // 列名
	Message string `json:"message"` // 错误信息
}

// UnifiedDataModel 统一数据模型
// 用于在不同数据源和Agent之间传递数据
type UnifiedDataModel struct {
	Metadata     DataMetadata             `json:"metadata"`     // 元数据
	Columns      []Column                 `json:"columns"`      // 列定义
	Records      []map[string]interface{} `json:"records"`      // 数据记录
	TotalRecords int                      `json:"totalRecords"` // 总记录数
	Errors       []ValidationError        `json:"errors"`       // 验证错误
}

// NewUnifiedDataModel 创建新的统一数据模型
func NewUnifiedDataModel(sourceType, sourcePath string) *UnifiedDataModel {
	return &UnifiedDataModel{
		Metadata: DataMetadata{
			SourceType:   sourceType,
			SourcePath:   sourcePath,
			CreatedAt:    time.Now(),
			PreviewCount: 0,
		},
		Columns:      make([]Column, 0),
		Records:      make([]map[string]interface{}, 0),
		TotalRecords: 0,
		Errors:       make([]ValidationError, 0),
	}
}
