package main

import (
	"encoding/json"
	"fmt"
	"log"
	"minds_iolite_backend/internal/datasource/providers/csv"
	"minds_iolite_backend/internal/models/datasource"
	"os"
)

func main() {
	fmt.Println("=== 开始CSV测试 ===")
	testCSV()
	fmt.Println("=== CSV测试完成 ===")

	fmt.Println("\n=== 开始CSV导入MongoDB测试 ===")
	testImportCSVToMongo()
	fmt.Println("=== CSV导入MongoDB测试完成 ===")

	// TODO: 后续可添加其他数据源测试
	// testMongoDB()
	// testMySQL()
}

// testCSV 测试CSV数据源功能
func testCSV() {
	// 测试CSV文件路径 - 请替换为实际文件路径
	csvFilePath := "test_data/sample.csv"

	// 创建测试CSV文件
	createTestCSVFile(csvFilePath)

	// 创建CSV数据源
	csvSource := datasource.NewCSVSource(csvFilePath)

	// 验证数据源
	if err := csvSource.Validate(); err != nil {
		log.Fatalf("数据源验证失败: %v", err)
	}

	// 创建解析器
	parser := csv.NewCSVParser(csvSource)

	// 解析CSV文件
	csvData, err := parser.Parse()
	if err != nil {
		log.Fatalf("解析CSV文件失败: %v", err)
	}

	fmt.Printf("解析CSV成功，共 %d 列, %d 行数据\n", len(csvData.Headers), len(csvData.Rows))
	fmt.Printf("列标题: %v\n", csvData.Headers)

	// 推断列类型
	fmt.Println("\n列类型:")
	for header, columnType := range csvData.ColumnTypes {
		fmt.Printf("  %s: %s\n", header, columnType)
	}

	// 创建转换器
	converter := csv.NewCSVConverter(nil, nil)

	// 转换为统一数据模型
	model, err := converter.ConvertToUnifiedModel(csvSource, csvData)
	if err != nil {
		log.Fatalf("转换数据失败: %v", err)
	}

	// 输出数据模型
	fmt.Printf("\n统一数据模型:\n")
	fmt.Printf("  数据源类型: %s\n", model.Metadata.SourceType)
	fmt.Printf("  数据源路径: %s\n", model.Metadata.SourcePath)
	fmt.Printf("  列数: %d\n", model.Metadata.ColumnCount)
	fmt.Printf("  总行数: %d\n", model.TotalRecords)

	// 输出列定义
	fmt.Printf("\n列定义:\n")
	for _, column := range model.Columns {
		fmt.Printf("  %s (%s): %s\n", column.Name, column.DisplayName, column.Type)
	}

	// 输出数据样本
	sampleCount := 5
	if len(model.Records) < sampleCount {
		sampleCount = len(model.Records)
	}
	fmt.Printf("\n数据样本 (%d 行):\n", sampleCount)
	for i := 0; i < sampleCount; i++ {
		recordJSON, _ := json.MarshalIndent(model.Records[i], "  ", "  ")
		fmt.Printf("  记录 %d: %s\n", i+1, string(recordJSON))
	}

	// 输出验证错误
	if len(model.Errors) > 0 {
		fmt.Printf("\n验证错误:\n")
		for _, err := range model.Errors {
			fmt.Printf("  行 %d, 列 %s: %s\n", err.Row, err.Column, err.Message)
		}
	}

	// 测试完成后删除测试文件
	os.Remove(csvFilePath)
	fmt.Println("\n测试完成，已删除测试文件")
}

// 创建测试CSV文件
func createTestCSVFile(filePath string) {
	// 确保目录存在
	dir := "test_data"
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.Mkdir(dir, 0755)
	}

	// 创建文件内容
	content := `id,name,age,email,is_active,score,birth_date
1,张三,30,zhangsan@example.com,true,85.5,1992-05-15
2,李四,25,lisi@example.com,false,92.8,1997-11-23
3,王五,40,wangwu@example.com,true,78.3,1982-03-10
4,赵六,35,zhaoliu@example.com,true,65.7,1987-09-28
5,钱七,28,qianqi@example.com,false,90.1,1994-07-12
6,孙八,33,sunba@example.com,true,81.9,1989-12-05
7,周九,45,zhoujiu@example.com,false,76.4,1977-08-17
8,吴十,22,wushi@example.com,true,95.2,2000-02-20
9,郑十一,38,zheng11@example.com,true,88.7,1984-04-30
10,王十二,31,wang12@example.com,false,72.6,1991-10-25`

	// 写入文件
	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		log.Fatalf("创建测试文件失败: %v", err)
	}

	fmt.Printf("测试文件已创建: %s\n", filePath)
}
