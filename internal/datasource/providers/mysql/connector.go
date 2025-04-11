package mysql

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"minds_iolite_backend/internal/services/datastorage"

	_ "github.com/go-sql-driver/mysql"
)

// MySQLConnector MySQL连接器
type MySQLConnector struct {
	db       *sql.DB
	dsn      string
	host     string
	port     int
	username string
	database string
}

// NewMySQLConnector 创建MySQL连接器
func NewMySQLConnector(host string, port int, username, password, database string) (*MySQLConnector, error) {
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
	db.SetConnMaxLifetime(time.Minute * 3)

	// 测试连接
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("MySQL服务器无响应: %w", err)
	}

	return &MySQLConnector{
		db:       db,
		dsn:      dsn,
		host:     host,
		port:     port,
		username: username,
		database: database,
	}, nil
}

// Close 关闭连接
func (c *MySQLConnector) Close() error {
	return c.db.Close()
}

// ExtractConnectionInfo 提取数据库连接信息
func (c *MySQLConnector) ExtractConnectionInfo() (*datastorage.MySQLConnectionInfo, error) {
	// 创建连接信息
	connInfo := &datastorage.MySQLConnectionInfo{
		Host:     c.host,
		Port:     c.port,
		Username: c.username,
		Password: "",
		Database: c.database,
		Tables:   make(map[string]datastorage.TableInformation),
	}

	// 获取表列表
	rows, err := c.db.Query("SHOW TABLES")
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
		return nil, fmt.Errorf("数据库 %s 中没有表", c.database)
	}

	// 处理每个表
	for _, tableName := range tables {
		// 获取表结构
		columnsRows, err := c.db.Query(fmt.Sprintf("DESCRIBE %s", tableName))
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
			fields[field] = getMySQLType(fieldType)
		}
		columnsRows.Close()

		// 获取样本数据
		sampleRows, err := c.db.Query(fmt.Sprintf("SELECT * FROM %s LIMIT 1", tableName))
		if err != nil {
			return nil, fmt.Errorf("获取表 %s 的样本数据失败: %w", tableName, err)
		}
		defer sampleRows.Close()

		// 获取列名
		columns, err := sampleRows.Columns()
		if err != nil {
			return nil, fmt.Errorf("获取列名失败: %w", err)
		}

		if !sampleRows.Next() {
			// 表为空，添加空样本
			connInfo.Tables[tableName] = datastorage.TableInformation{
				Fields:     fields,
				SampleData: "{}",
			}
			continue
		}

		// 准备扫描目标
		values := make([]interface{}, len(columns))
		scanArgs := make([]interface{}, len(columns))
		for i := range values {
			scanArgs[i] = &values[i]
		}

		// 扫描一行数据
		if err := sampleRows.Scan(scanArgs...); err != nil {
			return nil, fmt.Errorf("扫描数据失败: %w", err)
		}

		// 构建样本数据
		sampleData := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			// 处理NULL值
			if val == nil {
				sampleData[col] = nil
				continue
			}

			// 根据类型进行转换
			switch v := val.(type) {
			case []byte:
				// 尝试解析日期或数字
				sampleData[col] = string(v)
			default:
				sampleData[col] = v
			}
		}

		// 转换为JSON
		sampleJSON, err := json.Marshal(sampleData)
		if err != nil {
			return nil, fmt.Errorf("转换样本数据失败: %w", err)
		}

		// 添加表信息
		connInfo.Tables[tableName] = datastorage.TableInformation{
			Fields:     fields,
			SampleData: string(sampleJSON),
		}
	}

	return connInfo, nil
}

// getMySQLType 根据MySQL类型返回简化类型名称
func getMySQLType(mysqlType string) string {
	mysqlType = strings.ToLower(mysqlType)

	if strings.Contains(mysqlType, "int") {
		return "int"
	}
	if strings.Contains(mysqlType, "float") || strings.Contains(mysqlType, "double") || strings.Contains(mysqlType, "decimal") {
		return "float"
	}
	if strings.Contains(mysqlType, "char") || strings.Contains(mysqlType, "text") {
		return "str"
	}
	if strings.Contains(mysqlType, "date") || strings.Contains(mysqlType, "time") {
		return "date"
	}
	if strings.Contains(mysqlType, "blob") || strings.Contains(mysqlType, "binary") {
		return "binary"
	}
	if mysqlType == "tinyint(1)" {
		return "bool"
	}

	return "unknown"
}
