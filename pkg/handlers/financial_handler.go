package handlers

import (
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
	Remarks      string `json:"remarks"`
}

// GetAllFinancial 处理获取所有财务数据的请求
func GetAllFinancial(c *gin.Context) {
	var records []models.Financial
	result := database.DB.Find(&records)
	if result.Error != nil {
		c.JSON(500, gin.H{"error": "财务数据获取失败"})
		return
	}

	// 转换为响应格式
	response := make([]GetFinancialResponse, len(records))
	for i, record := range records {
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
			TroopType:    record.TroopTypeProject, // 改回使用TroopTypeProject字段
			Remarks:      record.Remarks,
		}
	}

	c.JSON(200, response)
}
