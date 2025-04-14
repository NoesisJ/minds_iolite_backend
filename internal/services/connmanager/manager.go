package connmanager

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/net/context"
)

// ConnectionType 定义连接类型
type ConnectionType string

const (
	MongoDB ConnectionType = "mongodb"
	MySQL   ConnectionType = "mysql"
	SQLite  ConnectionType = "sqlite"
)

// ConnectionInfo 存储连接信息
type ConnectionInfo struct {
	Type     ConnectionType `json:"type"`
	Host     string         `json:"host,omitempty"`
	Port     interface{}    `json:"port,omitempty"`
	Username string         `json:"username,omitempty"`
	Password string         `json:"password,omitempty"`
	Database string         `json:"database,omitempty"`
	FilePath string         `json:"filePath,omitempty"` // 用于SQLite
	URI      string         `json:"uri,omitempty"`      // 用于MongoDB
}

// ConnectionState 记录连接状态
type ConnectionState struct {
	Info        ConnectionInfo `json:"info"`
	Connected   bool           `json:"connected"`
	LastActive  time.Time      `json:"lastActive"`
	Error       string         `json:"error,omitempty"`
	MongoConn   *mongo.Client  `json:"-"`
	SQLConn     *sql.DB        `json:"-"`
	Tables      map[string]any `json:"tables,omitempty"`      // MySQL/SQLite表信息
	Collections map[string]any `json:"collections,omitempty"` // MongoDB集合信息
}

// SessionManager 管理会话和连接
type SessionManager struct {
	sessions        map[string]*ConnectionState
	maxIdle         time.Duration
	mutex           sync.RWMutex
	cleanupInterval time.Duration
}

// 创建新的会话管理器
func NewSessionManager() *SessionManager {
	manager := &SessionManager{
		sessions:        make(map[string]*ConnectionState),
		maxIdle:         30 * time.Minute, // 默认30分钟超时
		cleanupInterval: 5 * time.Minute,  // 每5分钟清理过期连接
	}
	go manager.cleanupRoutine()
	return manager
}

// 定期清理过期的连接
func (m *SessionManager) cleanupRoutine() {
	ticker := time.NewTicker(m.cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		m.CleanupSessions()
	}
}

// CleanupSessions 清理过期连接
func (m *SessionManager) CleanupSessions() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	now := time.Now()
	for sessionID, state := range m.sessions {
		if now.Sub(state.LastActive) > m.maxIdle {
			// 关闭连接
			m.closeConnection(state)
			delete(m.sessions, sessionID)
			log.Printf("已清理过期连接: %s (%s)", sessionID, state.Info.Type)
		}
	}
}

// GetSession 获取指定会话
func (m *SessionManager) GetSession(sessionID string) (*ConnectionState, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if session, exists := m.sessions[sessionID]; exists {
		// 更新最后活跃时间
		session.LastActive = time.Now()
		return session, nil
	}
	return nil, errors.New("会话不存在")
}

// CreateSession 创建新会话
func (m *SessionManager) CreateSession(sessionID string, info ConnectionInfo) (string, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 如果已存在，先关闭旧连接
	if oldSession, exists := m.sessions[sessionID]; exists {
		m.closeConnection(oldSession)
		delete(m.sessions, sessionID)
	}

	// 创建新的连接状态
	state := &ConnectionState{
		Info:       info,
		Connected:  false,
		LastActive: time.Now(),
	}

	// 根据类型建立连接
	var err error
	switch info.Type {
	case MongoDB:
		err = m.connectMongoDB(state)
	case MySQL:
		err = m.connectMySQL(state)
	case SQLite:
		err = m.connectSQLite(state)
	default:
		return "", errors.New("不支持的数据库类型")
	}

	if err != nil {
		state.Error = err.Error()
		return "", err
	}

	// 保存会话
	m.sessions[sessionID] = state
	return sessionID, nil
}

// CloseSession 关闭指定会话
func (m *SessionManager) CloseSession(sessionID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if session, exists := m.sessions[sessionID]; exists {
		m.closeConnection(session)
		delete(m.sessions, sessionID)
		return nil
	}
	return errors.New("会话不存在")
}

// 关闭连接
func (m *SessionManager) closeConnection(state *ConnectionState) {
	if state == nil {
		return
	}

	switch state.Info.Type {
	case MongoDB:
		if state.MongoConn != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			state.MongoConn.Disconnect(ctx)
			state.MongoConn = nil
		}
	case MySQL, SQLite:
		if state.SQLConn != nil {
			state.SQLConn.Close()
			state.SQLConn = nil
		}
	}
	state.Connected = false
}

// 连接MongoDB
func (m *SessionManager) connectMongoDB(state *ConnectionState) error {
	uri := state.Info.URI
	if uri == "" {
		// 从字段构建URI
		host := state.Info.Host
		if host == "" {
			host = "localhost"
		}

		// 处理端口
		portStr := "27017" // 默认端口
		if state.Info.Port != nil {
			switch v := state.Info.Port.(type) {
			case string:
				portStr = v
			case float64:
				portStr = fmt.Sprintf("%d", int(v))
			case int:
				portStr = fmt.Sprintf("%d", v)
			}
		}

		// 构建URI
		if state.Info.Username != "" && state.Info.Password != "" {
			uri = fmt.Sprintf("mongodb://%s:%s@%s:%s/%s",
				state.Info.Username,
				state.Info.Password,
				host,
				portStr,
				state.Info.Database)
		} else if state.Info.Username != "" {
			uri = fmt.Sprintf("mongodb://%s@%s:%s/%s",
				state.Info.Username,
				host,
				portStr,
				state.Info.Database)
		} else {
			uri = fmt.Sprintf("mongodb://%s:%s/%s",
				host,
				portStr,
				state.Info.Database)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return err
	}

	// 验证连接
	err = client.Ping(ctx, nil)
	if err != nil {
		return err
	}

	state.MongoConn = client
	state.Connected = true
	return nil
}

// 连接MySQL
func (m *SessionManager) connectMySQL(state *ConnectionState) error {
	// 构建DSN
	host := state.Info.Host
	if host == "" {
		host = "localhost"
	}

	// 处理端口
	port := 3306 // 默认端口
	if state.Info.Port != nil {
		switch v := state.Info.Port.(type) {
		case string:
			if p, err := strconv.Atoi(v); err == nil {
				port = p
			}
		case float64:
			port = int(v)
		case int:
			port = v
		}
	}

	// 构建DSN (Data Source Name)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=true&loc=Local",
		state.Info.Username,
		state.Info.Password,
		host,
		port,
		state.Info.Database)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}

	// 设置连接池
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)

	// 验证连接
	err = db.Ping()
	if err != nil {
		return err
	}

	state.SQLConn = db
	state.Connected = true
	return nil
}

// 连接SQLite
func (m *SessionManager) connectSQLite(state *ConnectionState) error {
	if state.Info.FilePath == "" {
		return errors.New("未提供SQLite文件路径")
	}

	db, err := sql.Open("sqlite3", state.Info.FilePath)
	if err != nil {
		return err
	}

	// 验证连接
	err = db.Ping()
	if err != nil {
		return err
	}

	state.SQLConn = db
	state.Connected = true
	return nil
}

// GetAllSessions 获取所有会话
func (m *SessionManager) GetAllSessions() map[string]*ConnectionState {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// 创建副本以避免并发问题
	sessions := make(map[string]*ConnectionState, len(m.sessions))
	for id, session := range m.sessions {
		sessions[id] = session
	}
	return sessions
}

// RefreshSession 刷新会话状态
func (m *SessionManager) RefreshSession(sessionID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	session, exists := m.sessions[sessionID]
	if !exists {
		return errors.New("会话不存在")
	}

	// 检查连接是否还活跃
	var err error
	switch session.Info.Type {
	case MongoDB:
		if session.MongoConn != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			err = session.MongoConn.Ping(ctx, nil)
		} else {
			err = errors.New("MongoDB连接已关闭")
		}
	case MySQL, SQLite:
		if session.SQLConn != nil {
			err = session.SQLConn.Ping()
		} else {
			err = errors.New("SQL连接已关闭")
		}
	}

	if err != nil {
		// 连接已断开，尝试重新连接
		session.Connected = false
		session.Error = err.Error()

		// 关闭旧连接
		m.closeConnection(session)

		// 尝试重新连接
		switch session.Info.Type {
		case MongoDB:
			err = m.connectMongoDB(session)
		case MySQL:
			err = m.connectMySQL(session)
		case SQLite:
			err = m.connectSQLite(session)
		}

		if err != nil {
			session.Error = err.Error()
			return err
		}
	}

	session.LastActive = time.Now()
	return nil
}

// 全局连接管理器实例
var globalManager *SessionManager
var once sync.Once

// GetManager 获取全局连接管理器实例
func GetManager() *SessionManager {
	once.Do(func() {
		globalManager = NewSessionManager()
	})
	return globalManager
}
