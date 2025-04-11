package sqlite

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"minds_iolite_backend/internal/services/datastorage"

	_ "github.com/mattn/go-sqlite3"
)

// SQLiteConnector SQLite连接器
type SQLiteConnector struct {
	db       *sql.DB
	filePath string
}

// NewSQLiteConnector 创建SQLite连接器
func NewSQLiteConnector(filePath string) (*SQLiteConnector, error) {
	// 连接SQLite数据库
	db, err := sql.Open("sqlite3", filePath)
	if err != nil {
		return nil, fmt.Errorf("连接SQLite数据库失败: %w", err)
	}

	// 设置连接参数
	db.SetMaxOpenConns(1) // SQLite建议只用一个连接
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(time.Minute * 3)

	// 测试连接
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("SQLite数据库无响应: %w", err)
	}

	return &SQLiteConnector{
		db:       db,
		filePath: filePath,
	}, nil
}

// Close 关闭连接
func (c *SQLiteConnector) Close() error {
	return c.db.Close()
}

// GetTableNames 获取所有表名
func (c *SQLiteConnector) GetTableNames() ([]string, error) {
	// 查询所有表名
	rows, err := c.db.Query(`
		SELECT name FROM sqlite_master 
		WHERE type='table' AND name NOT LIKE 'sqlite_%'
	`)
	if err != nil {
		return nil, fmt.Errorf("获取表列表失败: %w", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, fmt.Errorf("读取表名失败: %w", err)
		}
		tables = append(tables, tableName)
	}

	// 若没有表，返回错误
	if len(tables) == 0 {
		return nil, fmt.Errorf("SQLite数据库中没有表")
	}

	return tables, nil
}

// ExtractTableInfo 提取表结构信息
func (c *SQLiteConnector) ExtractTableInfo(tableName string) (*datastorage.TableInformation, error) {
	// 获取表结构
	rows, err := c.db.Query(fmt.Sprintf("PRAGMA table_info('%s')", tableName))
	if err != nil {
		return nil, fmt.Errorf("获取表 %s 结构失败: %w", tableName, err)
	}
	defer rows.Close()

	fields := make(map[string]string)
	for rows.Next() {
		var cid int
		var name, dataType string
		var notNull, pk int
		var dfltValue interface{}
		if err := rows.Scan(&cid, &name, &dataType, &notNull, &dfltValue, &pk); err != nil {
			return nil, fmt.Errorf("读取表结构失败: %w", err)
		}
		fields[name] = getSQLiteType(dataType)
	}

	// 获取样本数据
	var sampleData string
	sampleRow, err := c.db.Query(fmt.Sprintf("SELECT * FROM '%s' LIMIT 1", tableName))
	if err != nil {
		return nil, fmt.Errorf("获取表 %s 的样本数据失败: %w", tableName, err)
	}
	defer sampleRow.Close()

	if sampleRow.Next() {
		columns, err := sampleRow.Columns()
		if err != nil {
			return nil, fmt.Errorf("获取列名失败: %w", err)
		}

		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := sampleRow.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("读取样本行失败: %w", err)
		}

		// 构建样本数据
		sample := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			if val == nil {
				sample[col] = nil
				continue
			}

			switch v := val.(type) {
			case []byte:
				sample[col] = string(v)
			default:
				sample[col] = v
			}
		}

		// 转换为JSON字符串
		sampleBytes, err := json.Marshal(sample)
		if err != nil {
			return nil, fmt.Errorf("转换样本数据失败: %w", err)
		}
		sampleData = string(sampleBytes)
	} else {
		sampleData = "{}"
	}

	return &datastorage.TableInformation{
		Fields:     fields,
		SampleData: sampleData,
	}, nil
}

// ExtractConnectionInfo 提取数据库连接信息
func (c *SQLiteConnector) ExtractConnectionInfo() (*datastorage.SQLiteConnectionInfo, error) {
	// 获取表名列表
	tables, err := c.GetTableNames()
	if err != nil {
		return nil, err
	}

	// 创建连接信息
	connInfo := &datastorage.SQLiteConnectionInfo{
		FilePath:  c.filePath,
		TableInfo: make(map[string]datastorage.TableInformation),
	}

	// 获取各表信息
	for _, tableName := range tables {
		tableInfo, err := c.ExtractTableInfo(tableName)
		if err != nil {
			return nil, err
		}
		connInfo.TableInfo[tableName] = *tableInfo
	}

	return connInfo, nil
}

// ExtractTableConnectionInfo 提取指定表的连接信息
func (c *SQLiteConnector) ExtractTableConnectionInfo(tableName string) (*datastorage.SQLiteConnectionInfo, error) {
	// 创建连接信息
	connInfo := &datastorage.SQLiteConnectionInfo{
		FilePath:  c.filePath,
		TableInfo: make(map[string]datastorage.TableInformation),
	}

	// 获取表信息
	tableInfo, err := c.ExtractTableInfo(tableName)
	if err != nil {
		return nil, err
	}
	connInfo.TableInfo[tableName] = *tableInfo

	return connInfo, nil
}

// 根据SQLite类型返回统一类型
func getSQLiteType(sqliteType string) string {
	sqliteType = strings.ToLower(sqliteType)

	switch sqliteType {
	case "integer", "int":
		return "int"
	case "real", "float", "double", "numeric", "decimal":
		return "float"
	case "text", "char", "varchar", "varying character", "character":
		return "str"
	case "boolean":
		return "bool"
	case "date", "datetime", "timestamp":
		return "date"
	case "blob":
		return "binary"
	default:
		return "unknown"
	}
}
