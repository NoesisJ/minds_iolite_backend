package handlers

import (
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/NoesisJ/minds_iolite_backend/pkg/database"
	"github.com/NoesisJ/minds_iolite_backend/pkg/models"
	"github.com/gin-gonic/gin"
)

// GetFinancialResponse 定义财务数据响应结构
type GetFinancialResponse struct {
	ID           uint   `json:"id"`
	Name         string `json:"name"`
	Model        string `json:"model"`
	Quantity     string `json:"quantity"`
	Unit         string `json:"unit"`
	Price        string `json:"price"`
	ExtraPrice   string `json:"extra_price"`
	PurchaseLink string `json:"purchase_link"`
	PostDate     string `json:"post_date"`
	Purchaser    string `json:"purchaser"`
	Campus       string `json:"campus"`
	GroupName    string `json:"group_name"`
	TroopType    string `json:"troop_type"`
	Project      string `json:"project"`
	Remarks      string `json:"remarks"`
}

// FinancialRequest 定义请求体结构
type FinancialRequest struct {
	Name         string `json:"name" binding:"required"`
	Model        string `json:"model"`
	Quantity     string `json:"quantity"`
	Unit         string `json:"unit"`
	Price        string `json:"price"`
	ExtraPrice   string `json:"extra_price"`
	PurchaseLink string `json:"purchase_link"`
	PostDate     string `json:"post_date"`
	Purchaser    string `json:"purchaser"`
	Campus       string `json:"campus"`
	GroupName    string `json:"group_name"`
	TroopType    string `json:"troop_type"`
	Project      string `json:"project"`
	Remarks      string `json:"remarks"`
}

// BatchDeleteRequest 批量删除请求体
type BatchDeleteRequest struct {
	Ids []uint `json:"ids" binding:"required"`
}

// 辅助函数：从TroopTypeProject字段中提取troop_type和project
func extractTroopTypeAndProject(combined string) (troopType, project string) {
	parts := strings.Split(combined, " - ")
	if len(parts) > 1 {
		return parts[0], parts[1]
	}
	return combined, ""
}

// GetAllFinancial 处理获取所有财务数据的请求
func GetAllFinancial(c *gin.Context) {
	var records []models.Financial
	result := database.DB.Find(&records)
	if result.Error != nil {
		c.JSON(500, gin.H{"success": false, "error": "财务数据获取失败", "data": nil})
		return
	}

	// 转换为响应格式
	response := make([]GetFinancialResponse, len(records))
	for i, record := range records {
		troopType, project := extractTroopTypeAndProject(record.TroopTypeProject)
		response[i] = GetFinancialResponse{
			ID:           record.ID,
			Name:         record.Name,
			Model:        record.Model,
			Quantity:     record.Quantity,
			Unit:         record.Unit,
			Price:        record.Price,
			ExtraPrice:   record.ExtraPrice,
			PurchaseLink: record.PurchaseLink,
			PostDate:     record.PostDate,
			Purchaser:    record.Purchaser,
			Campus:       record.Campus,
			GroupName:    record.GroupName,
			TroopType:    troopType,
			Project:      project,
			Remarks:      record.Remarks,
		}
	}

	// 使用统一的响应格式
	c.JSON(200, gin.H{"success": true, "error": nil, "data": response})
}

// GetFinancialByID 获取单个财务数据记录
func GetFinancialByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "无效的ID参数", "data": nil})
		return
	}

	var record models.Financial
	result := database.DB.First(&record, uint(id))
	if result.Error != nil {
		c.JSON(404, gin.H{"success": false, "error": "未找到该记录", "data": nil})
		return
	}

	troopType, project := extractTroopTypeAndProject(record.TroopTypeProject)
	response := GetFinancialResponse{
		ID:           record.ID,
		Name:         record.Name,
		Model:        record.Model,
		Quantity:     record.Quantity,
		Unit:         record.Unit,
		Price:        record.Price,
		ExtraPrice:   record.ExtraPrice,
		PurchaseLink: record.PurchaseLink,
		PostDate:     record.PostDate,
		Purchaser:    record.Purchaser,
		Campus:       record.Campus,
		GroupName:    record.GroupName,
		TroopType:    troopType,
		Project:      project,
		Remarks:      record.Remarks,
	}

	c.JSON(200, gin.H{"success": true, "error": nil, "data": response})
}

// CreateFinancial 创建新的财务记录
func CreateFinancial(c *gin.Context) {
	var req FinancialRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "请求参数有误: " + err.Error(), "data": nil})
		return
	}

	// 验证日期格式和有效范围
	postDate, err := time.Parse("2006-01-02", req.PostDate)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "日期格式无效，请使用YYYY-MM-DD格式", "data": nil})
		return
	}

	// 验证日期范围 (MySQL通常支持1000-01-01到9999-12-31，但为安全起见我们限制为2000年以后)
	minDate, _ := time.Parse("2006-01-02", "2000-01-01")
	maxDate, _ := time.Parse("2006-01-02", "2099-12-31")
	if postDate.Before(minDate) || postDate.After(maxDate) {
		c.JSON(400, gin.H{"success": false, "error": "日期超出有效范围，请使用2000-01-01至2099-12-31之间的日期", "data": nil})
		return
	}

	// 合并troop_type和project
	troopTypeProject := req.TroopType
	if req.Project != "" {
		troopTypeProject = troopTypeProject + " - " + req.Project
	}

	// 创建财务记录
	financial := models.Financial{
		Name:             req.Name,
		Model:            req.Model,
		Quantity:         req.Quantity,
		Unit:             req.Unit,
		Price:            req.Price,
		ExtraPrice:       req.ExtraPrice,
		PurchaseLink:     req.PurchaseLink,
		PostDate:         req.PostDate,
		Purchaser:        req.Purchaser,
		Campus:           req.Campus,
		GroupName:        req.GroupName,
		TroopTypeProject: troopTypeProject,
		Project:          req.Project,
		Remarks:          req.Remarks,
	}

	// 保存记录
	result := database.DB.Create(&financial)
	if result.Error != nil {
		var errorMsg string
		if strings.Contains(result.Error.Error(), "Incorrect datetime") {
			errorMsg = "日期格式错误，请使用有效的日期(2000年至2099年之间)"
		} else {
			errorMsg = "创建记录失败: " + result.Error.Error()
		}
		c.JSON(500, gin.H{"success": false, "error": "服务器错误: " + errorMsg, "data": nil})
		return
	}

	// 返回成功响应
	c.JSON(201, gin.H{"success": true, "error": nil, "data": financial})
}

// UpdateFinancial 更新财务记录
func UpdateFinancial(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "无效的ID参数", "data": nil})
		return
	}

	// 检查记录是否存在
	var existingRecord models.Financial
	result := database.DB.First(&existingRecord, uint(id))
	if result.Error != nil {
		c.JSON(404, gin.H{"success": false, "error": "未找到该记录", "data": nil})
		return
	}

	var req FinancialRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "请求参数有误: " + err.Error(), "data": nil})
		return
	}

	// 验证日期格式和有效范围
	postDate, err := time.Parse("2006-01-02", req.PostDate)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "日期格式无效，请使用YYYY-MM-DD格式", "data": nil})
		return
	}

	// 验证日期范围
	minDate, _ := time.Parse("2006-01-02", "2000-01-01")
	maxDate, _ := time.Parse("2006-01-02", "2099-12-31")
	if postDate.Before(minDate) || postDate.After(maxDate) {
		c.JSON(400, gin.H{"success": false, "error": "日期超出有效范围，请使用2000-01-01至2099-12-31之间的日期", "data": nil})
		return
	}

	// 添加这部分代码处理字段映射
	troopTypeProject := req.TroopType
	if req.Project != "" {
		troopTypeProject = troopTypeProject + " - " + req.Project
	}

	// 更新记录
	updateFields := models.Financial{
		ID:               uint(id),
		Name:             req.Name,
		Model:            req.Model,
		Quantity:         req.Quantity,
		Unit:             req.Unit,
		Price:            req.Price,
		ExtraPrice:       req.ExtraPrice,
		PurchaseLink:     req.PurchaseLink,
		PostDate:         req.PostDate,
		Purchaser:        req.Purchaser,
		Campus:           req.Campus,
		GroupName:        req.GroupName,
		TroopTypeProject: troopTypeProject,
		Project:          req.Project,
		Remarks:          req.Remarks,
	}

	result = database.DB.Save(&updateFields)
	if result.Error != nil {
		c.JSON(500, gin.H{"success": false, "error": "更新记录失败: " + result.Error.Error(), "data": nil})
		return
	}

	// 正确拆分TroopTypeProject为两个字段
	troopType, project := extractTroopTypeAndProject(updateFields.TroopTypeProject)

	// 返回更新后的记录
	response := GetFinancialResponse{
		ID:           updateFields.ID,
		Name:         updateFields.Name,
		Model:        updateFields.Model,
		Quantity:     updateFields.Quantity,
		Unit:         updateFields.Unit,
		Price:        updateFields.Price,
		ExtraPrice:   updateFields.ExtraPrice,
		PurchaseLink: updateFields.PurchaseLink,
		PostDate:     updateFields.PostDate,
		Purchaser:    updateFields.Purchaser,
		Campus:       updateFields.Campus,
		GroupName:    updateFields.GroupName,
		TroopType:    troopType,
		Project:      project,
		Remarks:      updateFields.Remarks,
	}

	c.JSON(200, gin.H{"success": true, "error": nil, "data": response})
}

// DeleteFinancial 删除财务记录
func DeleteFinancial(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "无效的ID参数", "data": nil})
		return
	}

	// 检查记录是否存在
	var existingRecord models.Financial
	result := database.DB.First(&existingRecord, uint(id))
	if result.Error != nil {
		c.JSON(404, gin.H{"success": false, "error": "未找到该记录", "data": nil})
		return
	}

	// 删除记录
	result = database.DB.Delete(&models.Financial{}, uint(id))
	if result.Error != nil {
		c.JSON(500, gin.H{"success": false, "error": "删除记录失败: " + result.Error.Error(), "data": nil})
		return
	}

	c.JSON(200, gin.H{"success": true, "error": nil, "data": nil})
}

// BatchDeleteFinancial 批量删除财务记录
func BatchDeleteFinancial(c *gin.Context) {
	var req BatchDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "请求参数有误: " + err.Error(), "data": nil})
		return
	}

	if len(req.Ids) == 0 {
		c.JSON(400, gin.H{"success": false, "error": "请至少提供一个ID", "data": nil})
		return
	}

	// 批量删除记录
	result := database.DB.Delete(&models.Financial{}, req.Ids)
	if result.Error != nil {
		c.JSON(500, gin.H{"success": false, "error": "批量删除记录失败: " + result.Error.Error(), "data": nil})
		return
	}

	c.JSON(200, gin.H{"success": true, "error": nil, "data": nil})
}

// ExportFinancialCSV 导出财务数据为CSV
func ExportFinancialCSV(c *gin.Context) {
	var records []models.Financial
	result := database.DB.Find(&records)
	if result.Error != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取财务数据失败", "data": nil})
		return
	}

	// 设置CSV响应头
	currentTime := time.Now().Format("2006-01-02")
	fileName := fmt.Sprintf("财务数据_%s.csv", currentTime)
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))

	// 创建CSV写入器
	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()

	// 写入CSV标头
	headers := []string{"ID", "名称", "型号", "数量", "单位", "价格", "额外价格", "购买链接",
		"日期", "采购人", "校区", "组别", "兵种", "项目", "备注"}
	writer.Write(headers)

	// 写入CSV数据行
	for _, record := range records {
		troopType, project := extractTroopTypeAndProject(record.TroopTypeProject)
		row := []string{
			strconv.FormatUint(uint64(record.ID), 10),
			record.Name,
			record.Model,
			record.Quantity,
			record.Unit,
			record.Price,
			record.ExtraPrice,
			record.PurchaseLink,
			record.PostDate,
			record.Purchaser,
			record.Campus,
			record.GroupName,
			troopType,
			project,
			record.Remarks,
		}
		writer.Write(row)
	}
}

// CheckFinancialAPIHealth 检查财务API的健康状态
func CheckFinancialAPIHealth(c *gin.Context) {
	// 简单检查数据库连接
	var count int64
	result := database.DB.Model(&models.Financial{}).Count(&count)

	if result.Error != nil {
		c.JSON(500, gin.H{"status": "error", "message": "数据库连接异常"})
		return
	}

	c.JSON(200, gin.H{"status": "ok", "message": "财务API正常运行", "recordCount": count})
}
