package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Username string `gorm:"type:varchar(191);uniqueIndex;not null"`
	Email    string `gorm:"type:varchar(191);uniqueIndex;not null"`
	Password string `gorm:"type:varchar(191);not null"`
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&User{}, &Data{}) // 仅执行结构迁移，不会删除数据
}
