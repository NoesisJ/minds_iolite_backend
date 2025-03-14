package models

import (
	"gorm.io/gorm"
)

type Financial struct {
	ID               uint   `gorm:"primaryKey;autoIncrement:false;column:id"` // 明确指定不自动增长
	Name             string `gorm:"column:name"`
	Model            string `gorm:"column:model"`
	Quantity         string `gorm:"column:quantity"`
	Unit             string `gorm:"column:unit"`
	Price            string `gorm:"column:price"`
	ExtraPrice       string `gorm:"column:extra_price"`
	PurchaseLink     string `gorm:"column:purchase_link"`
	PostDate         string `gorm:"column:post_date"`
	Purchaser        string `gorm:"column:purchaser"`
	Campus           string `gorm:"column:campus"`
	GroupName        string `gorm:"column:group_name"`
	TroopTypeProject string `gorm:"column:troop_type"` // 映射数据库中的troop_type字段
	Remarks          string `gorm:"column:remarks"`
}

func (Financial) TableName() string {
	return "Financial_Log" // 表名为Financial_Log
}

func (f *Financial) BeforeCreate(tx *gorm.DB) (err error) {
	// 在创建数据之前执行的逻辑
	return nil
}

func (f *Financial) BeforeUpdate(tx *gorm.DB) (err error) {
	// 在更新数据之前执行的逻辑
	return nil
}
