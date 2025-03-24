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
	Project          string `gorm:"column:project"`    // 新增project字段
	Remarks          string `gorm:"column:remarks"`
}

func (Financial) TableName() string {
	return "Financial_Log" // 表名为Financial_Log
}

func (f *Financial) BeforeCreate(tx *gorm.DB) (err error) {
	// 可以在此添加创建前的验证逻辑
	// 例如：检查必填字段、设置默认值等
	return nil
}

func (f *Financial) BeforeUpdate(tx *gorm.DB) (err error) {
	// 可以在此添加更新前的验证逻辑
	// 例如：检查字段格式、记录修改历史等
	return nil
}

// CreateFinancial 创建新的财务记录
func CreateFinancial(db *gorm.DB, financial *Financial) error {
	result := db.Create(financial)
	return result.Error
}

// BatchCreateFinancials 批量创建财务记录
func BatchCreateFinancials(db *gorm.DB, financials []*Financial) error {
	result := db.Create(financials)
	return result.Error
}

// UpdateFinancial 根据ID更新财务记录
func UpdateFinancial(db *gorm.DB, financial *Financial) error {
	result := db.Save(financial)
	return result.Error
}

// UpdateFinancialByFields 根据ID更新财务记录的特定字段
func UpdateFinancialByFields(db *gorm.DB, id uint, updates map[string]interface{}) error {
	result := db.Model(&Financial{}).Where("id = ?", id).Updates(updates)
	return result.Error
}

// DeleteFinancial 根据ID删除财务记录
func DeleteFinancial(db *gorm.DB, id uint) error {
	result := db.Delete(&Financial{}, id)
	return result.Error
}

// BatchDeleteFinancials 批量删除财务记录
func BatchDeleteFinancials(db *gorm.DB, ids []uint) error {
	result := db.Delete(&Financial{}, ids)
	return result.Error
}

// GetFinancialByID 根据ID获取单条财务记录
func GetFinancialByID(db *gorm.DB, id uint) (*Financial, error) {
	var financial Financial
	result := db.First(&financial, id)
	return &financial, result.Error
}

// GetFinancialsByCondition 根据条件查询财务记录
func GetFinancialsByCondition(db *gorm.DB, condition map[string]interface{}, page, pageSize int) ([]*Financial, int64, error) {
	var financials []*Financial
	var total int64

	query := db.Model(&Financial{})

	// 添加查询条件
	for field, value := range condition {
		query = query.Where(field+" = ?", value)
	}

	// 获取总记录数
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	if page > 0 && pageSize > 0 {
		offset := (page - 1) * pageSize
		query = query.Offset(offset).Limit(pageSize)
	}

	// 获取记录
	err = query.Find(&financials).Error
	return financials, total, err
}

// GetFinancialsByDateRange 根据日期范围查询财务记录
func GetFinancialsByDateRange(db *gorm.DB, startDate, endDate string, page, pageSize int) ([]*Financial, int64, error) {
	var financials []*Financial
	var total int64

	query := db.Model(&Financial{}).Where("post_date BETWEEN ? AND ?", startDate, endDate)

	// 获取总记录数
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	if page > 0 && pageSize > 0 {
		offset := (page - 1) * pageSize
		query = query.Offset(offset).Limit(pageSize)
	}

	// 获取记录
	err = query.Find(&financials).Error
	return financials, total, err
}

// GetFinancialsByMultipleConditions 多条件组合查询
func GetFinancialsByMultipleConditions(db *gorm.DB, conditions map[string]interface{}, likeConditions map[string]string, page, pageSize int) ([]*Financial, int64, error) {
	var financials []*Financial
	var total int64

	query := db.Model(&Financial{})

	// 添加精确匹配条件
	for field, value := range conditions {
		query = query.Where(field+" = ?", value)
	}

	// 添加模糊匹配条件
	for field, value := range likeConditions {
		query = query.Where(field+" LIKE ?", "%"+value+"%")
	}

	// 获取总记录数
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	if page > 0 && pageSize > 0 {
		offset := (page - 1) * pageSize
		query = query.Offset(offset).Limit(pageSize)
	}

	// 获取记录
	err = query.Find(&financials).Error
	return financials, total, err
}

// CreateFinancialWithTransaction 在事务中创建财务记录
func CreateFinancialWithTransaction(tx *gorm.DB, financial *Financial) error {
	return tx.Create(financial).Error
}

// UpdateFinancialWithTransaction 在事务中更新财务记录
func UpdateFinancialWithTransaction(tx *gorm.DB, financial *Financial) error {
	return tx.Save(financial).Error
}

// DeleteFinancialWithTransaction 在事务中删除财务记录
func DeleteFinancialWithTransaction(tx *gorm.DB, id uint) error {
	return tx.Delete(&Financial{}, id).Error
}
