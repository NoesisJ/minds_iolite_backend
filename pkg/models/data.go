package models

import (
	"gorm.io/gorm"
)

type Data struct {
	ID             uint   `gorm:"primaryKey;autoIncrement:true;column:id"`
	Nickname       string `gorm:"column:nickname"`
	IDCard         string `gorm:"column:IDcard"`
	Sex            string `gorm:"column:sex"`
	Age            string `gorm:"column:age" validate:"numeric"`
	Address        string `gorm:"column:address"`
	Classification string `gorm:"column:classification"`
	School         string `gorm:"column:school"`
	Subjects       string `gorm:"column:subjects"`
	Phone          string `gorm:"column:phone"`
	Email          string `gorm:"column:email"`
	QQ             string `gorm:"column:qq"`
	Wechat         string `gorm:"column:wechat"`
	WebID          string `gorm:"column:web_id"`
	JLUGroup       string `gorm:"column:jlugroup"`
	Study          string `gorm:"column:study"`
	Identity       string `gorm:"column:identity"`
	State          string `gorm:"column:state"`
}

func (Data) TableName() string {
	return "Data" // 注意首字母大写
}

func (d *Data) BeforeCreate(tx *gorm.DB) (err error) {
	// 在创建数据之前执行的逻辑
	return nil
}

func (d *Data) BeforeUpdate(tx *gorm.DB) (err error) {
	// 在更新数据之前执行的逻辑
	return nil
}
