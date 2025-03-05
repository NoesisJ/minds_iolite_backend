package database_test

import (
	"os"
	"testing"

	"github.com/NoesisJ/minds_iolite_backend/pkg/config"
	"github.com/NoesisJ/minds_iolite_backend/pkg/database"
	"github.com/NoesisJ/minds_iolite_backend/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestDatabaseConnection(t *testing.T) {
	// 添加路径调试
	t.Log("Current working directory:", os.Getenv("PWD"))

	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatal("Config error:", err)
	}

	// 初始化数据库
	db, err := database.InitDB(cfg)
	assert.NoError(t, err, "应该能连接数据库")
	assert.NotNil(t, db, "数据库实例不应为nil")

	// 替换自动迁移为结构验证
	if !db.Migrator().HasTable(&models.Data{}) {
		// 仅当表不存在时创建
		err = models.AutoMigrate(db)
		assert.NoError(t, err, "应该能创建表结构")
	} else {
		// 验证现有表结构
		columnTypes, _ := db.Migrator().ColumnTypes(&models.Data{})
		for _, ct := range columnTypes {
			t.Logf("列验证: %s (%s)", ct.Name(), ct.DatabaseTypeName())
		}
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	err = sqlDB.Ping()
	assert.NoError(t, err, "数据库Ping失败")

	t.Logf("使用的表名: %s", db.Model(&models.Data{}).Statement.Table)
}

func TestReadData(t *testing.T) {
	cfg, _ := config.LoadConfig()
	db, _ := database.InitDB(cfg)

	// 查询前两个用户
	var records []models.Data
	result := db.Limit(2).Find(&records)

	// 断言验证
	assert.NoError(t, result.Error)
	assert.Greater(t, len(records), 0, "应至少查询到一条记录")

	// 验证字段映射
	for _, record := range records {
		assert.NotEmpty(t, record.Nickname, "用户昵称不应为空")
		assert.NotEmpty(t, record.Age, "用户年龄不应为空")
	}
}

func TestReadDataTop3Fields(t *testing.T) {
	cfg, _ := config.LoadConfig()
	db, _ := database.InitDB(cfg)

	// 查询前三个用户的所有字段
	var records []models.Data
	result := db.Limit(3).Find(&records)

	// 断言验证
	assert.NoError(t, result.Error)
	assert.Greater(t, len(records), 0, "应至少查询到一条记录")

	// 输出前三个用户的所有字段
	t.Log("查询结果 (前三个用户的所有字段):")
	for i, record := range records {
		t.Logf("记录 #%d:", i+1)
		t.Logf("  ID: %d", record.ID)
		t.Logf("  昵称: %s", record.Nickname)
		t.Logf("  身份证: %s", record.IDCard)
		t.Logf("  性别: %s", record.Sex)
		t.Logf("  年龄: %s", record.Age)
		t.Logf("  地址: %s", record.Address)
		t.Logf("  分类: %s", record.Classification)
		t.Logf("  学校: %s", record.School)
		t.Logf("  科目: %s", record.Subjects)
		t.Logf("  电话: %s", record.Phone)
		t.Logf("  邮箱: %s", record.Email)
		t.Logf("  QQ: %s", record.QQ)
		t.Logf("  微信: %s", record.Wechat)
		t.Logf("  网络ID: %s", record.WebID)
		t.Logf("  吉大组: %s", record.JLUGroup)
		t.Logf("  学习: %s", record.Study)
		t.Logf("  身份: %s", record.Identity)
		t.Logf("  状态: %s", record.State)
		t.Log("  -------------------")
	}
}

func TestMigrationSafety(t *testing.T) {
	cfg, _ := config.LoadConfig()
	db, _ := database.InitDB(cfg)

	// 验证表是否存在
	hasTable := db.Migrator().HasTable(&models.Data{})
	assert.True(t, hasTable, "数据表应存在")

	// 验证记录数
	var count int64
	db.Model(&models.Data{}).Count(&count)
	assert.Greater(t, count, int64(0), "生产环境数据不应为空")
}

func TestDataPagination(t *testing.T) {
	cfg, _ := config.LoadConfig()
	db, _ := database.InitDB(cfg)

	var page1 []models.Data
	db.Limit(10).Find(&page1)
	assert.Len(t, page1, 10, "应返回10条记录")
}

func BenchmarkDataQuery(b *testing.B) {
	cfg, _ := config.LoadConfig()
	db, _ := database.InitDB(cfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var records []models.Data
		db.Find(&records)
	}
}

func TestDataConsistency(t *testing.T) {
	cfg, _ := config.LoadConfig()
	db, _ := database.InitDB(cfg)

	var count1, count2 int64
	db.Model(&models.Data{}).Count(&count1)

	// 模拟网络中断
	sqlDB, _ := db.DB()
	sqlDB.Close()

	db2, _ := database.InitDB(cfg)
	db2.Model(&models.Data{}).Count(&count2)

	assert.Equal(t, count1, count2, "两次查询结果应一致")
}

func TestSchemaConsistency(t *testing.T) {
	cfg, _ := config.LoadConfig()
	db, _ := database.InitDB(cfg)

	// 获取模型期望的列信息
	expectedColumns := map[string]string{
		"ID":             "bigint",
		"nickname":       "varchar",
		"IDcard":         "longtext",
		"sex":            "varchar",
		"age":            "text",
		"address":        "text",
		"classification": "text",
		"school":         "text",
		"subjects":       "text",
		"phone":          "varchar",
		"email":          "varchar",
		"qq":             "varchar",
		"wechat":         "varchar",
		"webID":          "text",
		"jlugroup":       "text",
		"study":          "text",
		"identity":       "text",
		"state":          "text",
		"image1":         "longtext",
		"image2":         "longtext",
	}

	// 获取实际数据库列信息
	columnTypes, _ := db.Migrator().ColumnTypes(&models.Data{})
	for _, ct := range columnTypes {
		colName := ct.Name()
		if expectedType, ok := expectedColumns[colName]; ok {
			assert.Equal(t, expectedType, ct.DatabaseTypeName(),
				"列 %s 类型不匹配", colName)
		} else {
			t.Errorf("发现未预期的列: %s", colName)
		}
	}
}
