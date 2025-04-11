package csv

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"minds_iolite_backend/internal/models/datasource"
)

// CSVConverter CSV数据转换器
// 负责将CSV数据转换为统一数据模型
type CSVConverter struct {
	ColumnMapping map[string]string                // 列名到目标字段的映射
	TypeMapping   map[string]datasource.ColumnType // 列名到数据类型的映射
}

// NewCSVConverter 创建新的CSV转换器
func NewCSVConverter(columnMapping map[string]string, typeMapping map[string]datasource.ColumnType) *CSVConverter {
	if columnMapping == nil {
		columnMapping = make(map[string]string)
	}
	if typeMapping == nil {
		typeMapping = make(map[string]datasource.ColumnType)
	}
	return &CSVConverter{
		ColumnMapping: columnMapping,
		TypeMapping:   typeMapping,
	}
}

// ConvertToUnifiedModel 将CSV数据转换为统一数据模型
func (c *CSVConverter) ConvertToUnifiedModel(csvSource *datasource.CSVSource, csvData *CSVData) (*datasource.UnifiedDataModel, error) {
	if csvData == nil || len(csvData.Headers) == 0 {
		return nil, fmt.Errorf("无效的CSV数据")
	}

	// 创建统一数据模型
	model := datasource.NewUnifiedDataModel("csv", csvSource.FilePath)
	model.Metadata.HasHeader = csvSource.HasHeader
	model.Metadata.RowCount = csvData.LineCount
	model.Metadata.ColumnCount = len(csvData.Headers)
	model.TotalRecords = len(csvData.Rows)

	// 创建列定义
	for i, header := range csvData.Headers {
		columnType := datasource.ColumnTypeString

		// 使用推断的类型或指定的类型
		if inferredType, ok := csvData.ColumnTypes[header]; ok {
			columnType = inferredType
		}
		if mappedType, ok := c.TypeMapping[header]; ok {
			columnType = mappedType
		}

		// 创建列定义
		column := datasource.Column{
			Name:        header,
			DisplayName: c.getDisplayName(header),
			Type:        columnType,
			Required:    false,
			Description: fmt.Sprintf("CSV列 #%d", i+1),
		}
		model.Columns = append(model.Columns, column)
	}

	// 转换数据记录
	for rowIndex, row := range csvData.Rows {
		record := make(map[string]interface{})

		// 处理每一列
		for colIndex, value := range row {
			if colIndex >= len(csvData.Headers) {
				continue // 跳过超出列头数量的数据
			}

			header := csvData.Headers[colIndex]
			// 应用列映射
			targetField := header
			if mapped, ok := c.ColumnMapping[header]; ok && mapped != "" {
				targetField = mapped
			}

			// 根据列类型转换值
			columnType := csvData.ColumnTypes[header]
			if mappedType, ok := c.TypeMapping[header]; ok {
				columnType = mappedType
			}

			convertedValue, err := c.convertValue(value, columnType)
			if err != nil {
				// 添加转换错误到模型错误列表
				model.Errors = append(model.Errors, datasource.ValidationError{
					Row:     rowIndex + 1,
					Column:  header,
					Message: fmt.Sprintf("值转换失败: %v", err),
				})
				// 使用原始字符串值
				record[targetField] = value
			} else {
				record[targetField] = convertedValue
			}
		}

		model.Records = append(model.Records, record)
	}

	// 设置预览计数
	model.Metadata.PreviewCount = len(model.Records)

	return model, nil
}

// ValidateData 验证CSV数据
func (c *CSVConverter) ValidateData(csvData *CSVData) []datasource.ValidationError {
	var errors []datasource.ValidationError

	// 检查数据是否为空
	if csvData == nil || len(csvData.Headers) == 0 {
		errors = append(errors, datasource.ValidationError{
			Row:     0,
			Column:  "",
			Message: "CSV数据为空或无效",
		})
		return errors
	}

	// 检查每行数据
	for rowIndex, row := range csvData.Rows {
		// 检查列数是否匹配
		if len(row) != len(csvData.Headers) {
			errors = append(errors, datasource.ValidationError{
				Row:     rowIndex + 1,
				Column:  "",
				Message: fmt.Sprintf("列数不匹配: 期望 %d 列, 实际 %d 列", len(csvData.Headers), len(row)),
			})
		}

		// 检查每列数据
		for colIndex, value := range row {
			if colIndex >= len(csvData.Headers) {
				break
			}

			header := csvData.Headers[colIndex]
			columnType := csvData.ColumnTypes[header]

			// 根据列类型验证值
			if err := c.validateValue(value, columnType); err != nil {
				errors = append(errors, datasource.ValidationError{
					Row:     rowIndex + 1,
					Column:  header,
					Message: err.Error(),
				})
			}
		}
	}

	return errors
}

// convertValue 根据类型转换值
func (c *CSVConverter) convertValue(value string, columnType datasource.ColumnType) (interface{}, error) {
	value = strings.TrimSpace(value)

	// 处理空值
	if value == "" {
		return nil, nil
	}

	switch columnType {
	case datasource.ColumnTypeInteger:
		return strconv.ParseInt(value, 10, 64)
	case datasource.ColumnTypeFloat:
		return strconv.ParseFloat(value, 64)
	case datasource.ColumnTypeBoolean:
		return parseBool(value)
	case datasource.ColumnTypeDate:
		return parseDate(value)
	case datasource.ColumnTypeObject:
		// 简单实现，可以扩展为JSON解析
		return map[string]interface{}{"value": value}, nil
	default:
		return value, nil
	}
}

// validateValue 验证值是否符合类型要求
func (c *CSVConverter) validateValue(value string, columnType datasource.ColumnType) error {
	value = strings.TrimSpace(value)

	// 空值直接通过
	if value == "" {
		return nil
	}

	switch columnType {
	case datasource.ColumnTypeInteger:
		_, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fmt.Errorf("值 '%s' 不是有效的整数", value)
		}
	case datasource.ColumnTypeFloat:
		_, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("值 '%s' 不是有效的浮点数", value)
		}
	case datasource.ColumnTypeBoolean:
		_, err := parseBool(value)
		if err != nil {
			return fmt.Errorf("值 '%s' 不是有效的布尔值", value)
		}
	case datasource.ColumnTypeDate:
		_, err := parseDate(value)
		if err != nil {
			return fmt.Errorf("值 '%s' 不是有效的日期", value)
		}
	}

	return nil
}

// getDisplayName 从字段名生成显示名称
func (c *CSVConverter) getDisplayName(fieldName string) string {
	// 将下划线替换为空格
	name := strings.ReplaceAll(fieldName, "_", " ")

	// 首字母大写
	words := strings.Split(name, " ")
	for i, word := range words {
		if word == "" {
			continue
		}
		words[i] = strings.ToUpper(word[:1]) + word[1:]
	}

	return strings.Join(words, " ")
}

// 辅助函数

// parseBool 解析布尔值
func parseBool(value string) (bool, error) {
	value = strings.ToLower(value)
	switch value {
	case "true", "yes", "1", "y", "t":
		return true, nil
	case "false", "no", "0", "n", "f":
		return false, nil
	default:
		return false, fmt.Errorf("无法将 '%s' 解析为布尔值", value)
	}
}

// parseDate 解析日期
func parseDate(value string) (time.Time, error) {
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
		if t, err := time.Parse(format, value); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("无法将 '%s' 解析为日期", value)
}
