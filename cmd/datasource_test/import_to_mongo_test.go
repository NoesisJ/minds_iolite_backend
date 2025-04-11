package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"minds_iolite_backend/internal/datasource/providers/csv"
	"minds_iolite_backend/internal/models/datasource"
	"minds_iolite_backend/internal/services/datastorage"
)

// testImportCSVToMongo 测试将CSV导入MongoDB的功能
func testImportCSVToMongo() {
	// 测试CSV文件路径
	csvFilePath := "test_data/employees.csv"

	// 创建测试CSV文件
	createEmployeeCSVFile(csvFilePath)

	fmt.Println("=== 开始测试CSV导入MongoDB ===")

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

	// 创建转换器
	converter := csv.NewCSVConverter(nil, nil)

	// 转换为统一数据模型
	model, err := converter.ConvertToUnifiedModel(csvSource, csvData)
	if err != nil {
		log.Fatalf("转换数据失败: %v", err)
	}

	// 创建MongoDB存储服务
	storage, err := datastorage.NewMongoStorage("mongodb://localhost:27017")
	if err != nil {
		log.Fatalf("连接MongoDB失败: %v", err)
	}
	defer storage.Close()

	// 导入数据到MongoDB
	// 使用CSV文件名作为数据库名
	fileName := filepath.Base(csvFilePath)
	dbName := "csv_" + fileName[:len(fileName)-4] // 去掉.csv后缀

	connInfo, err := storage.ImportCSVToMongoDB(model, dbName, "employees")
	if err != nil {
		log.Fatalf("导入数据到MongoDB失败: %v", err)
	}

	// 输出连接信息
	fmt.Println("\nMongoDB连接信息:")
	fmt.Printf("Host: %s\n", connInfo.Host)
	fmt.Printf("Port: %d\n", connInfo.Port)
	fmt.Printf("Database: %s\n", connInfo.Database)
	fmt.Println("Collections:")

	for collName, collInfo := range connInfo.Collections {
		fmt.Printf("  %s:\n", collName)
		fmt.Println("    Fields:")
		for field, fieldType := range collInfo.Fields {
			fmt.Printf("      %s: %s\n", field, fieldType)
		}
		fmt.Printf("    Sample: %s\n", collInfo.SampleData)
	}

	fmt.Println("\n=== CSV导入MongoDB测试完成 ===")
}

// createEmployeeCSVFile 创建员工测试CSV文件
func createEmployeeCSVFile(filePath string) {
	// 确保目录存在
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Fatalf("创建目录失败: %v", err)
	}

	// 创建文件内容
	content := `id,姓名,部门,职位,入职日期,薪资,是否在职
1,张三,技术部,软件工程师,2021-05-12,15000,true
2,李四,人力资源,招聘经理,2020-08-23,12000,true
3,王五,市场部,市场专员,2022-01-15,10000,true
4,赵六,财务部,会计,2019-11-07,11000,true
5,钱七,销售部,销售经理,2021-03-20,18000,true
6,孙八,技术部,前端开发,2022-06-01,14000,true
7,周九,技术部,后端开发,2020-12-15,16000,true
8,吴十,销售部,销售代表,2022-02-28,9000,false
9,郑十一,人力资源,培训师,2021-07-19,11500,true
10,王十二,市场部,市场经理,2019-05-30,17000,true`

	// 写入文件
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		log.Fatalf("创建测试文件失败: %v", err)
	}

	fmt.Printf("员工测试CSV文件已创建: %s\n", filePath)
}
