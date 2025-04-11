package main

import (
	"fmt"
	"minds_iolite_backend/cmd/test"
)

func main() {
	fmt.Println("=== 开始CSV测试 ===")
	test.TestCSV()
	fmt.Println("=== CSV测试完成 ===")

	// TODO: 后续可添加其他数据源测试
	// test.TestMongoDB()
	// test.TestMySQL()
}
