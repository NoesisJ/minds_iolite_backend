package datasource

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// CSVSource 定义CSV数据源的配置
type CSVSource struct {
	FilePath    string            `json:"filePath"`    // CSV文件路径
	Delimiter   string            `json:"delimiter"`   // 分隔符，默认为逗号
	HasHeader   bool              `json:"hasHeader"`   // 是否有表头
	SkipRows    int               `json:"skipRows"`    // 跳过起始行数
	Encoding    string            `json:"encoding"`    // 文件编码
	ColumnTypes map[string]string `json:"columnTypes"` // 列数据类型映射
}

// NewCSVSource 创建一个新的CSV数据源配置，使用默认值
func NewCSVSource(filePath string) *CSVSource {
	return &CSVSource{
		FilePath:    filePath,
		Delimiter:   ",",
		HasHeader:   true,
		SkipRows:    0,
		Encoding:    "utf-8",
		ColumnTypes: make(map[string]string),
	}
}

// Validate 验证CSV数据源配置的有效性
func (s *CSVSource) Validate() error {
	// 检查文件路径
	if s.FilePath == "" {
		return errors.New("文件路径不能为空")
	}

	// 验证文件是否存在
	_, err := os.Stat(s.FilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("文件不存在: %s", s.FilePath)
		}
		return fmt.Errorf("无法访问文件: %w", err)
	}

	// 验证文件扩展名
	ext := strings.ToLower(filepath.Ext(s.FilePath))
	if ext != ".csv" {
		return fmt.Errorf("不支持的文件类型，期望 .csv，实际为 %s", ext)
	}

	// 验证分隔符
	if len(s.Delimiter) == 0 {
		return errors.New("分隔符不能为空")
	}

	// 验证编码
	supportedEncodings := map[string]bool{
		"utf-8":      true,
		"utf8":       true,
		"gbk":        true,
		"gb18030":    true,
		"iso-8859-1": true,
	}
	s.Encoding = strings.ToLower(s.Encoding)
	if !supportedEncodings[s.Encoding] {
		return fmt.Errorf("不支持的编码: %s", s.Encoding)
	}

	// 验证跳过行数
	if s.SkipRows < 0 {
		return errors.New("跳过行数不能为负数")
	}

	return nil
}

// GetDelimiterRune 返回分隔符的rune表示
func (s *CSVSource) GetDelimiterRune() rune {
	if len(s.Delimiter) == 0 {
		return ','
	}
	return rune(s.Delimiter[0])
}
