package handlers

import (
	"net/http"

	"minds_iolite_backend/internal/session"

	"github.com/gin-gonic/gin"
)

// 全局会话管理器
var sessionManager *session.Manager

// NewSessionHandler 创建新的会话处理器
func NewSessionHandler() *SessionHandler {
	return &SessionHandler{}
}

// SessionHandler 处理会话相关请求
type SessionHandler struct {
}

// InitSessionManager 初始化全局会话管理器
func InitSessionManager() {
	sessionManager = session.NewManager()
}

// CreateSessionRequest 表示创建会话的请求
type CreateSessionRequest struct {
	Type          string                 `json:"type"`
	Host          string                 `json:"host,omitempty"`
	Port          int                    `json:"port,omitempty"`
	Username      string                 `json:"username,omitempty"`
	Password      string                 `json:"password,omitempty"`
	Database      string                 `json:"database,omitempty"`
	URI           string                 `json:"uri,omitempty"`
	FilePath      string                 `json:"filePath,omitempty"`
	Options       map[string]interface{} `json:"options,omitempty"`
	MongoDbName   string                 `json:"mongoDbName,omitempty"`
	MongoCollName string                 `json:"mongoCollName,omitempty"`
	Collections   map[string]interface{} `json:"collections,omitempty"`
}

// CreateSession 创建新的持久连接会话
func (h *SessionHandler) CreateSession(c *gin.Context) {
	var req CreateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "无效的请求数据: " + err.Error()})
		return
	}

	// 确保至少指定了类型
	if req.Type == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "必须指定数据库类型"})
		return
	}

	// 根据类型检查必要的参数
	if req.Type == "mongodb" {
		if req.URI == "" && (req.Host == "" || req.Database == "") {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "MongoDB连接需要提供URI或主机和数据库名",
			})
			return
		}
	} else if req.Type == "mysql" {
		if req.Host == "" || req.Database == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "MySQL连接需要提供主机和数据库名",
			})
			return
		}
	} else if req.Type == "csv" {
		if req.FilePath == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "CSV处理需要提供文件路径",
			})
			return
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "不支持的数据库类型: " + req.Type,
		})
		return
	}

	// 创建连接信息
	info := session.ConnectionInfo{
		Type:     req.Type,
		Host:     req.Host,
		Port:     req.Port,
		Username: req.Username,
		Password: req.Password,
		Database: req.Database,
		URI:      req.URI,
	}

	// 创建会话
	sessionID := sessionManager.CreateSession(info, req.Collections, nil)

	// 获取创建的会话状态
	state, _ := sessionManager.GetSession(sessionID)

	// 返回会话信息
	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"sessionId": sessionID,
		"state":     state,
	})
}

// GetSession 获取特定会话的信息
func (h *SessionHandler) GetSession(c *gin.Context) {
	sessionID := c.Param("sessionId")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "未提供会话ID"})
		return
	}

	state, exists := sessionManager.GetSession(sessionID)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "会话不存在或已过期"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"state":   state,
	})
}

// GetAllSessions 获取所有活动会话
func (h *SessionHandler) GetAllSessions(c *gin.Context) {
	sessions := sessionManager.GetAllSessions()
	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"sessions": sessions,
	})
}

// RefreshSession 刷新会话
func (h *SessionHandler) RefreshSession(c *gin.Context) {
	sessionID := c.Param("sessionId")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "未提供会话ID"})
		return
	}

	success := sessionManager.RefreshSession(sessionID)
	if !success {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "会话不存在或已过期"})
		return
	}

	state, _ := sessionManager.GetSession(sessionID)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"state":   state,
	})
}

// CloseSession 关闭会话
func (h *SessionHandler) CloseSession(c *gin.Context) {
	sessionID := c.Param("sessionId")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "未提供会话ID"})
		return
	}

	success := sessionManager.CloseSession(sessionID)
	if !success {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "会话不存在或已过期"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "会话已关闭",
	})
}
