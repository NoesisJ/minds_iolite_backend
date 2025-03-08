package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Username string `gorm:"type:varchar(191);uniqueIndex;not null"`
	Email    string `gorm:"type:varchar(191);uniqueIndex;not null"`
	Password string `gorm:"type:varchar(191);not null"`
}

func AutoMigrate(db *gorm.DB) error {
	// 由于Data表已存在且有id列，我们只迁移User模型
	return db.AutoMigrate(&User{}) // 仅执行结构迁移，不会删除数据
}
