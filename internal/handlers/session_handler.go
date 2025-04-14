package handlers

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// SessionState 表示会话的当前状态
type SessionState struct {
	Info        ConnectionInfo         `json:"info"`
	Connected   bool                   `json:"connected"`
	LastActive  time.Time              `json:"lastActive"`
	Collections map[string]interface{} `json:"collections,omitempty"`
	Tables      map[string]interface{} `json:"tables,omitempty"`
}

// ConnectionInfo 表示连接信息
type ConnectionInfo struct {
	Type     string `json:"type"`
	Host     string `json:"host,omitempty"`
	Port     int    `json:"port,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"-"` // 不在JSON中暴露密码
	Database string `json:"database,omitempty"`
	URI      string `json:"-"` // 不在JSON中暴露完整URI
}

// Manager 管理所有会话
type Manager struct {
	sessions        map[string]*SessionState
	mutex           sync.RWMutex
	cleanupInterval time.Duration // 每隔30分钟检查一次过期会话
	sessionTimeout  time.Duration // 会话超时时间为30分钟
}

// 全局会话管理器
var sessionManager *Manager

// 创建新的会话管理器
func NewManager() *Manager {
	manager := &Manager{
		sessions:        make(map[string]*SessionState),
		mutex:           sync.RWMutex{},
		cleanupInterval: 30 * time.Minute,
		sessionTimeout:  30 * time.Minute,
	}

	// 启动定期清理过期会话的goroutine
	go manager.cleanup()

	return manager
}

// CreateSession 创建新会话
func (m *Manager) CreateSession(info ConnectionInfo, collections, tables map[string]interface{}) string {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 生成唯一会话ID
	sessionID := uuid.New().String()

	// 创建会话状态
	state := &SessionState{
		Info:        info,
		Connected:   true,
		LastActive:  time.Now(),
		Collections: collections,
		Tables:      tables,
	}

	// 保存会话
	m.sessions[sessionID] = state

	return sessionID
}

// GetSession 获取特定会话状态
func (m *Manager) GetSession(sessionID string) (*SessionState, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	state, exists := m.sessions[sessionID]
	if !exists {
		return nil, false
	}

	// 检查会话是否过期
	if time.Since(state.LastActive) > m.sessionTimeout {
		// 会话已过期，但我们在读锁中，所以不能删除
		// 返回nil表示不存在
		return nil, false
	}

	return state, true
}

// RefreshSession 刷新会话活动时间
func (m *Manager) RefreshSession(sessionID string) bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	state, exists := m.sessions[sessionID]
	if !exists {
		return false
	}

	// 检查会话是否过期
	if time.Since(state.LastActive) > m.sessionTimeout {
		// 会话已过期，删除它
		delete(m.sessions, sessionID)
		return false
	}

	// 更新最后活动时间
	state.LastActive = time.Now()
	return true
}

// CloseSession 关闭并删除会话
func (m *Manager) CloseSession(sessionID string) bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	_, exists := m.sessions[sessionID]
	if !exists {
		return false
	}

	// 删除会话
	delete(m.sessions, sessionID)
	return true
}

// GetAllSessions 获取所有活动会话
func (m *Manager) GetAllSessions() map[string]*SessionState {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// 创建一个副本以避免并发访问问题
	result := make(map[string]*SessionState)
	for id, state := range m.sessions {
		// 只返回未过期的会话
		if time.Since(state.LastActive) <= m.sessionTimeout {
			result[id] = state
		}
	}

	return result
}

// cleanup 定期清理过期会话
func (m *Manager) cleanup() {
	ticker := time.NewTicker(m.cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		m.mutex.Lock()
		now := time.Now()
		for id, state := range m.sessions {
			if now.Sub(state.LastActive) > m.sessionTimeout {
				delete(m.sessions, id)
			}
		}
		m.mutex.Unlock()
	}
}

// SessionHandler 会话处理器
type SessionHandler struct{}

// NewSessionHandler 创建新的会话处理器
func NewSessionHandler() *SessionHandler {
	return &SessionHandler{}
}

// InitSessionManager 初始化会话管理器
func InitSessionManager() {
	sessionManager = NewManager()
}

// CreateSessionRequest 创建会话请求
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
	info := ConnectionInfo{
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
