package datastorage

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"

	"minds_iolite_backend/internal/datasource/providers/mongodb"
	"minds_iolite_backend/internal/models/datasource"

	_ "github.com/mattn/go-sqlite3"
	"go.mongodb.org/mongo-driver/mongo"
)

// SQLiteStorage 提供SQLite存储功能
type SQLiteStorage struct {
	db       *sql.DB
	filePath string
}

// SQLiteConnectionInfo 表示SQLite连接信息
type SQLiteConnectionInfo struct {
	FilePath  string                      `json:"filePath"`
	TableInfo map[string]TableInformation `json:"tables"`
}

// NewSQLiteStorage 创建新的SQLite存储服务
func NewSQLiteStorage(filePath string) (*SQLiteStorage, error) {
	// 连接SQLite数据库
	db, err := sql.Open("sqlite3", filePath)
	if err != nil {
		return nil, fmt.Errorf("连接SQLite数据库失败: %w", err)
	}

	// 测试连接
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("SQLite数据库无响应: %w", err)
	}

	return &SQLiteStorage{
		db:       db,
		filePath: filePath,
	}, nil
}

// Close 关闭连接
func (s *SQLiteStorage) Close() error {
	return s.db.Close()
}

// ImportSQLiteToMongoDB 将SQLite数据导入MongoDB
func (s *SQLiteStorage) ImportSQLiteToMongoDB(tableName, dbName, collName string, mongoURI string) (*mongodb.MongoDBConnectionInfo, error) {
	// 如果没有提供数据库名，使用SQLite文件名
	if dbName == "" {
		fileName := filepath.Base(s.filePath)
		fileNameWithoutExt := strings.TrimSuffix(fileName, filepath.Ext(fileName))
		dbName = "sqlite_" + fileNameWithoutExt
	}

	// 如果没有提供集合名，使用表名
	if collName == "" {
		collName = tableName
	}

	// 创建MongoDB连接器
	connector, err := mongodb.NewMongoDBConnector(mongoURI)
	if err != nil {
		return nil, fmt.Errorf("连接MongoDB失败: %w", err)
	}
	defer connector.Close()

	// 获取MongoDB客户端和集合
	client := connector.GetClient()
	coll := client.Database(dbName).Collection(collName)

	// 先清空集合
	if err := coll.Drop(nil); err != nil && err != mongo.ErrNilDocument {
		return nil, fmt.Errorf("清空集合失败: %w", err)
	}

	// 获取SQLite表的全部数据
	rows, err := s.db.Query(fmt.Sprintf("SELECT * FROM '%s'", tableName))
	if err != nil {
		return nil, fmt.Errorf("获取表数据失败: %w", err)
	}
	defer rows.Close()

	// 获取列名
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("获取列名失败: %w", err)
	}

	// 遍历数据行并导入MongoDB
	var documents []interface{}
	for rows.Next() {
		// 准备扫描目标
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		// 扫描一行数据
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("扫描数据失败: %w", err)
		}

		// 构建文档
		doc := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			if val == nil {
				doc[col] = nil
				continue
			}

			// 根据类型进行转换
			switch v := val.(type) {
			case []byte:
				doc[col] = string(v)
			default:
				doc[col] = v
			}
		}
		documents = append(documents, doc)
	}

	// 批量插入文档
	if len(documents) > 0 {
		if _, err := coll.InsertMany(nil, documents); err != nil {
			return nil, fmt.Errorf("导入数据失败: %w", err)
		}
	}

	// 提取MongoDB连接信息
	connInfo, err := connector.ExtractConnectionInfo(dbName)
	if err != nil {
		return nil, fmt.Errorf("获取连接信息失败: %w", err)
	}

	return connInfo, nil
}

// GenerateUnifiedModel 生成统一数据模型
func (s *SQLiteStorage) GenerateUnifiedModel(tableName string) (*datasource.UnifiedDataModel, error) {
	// 获取表结构
	rows, err := s.db.Query(fmt.Sprintf("PRAGMA table_info('%s')", tableName))
	if err != nil {
		return nil, fmt.Errorf("获取表结构失败: %w", err)
	}
	defer rows.Close()

	// 解析表结构
	var columns []datasource.Column
	for rows.Next() {
		var cid int
		var name, dataType string
		var notNull, pk int
		var dfltValue interface{}
		if err := rows.Scan(&cid, &name, &dataType, &notNull, &dfltValue, &pk); err != nil {
			return nil, fmt.Errorf("读取表结构失败: %w", err)
		}

		colType := getColumnType(dataType)
		columns = append(columns, datasource.Column{
			Name:        name,
			DisplayName: name,
			Type:        colType,
		})
	}

	// 获取数据行
	dataRows, err := s.db.Query(fmt.Sprintf("SELECT * FROM '%s'", tableName))
	if err != nil {
		return nil, fmt.Errorf("获取表数据失败: %w", err)
	}
	defer dataRows.Close()

	// 获取列名
	columnNames, err := dataRows.Columns()
	if err != nil {
		return nil, fmt.Errorf("获取列名失败: %w", err)
	}

	// 读取所有数据
	var records []map[string]interface{}
	for dataRows.Next() {
		// 准备扫描目标
		values := make([]interface{}, len(columnNames))
		valuePtrs := make([]interface{}, len(columnNames))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		// 扫描一行数据
		if err := dataRows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("扫描数据失败: %w", err)
		}

		// 构建记录
		record := make(map[string]interface{})
		for i, col := range columnNames {
			val := values[i]
			if val == nil {
				record[col] = nil
				continue
			}

			// 根据类型进行转换
			switch v := val.(type) {
			case []byte:
				record[col] = string(v)
			default:
				record[col] = v
			}
		}
		records = append(records, record)
	}

	// 创建统一数据模型
	model := &datasource.UnifiedDataModel{
		Metadata: datasource.DataMetadata{
			SourceType:  "sqlite",
			SourcePath:  s.filePath,
			ColumnCount: len(columns),
		},
		Columns:      columns,
		Records:      records,
		TotalRecords: len(records),
		Errors:       nil,
	}

	return model, nil
}

// 将SQLite类型映射到统一列类型
func getColumnType(sqliteType string) datasource.ColumnType {
	sqliteType = strings.ToLower(sqliteType)

	switch sqliteType {
	case "integer", "int":
		return datasource.ColumnTypeInteger
	case "real", "float", "double", "numeric", "decimal":
		return datasource.ColumnTypeFloat
	case "boolean":
		return datasource.ColumnTypeBoolean
	case "date", "datetime", "timestamp":
		return datasource.ColumnTypeDate
	default:
		return datasource.ColumnTypeString
	}
}
