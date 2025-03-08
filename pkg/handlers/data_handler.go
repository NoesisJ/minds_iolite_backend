package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/NoesisJ/minds_iolite_backend/pkg/database"
	"github.com/NoesisJ/minds_iolite_backend/pkg/models"
)

// GetDataResponse 定义响应结构
type GetDataResponse struct {
	ID       uint   `json:"id"`
	Nickname string `json:"nickname"`
	Sex      string `json:"sex"`
	Age      string `json:"age"`
	JLUGroup string `json:"jlugroup"`
	Identity string `json:"identity"`
	Study    string `json:"study"`
	School   string `json:"school"`
	Subjects string `json:"subjects"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	QQ       string `json:"qq"`
	Wechat   string `json:"wechat"`
}

// GetAllData 处理获取所有数据的请求
func GetAllData(c *gin.Context) {
	var records []models.Data
	result := database.DB.Find(&records)
	if result.Error != nil {
		c.JSON(500, gin.H{"error": "数据获取失败"})
		return
	}

	// 转换为响应格式
	response := make([]GetDataResponse, len(records))
	for i, record := range records {
		response[i] = GetDataResponse{
			ID:       record.ID,
			Nickname: record.Nickname,
			Sex:      record.Sex,
			Age:      record.Age,
			JLUGroup: record.JLUGroup,
			Identity: record.Identity,
			Study:    record.Study,
			School:   record.School,
			Subjects: record.Subjects,
			Phone:    record.Phone,
			Email:    record.Email,
			QQ:       record.QQ,
			Wechat:   record.Wechat,
		}
	}

	c.JSON(200, response)
}