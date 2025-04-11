package datastorage

import (
	"database/sql"
	"encoding/json"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

// MySQLStorage 提供MySQL存储功能
type MySQLStorage struct {
	db       *sql.DB
	host     string
	port     int
	username string
	database string
}

// MySQLConnectionInfo 表示MySQL连接信息
type MySQLConnectionInfo struct {
	Host     string                      `json:"host"`
	Port     int                         `json:"port"`
	Username string                      `json:"username"`
	Password string                      `json:"password,omitempty"`
	Database string                      `json:"database"`
	Tables   map[string]TableInformation `json:"tables"`
}

// TableInformation 表示表信息
type TableInformation struct {
	Fields     map[string]string `json:"fields"`
	SampleData string            `json:"sample_data"`
}

// NewMySQLStorage 创建新的MySQL存储服务
func NewMySQLStorage(host string, port int, username, password, database string) (*MySQLStorage, error) {
	// 构建DSN (Data Source Name)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=true&loc=Local",
		username, password, host, port, database)

	// 连接数据库
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("连接MySQL失败: %w", err)
	}

	// 设置连接池参数
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	// 测试连接
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("MySQL服务器无响应: %w", err)
	}

	return &MySQLStorage{
		db:       db,
		host:     host,
		port:     port,
		username: username,
		database: database,
	}, nil
}

// Close 关闭MySQL连接
func (s *MySQLStorage) Close() error {
	return s.db.Close()
}

// GenerateConnectionInfo 生成连接信息
func (s *MySQLStorage) GenerateConnectionInfo() (*MySQLConnectionInfo, error) {
	// 创建连接信息
	connInfo := &MySQLConnectionInfo{
		Host:     s.host,
		Port:     s.port,
		Username: s.username,
		Password: "",
		Database: s.database,
		Tables:   make(map[string]TableInformation),
	}

	// 获取表列表
	rows, err := s.db.Query("SHOW TABLES")
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
		return nil, fmt.Errorf("数据库 %s 中没有表", s.database)
	}

	// 处理每个表
	for _, tableName := range tables {
		// 获取表结构
		columnsRows, err := s.db.Query(fmt.Sprintf("DESCRIBE %s", tableName))
		if err != nil {
			return nil, fmt.Errorf("获取表 %s 结构失败: %w", tableName, err)
		}

		fields := make(map[string]string)
		for columnsRows.Next() {
			var field, fieldType, null, key, extra string
			var defaultValue sql.NullString
			if err := columnsRows.Scan(&field, &fieldType, &null, &key, &defaultValue, &extra); err != nil {
				columnsRows.Close()
				return nil, fmt.Errorf("读取表结构失败: %w", err)
			}
			fields[field] = s.getMySQLType(fieldType)
		}
		columnsRows.Close()

		// 添加表信息 - 暂不获取样本数据以简化实现
		connInfo.Tables[tableName] = TableInformation{
			Fields:     fields,
			SampleData: "{}",
		}
	}

	return connInfo, nil
}

// GenerateConnectionInfoForTable 生成指定表的连接信息
func (s *MySQLStorage) GenerateConnectionInfoForTable(tableName string) (*MySQLConnectionInfo, error) {
	// 创建连接信息
	connInfo := &MySQLConnectionInfo{
		Host:     s.host,
		Port:     s.port,
		Username: s.username,
		Password: "",
		Database: s.database,
		Tables:   make(map[string]TableInformation),
	}

	// 获取表结构
	columnsRows, err := s.db.Query(fmt.Sprintf("DESCRIBE %s", tableName))
	if err != nil {
		return nil, fmt.Errorf("获取表 %s 结构失败: %w", tableName, err)
	}
	defer columnsRows.Close()

	fields := make(map[string]string)
	for columnsRows.Next() {
		var field, fieldType, null, key, extra string
		var defaultValue sql.NullString
		if err := columnsRows.Scan(&field, &fieldType, &null, &key, &defaultValue, &extra); err != nil {
			return nil, fmt.Errorf("读取表结构失败: %w", err)
		}
		fields[field] = s.getMySQLType(fieldType)
	}

	// 获取样本数据
	sampleRow, err := s.db.Query(fmt.Sprintf("SELECT * FROM %s LIMIT 1", tableName))
	if err != nil {
		return nil, fmt.Errorf("获取表 %s 的样本数据失败: %w", tableName, err)
	}
	defer sampleRow.Close()

	var sampleData string
	if sampleRow.Next() {
		// 获取列名
		columns, err := sampleRow.Columns()
		if err != nil {
			return nil, fmt.Errorf("获取列名失败: %w", err)
		}

		// 准备扫描目标
		values := make([]interface{}, len(columns))
		scanArgs := make([]interface{}, len(columns))
		for i := range values {
			scanArgs[i] = &values[i]
		}

		// 扫描一行数据
		if err := sampleRow.Scan(scanArgs...); err != nil {
			return nil, fmt.Errorf("扫描数据失败: %w", err)
		}

		// 构建样本数据
		sampleMap := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			if val == nil {
				sampleMap[col] = nil
				continue
			}

			// 根据类型进行转换
			switch v := val.(type) {
			case []byte:
				sampleMap[col] = string(v)
			default:
				sampleMap[col] = v
			}
		}

		// 转换为JSON字符串
		sampleBytes, err := json.Marshal(sampleMap)
		if err != nil {
			return nil, fmt.Errorf("转换样本数据失败: %w", err)
		}
		sampleData = string(sampleBytes)
	} else {
		sampleData = "{}"
	}

	// 添加表信息
	connInfo.Tables[tableName] = TableInformation{
		Fields:     fields,
		SampleData: sampleData,
	}

	return connInfo, nil
}

// getMySQLType 根据MySQL类型返回简化类型名称
func (s *MySQLStorage) getMySQLType(mysqlType string) string {
	if len(mysqlType) == 0 {
		return "unknown"
	}

	// 提取MySQL类型前缀
	typePrefix := mysqlType
	// 去除括号和参数部分，如int(11)只保留int
	for i, char := range typePrefix {
		if char == '(' {
			typePrefix = typePrefix[:i]
			break
		}
	}

	// 按类型返回统一的类型名称
	switch typePrefix {
	case "int", "smallint", "mediumint", "bigint":
		return "int"
	case "tinyint":
		if mysqlType == "tinyint(1)" {
			return "bool"
		}
		return "int"
	case "varchar", "char", "text", "longtext":
		return "str"
	case "datetime", "date", "timestamp":
		return "date"
	case "decimal", "float", "double":
		return "float"
	case "blob", "longblob":
		return "binary"
	default:
		return "unknown"
	}
}
