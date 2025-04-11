package csv

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"minds_iolite_backend/internal/models/datasource"
)

// CSVParser CSV文件解析器
type CSVParser struct {
	source *datasource.CSVSource
}

// CSVData 解析后的CSV数据
type CSVData struct {
	Headers     []string                         // 列头
	Rows        [][]string                       // 数据行
	LineCount   int                              // 总行数
	ColumnTypes map[string]datasource.ColumnType // 推断的列类型
}

// NewCSVParser 创建一个新的CSV解析器
func NewCSVParser(source *datasource.CSVSource) *CSVParser {
	return &CSVParser{
		source: source,
	}
}

// Parse 解析CSV文件，返回解析结果
func (p *CSVParser) Parse() (*CSVData, error) {
	// 首先验证数据源配置
	if err := p.source.Validate(); err != nil {
		return nil, fmt.Errorf("数据源配置无效: %w", err)
	}

	// 检查文件路径安全性
	if err := validateFilePath(p.source.FilePath); err != nil {
		return nil, fmt.Errorf("文件路径不安全: %w", err)
	}

	// 打开文件
	file, err := os.Open(p.source.FilePath)
	if err != nil {
		return nil, fmt.Errorf("无法打开文件: %w", err)
	}
	defer file.Close()

	// 创建CSV读取器
	reader := csv.NewReader(file)
	reader.Comma = p.source.GetDelimiterRune()
	reader.LazyQuotes = true // 允许宽松的引号处理
	reader.TrimLeadingSpace = true

	// 跳过指定的行数
	for i := 0; i < p.source.SkipRows; i++ {
		_, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				return nil, fmt.Errorf("文件行数少于需要跳过的行数")
			}
			return nil, fmt.Errorf("跳过行时出错: %w", err)
		}
	}

	// 读取标题行
	var headers []string
	if p.source.HasHeader {
		headers, err = reader.Read()
		if err != nil {
			return nil, fmt.Errorf("读取标题行失败: %w", err)
		}
		// 规范化标题
		for i, header := range headers {
			headers[i] = normalizeHeader(header)
		}
	}

	// 读取所有数据行
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("读取数据行失败: %w", err)
	}

	// 如果没有标题行，生成默认标题
	if !p.source.HasHeader {
		if len(rows) > 0 {
			headers = make([]string, len(rows[0]))
			for i := range headers {
				headers[i] = fmt.Sprintf("Column%d", i+1)
			}
		}
	}

	// 创建结果
	result := &CSVData{
		Headers:     headers,
		Rows:        rows,
		LineCount:   len(rows) + p.source.SkipRows + (map[bool]int{true: 1, false: 0})[p.source.HasHeader],
		ColumnTypes: make(map[string]datasource.ColumnType),
	}

	// 推断列类型
	if err := p.inferColumnTypes(result); err != nil {
		return nil, fmt.Errorf("推断列类型失败: %w", err)
	}

	return result, nil
}

// ParseStream 流式解析大文件
func (p *CSVParser) ParseStream(callback func(rowIndex int, row []string) error) error {
	// 验证数据源配置
	if err := p.source.Validate(); err != nil {
		return fmt.Errorf("数据源配置无效: %w", err)
	}

	// 检查文件路径安全性
	if err := validateFilePath(p.source.FilePath); err != nil {
		return fmt.Errorf("文件路径不安全: %w", err)
	}

	// 打开文件
	file, err := os.Open(p.source.FilePath)
	if err != nil {
		return fmt.Errorf("无法打开文件: %w", err)
	}
	defer file.Close()

	// 创建CSV读取器
	reader := csv.NewReader(file)
	reader.Comma = p.source.GetDelimiterRune()
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true

	// 跳过指定的行数
	for i := 0; i < p.source.SkipRows; i++ {
		_, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				return fmt.Errorf("文件行数少于需要跳过的行数")
			}
			return fmt.Errorf("跳过行时出错: %w", err)
		}
	}

	// 读取标题行
	if p.source.HasHeader {
		_, err = reader.Read()
		if err != nil {
			return fmt.Errorf("读取标题行失败: %w", err)
		}
	}

	// 逐行读取并处理
	rowIndex := 0
	for {
		row, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("读取数据行失败: %w", err)
		}

		// 调用回调函数处理行
		if err := callback(rowIndex, row); err != nil {
			return fmt.Errorf("处理行 %d 失败: %w", rowIndex, err)
		}
		rowIndex++
	}

	return nil
}

// DetectColumnTypes 推断列数据类型
func (p *CSVParser) DetectColumnTypes(sampleSize int) (map[string]datasource.ColumnType, error) {
	data, err := p.Parse()
	if err != nil {
		return nil, err
	}

	// 如果没有数据，返回空结果
	if len(data.Rows) == 0 || len(data.Headers) == 0 {
		return make(map[string]datasource.ColumnType), nil
	}

	// 限制样本大小
	sampleRows := data.Rows
	if sampleSize > 0 && len(sampleRows) > sampleSize {
		sampleRows = sampleRows[:sampleSize]
	}

	return p.inferColumnTypesFromSample(data.Headers, sampleRows), nil
}

// inferColumnTypes 从数据推断列类型
func (p *CSVParser) inferColumnTypes(data *CSVData) error {
	if len(data.Rows) == 0 {
		return nil
	}

	data.ColumnTypes = p.inferColumnTypesFromSample(data.Headers, data.Rows)
	return nil
}

// inferColumnTypesFromSample 从样本数据推断列类型
func (p *CSVParser) inferColumnTypesFromSample(headers []string, rows [][]string) map[string]datasource.ColumnType {
	columnTypes := make(map[string]datasource.ColumnType)

	// 初始化所有列为字符串类型
	for _, header := range headers {
		columnTypes[header] = datasource.ColumnTypeString
	}

	// 对每一列进行类型推断
	for colIndex, header := range headers {
		// 统计不同类型的出现次数
		typeCount := map[datasource.ColumnType]int{
			datasource.ColumnTypeInteger: 0,
			datasource.ColumnTypeFloat:   0,
			datasource.ColumnTypeBoolean: 0,
			datasource.ColumnTypeDate:    0,
			datasource.ColumnTypeString:  0,
		}

		// 检查每一行的值
		for _, row := range rows {
			if colIndex >= len(row) {
				continue
			}
			value := strings.TrimSpace(row[colIndex])

			// 忽略空值
			if value == "" {
				continue
			}

			// 尝试各种类型转换
			if isBoolean(value) {
				typeCount[datasource.ColumnTypeBoolean]++
			} else if isInteger(value) {
				typeCount[datasource.ColumnTypeInteger]++
			} else if isFloat(value) {
				typeCount[datasource.ColumnTypeFloat]++
			} else if isDate(value) {
				typeCount[datasource.ColumnTypeDate]++
			} else {
				typeCount[datasource.ColumnTypeString]++
			}
		}

		// 确定最可能的类型
		var mostLikelyType datasource.ColumnType = datasource.ColumnTypeString

		// 按优先级顺序检查
		booleanRatio := float64(typeCount[datasource.ColumnTypeBoolean]) / float64(len(rows))
		integerRatio := float64(typeCount[datasource.ColumnTypeInteger]) / float64(len(rows))
		floatRatio := float64(typeCount[datasource.ColumnTypeFloat]) / float64(len(rows))
		dateRatio := float64(typeCount[datasource.ColumnTypeDate]) / float64(len(rows))

		// 1. 如果布尔值占比高，判断为布尔型
		if booleanRatio > 0.8 {
			mostLikelyType = datasource.ColumnTypeBoolean
		} else if integerRatio > 0.8 {
			// 2. 如果整数占比高，判断为整数型
			mostLikelyType = datasource.ColumnTypeInteger
		} else if floatRatio > 0.8 {
			// 3. 如果浮点数占比高，判断为浮点型
			mostLikelyType = datasource.ColumnTypeFloat
		} else if dateRatio > 0.8 {
			// 4. 如果日期占比高，判断为日期型
			mostLikelyType = datasource.ColumnTypeDate
		} else {
			// 5. 其他情况，判断为字符串型
			mostLikelyType = datasource.ColumnTypeString
		}

		columnTypes[header] = mostLikelyType
	}

	return columnTypes
}

// validateFilePath 验证文件路径是否安全
func validateFilePath(path string) error {
	// 获取绝对路径
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("无法获取绝对路径: %w", err)
	}

	// 检查文件是否存在且可读
	info, err := os.Stat(absPath)
	if err != nil {
		return fmt.Errorf("无法访问文件: %w", err)
	}

	// 确保是文件而不是目录
	if info.IsDir() {
		return fmt.Errorf("路径指向的是目录，而不是文件")
	}

	// TODO: 可以添加其他安全检查，例如权限检查、路径黑名单等

	return nil
}

// normalizeHeader 规范化列标题
func normalizeHeader(header string) string {
	// 去除前后空白
	header = strings.TrimSpace(header)

	// 如果为空，使用默认值
	if header == "" {
		return "Untitled"
	}

	// 替换特殊字符为下划线
	for _, c := range []string{" ", "-", ".", "/", "\\", ":", ";"} {
		header = strings.ReplaceAll(header, c, "_")
	}

	// 确保以字母开头
	if len(header) > 0 && !((header[0] >= 'a' && header[0] <= 'z') || (header[0] >= 'A' && header[0] <= 'Z')) {
		header = "col_" + header
	}

	return header
}

// 类型判断辅助函数
func isInteger(s string) bool {
	_, err := strconv.ParseInt(s, 10, 64)
	return err == nil
}

func isFloat(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

func isBoolean(s string) bool {
	s = strings.ToLower(s)
	return s == "true" || s == "false" || s == "yes" || s == "no" || s == "1" || s == "0"
}

func isDate(s string) bool {
	formats := []string{
		"2006-01-02",
		"2006/01/02",
		"02-01-2006",
		"02/01/2006",
		"2006-01-02 15:04:05",
		"2006/01/02 15:04:05",
		time.RFC3339,
	}

	for _, format := range formats {
		_, err := time.Parse(format, s)
		if err == nil {
			return true
		}
	}

	return false
}
