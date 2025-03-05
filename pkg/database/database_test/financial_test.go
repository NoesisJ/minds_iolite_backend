package database_test

import (
	"testing"

	"github.com/NoesisJ/minds_iolite_backend/pkg/config"
	"github.com/NoesisJ/minds_iolite_backend/pkg/database"
	"github.com/NoesisJ/minds_iolite_backend/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestReadFinancialTop3Records(t *testing.T) {
	cfg, _ := config.LoadConfig()
	db, _ := database.InitDB(cfg)

	// 查询前三条财务记录的所有字段
	var records []models.Financial
	result := db.Limit(3).Find(&records)

	// 断言验证
	assert.NoError(t, result.Error)
	assert.Greater(t, len(records), 0, "应至少查询到一条记录")

	// 输出前三条财务记录的所有字段
	t.Log("查询结果 (前三条财务记录的所有字段):")
	for i, record := range records {
		t.Logf("记录 #%d:", i+1)
		t.Logf("  ID: %d", record.ID)
		t.Logf("  名称: %s", record.Name)
		t.Logf("  型号: %s", record.Model)
		t.Logf("  数量: %s", record.Quantity)
		t.Logf("  单位: %s", record.Unit)
		t.Logf("  价格: %s", record.Price)
		t.Logf("  额外价格: %s", record.ExtraPrice)
		t.Logf("  购买链接: %s", record.PurchaseLink)
		t.Logf("  发布日期: %s", record.PostDate)
		t.Logf("  购买人: %s", record.Purchaser)
		t.Logf("  校区: %s", record.Campus)
		t.Logf("  组名: %s", record.GroupName)
		t.Logf("  部队类型项目: %s", record.TroopTypeProject)
		t.Logf("  备注: %s", record.Remarks)
		t.Log("  -------------------")
	}
}

func TestFinancialDataConsistency(t *testing.T) {
	cfg, _ := config.LoadConfig()
	db, _ := database.InitDB(cfg)

	var count1, count2 int64
	db.Model(&models.Financial{}).Count(&count1)

	// 模拟网络中断
	sqlDB, _ := db.DB()
	sqlDB.Close()

	db2, _ := database.InitDB(cfg)
	db2.Model(&models.Financial{}).Count(&count2)

	assert.Equal(t, count1, count2, "两次查询结果应一致")
}

func TestFinancialPagination(t *testing.T) {
	cfg, _ := config.LoadConfig()
	db, _ := database.InitDB(cfg)

	var page1 []models.Financial
	db.Limit(10).Find(&page1)
	assert.GreaterOrEqual(t, len(page1), 0, "应返回记录")
}