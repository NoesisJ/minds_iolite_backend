package datasource

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SQLiteSource 定义SQLite数据源配置
type SQLiteSource struct {
	FilePath string `json:"filePath" binding:"required"` // SQLite文件路径
	Table    string `json:"table"`                       // 指定要导入的表名，为空则导入所有表
}

// NewSQLiteSource 创建带默认值的SQLite数据源
func NewSQLiteSource(filePath string) *SQLiteSource {
	return &SQLiteSource{
		FilePath: filePath,
	}
}

// Validate 验证SQLite数据源配置的有效性
func (s *SQLiteSource) Validate() error {
	// 检查文件路径是否为空
	if s.FilePath == "" {
		return fmt.Errorf("SQLite文件路径不能为空")
	}

	// 检查文件是否存在
	_, err := os.Stat(s.FilePath)
	if os.IsNotExist(err) {
		return fmt.Errorf("SQLite文件不存在: %s", s.FilePath)
	}

	// 检查文件扩展名
	ext := strings.ToLower(filepath.Ext(s.FilePath))
	if ext != ".db" && ext != ".sqlite" && ext != ".sqlite3" {
		return fmt.Errorf("不支持的SQLite文件格式: %s, 仅支持.db, .sqlite, .sqlite3", ext)
	}

	return nil
}

// GetFileName 获取文件名（不含扩展名）
func (s *SQLiteSource) GetFileName() string {
	baseName := filepath.Base(s.FilePath)
	return strings.TrimSuffix(baseName, filepath.Ext(baseName))
}
