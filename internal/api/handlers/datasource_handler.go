package handlers

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"minds_iolite_backend/internal/datasource/providers/csv"
	"minds_iolite_backend/internal/datasource/providers/mongodb"
	"minds_iolite_backend/internal/datasource/providers/sqlite"
	"minds_iolite_backend/internal/models/datasource"
	"minds_iolite_backend/internal/services/datastorage"

	"github.com/gin-gonic/gin"
)

// DataSourceHandler 处理数据源相关请求
type DataSourceHandler struct {
}

// EncryptPassword 使用XOR加密密码并返回十六进制字符串
func EncryptPassword(password, key string) string {
	encrypted := make([]byte, len(password))
	for i := 0; i < len(password); i++ {
		encrypted[i] = password[i] ^ key[i%len(key)]
	}
	return hex.EncodeToString(encrypted)
}

// DecryptPassword 解密密码
func DecryptPassword(cipherHex, key string) (string, error) {
	cipher, err := hex.DecodeString(cipherHex)
	if err != nil {
		return "", err
	}

	decrypted := make([]byte, len(cipher))
	for i := 0; i < len(cipher); i++ {
		decrypted[i] = cipher[i] ^ key[i%len(key)]
	}
	return string(decrypted), nil
}

// NewDataSourceHandler 创建新的数据源处理器
func NewDataSourceHandler() *DataSourceHandler {
	return &DataSourceHandler{}
}

// ProcessCSVFile 处理CSV文件请求
func (h *DataSourceHandler) ProcessCSVFile(c *gin.Context) {
	var request struct {
		FilePath string                `json:"filePath" binding:"required"`
		Options  *datasource.CSVSource `json:"options"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的请求参数: " + err.Error(),
		})
		return
	}

	// 创建CSV数据源
	var csvSource *datasource.CSVSource
	if request.Options != nil {
		csvSource = request.Options
		csvSource.FilePath = request.FilePath
	} else {
		csvSource = datasource.NewCSVSource(request.FilePath)
	}

	// 验证数据源
	if err := csvSource.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "数据源验证失败: " + err.Error(),
		})
		return
	}

	// 创建解析器
	parser := csv.NewCSVParser(csvSource)

	// 解析CSV文件
	csvData, err := parser.Parse()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "解析CSV文件失败: " + err.Error(),
		})
		return
	}

	// 创建转换器
	converter := csv.NewCSVConverter(nil, nil)

	// 验证数据
	validationErrors := converter.ValidateData(csvData)

	// 转换为统一数据模型
	model, err := converter.ConvertToUnifiedModel(csvSource, csvData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "转换数据失败: " + err.Error(),
		})
		return
	}

	// 将验证错误添加到模型中
	if len(validationErrors) > 0 {
		model.Errors = append(model.Errors, validationErrors...)
	}

	// 这里可以添加与Agent的集成逻辑
	// TODO: 将统一数据模型传递给Agent进行处理

	// 返回处理结果
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    model,
	})
}

// GetColumnTypes 获取CSV文件的列类型
func (h *DataSourceHandler) GetColumnTypes(c *gin.Context) {
	var request struct {
		FilePath   string `json:"filePath" binding:"required"`
		Delimiter  string `json:"delimiter"`
		HasHeader  bool   `json:"hasHeader"`
		SampleSize int    `json:"sampleSize"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的请求参数: " + err.Error(),
		})
		return
	}

	// 创建CSV数据源
	csvSource := datasource.NewCSVSource(request.FilePath)
	if request.Delimiter != "" {
		csvSource.Delimiter = request.Delimiter
	}
	csvSource.HasHeader = request.HasHeader

	// 验证数据源
	if err := csvSource.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "数据源验证失败: " + err.Error(),
		})
		return
	}

	// 创建解析器
	parser := csv.NewCSVParser(csvSource)

	// 设置样本大小
	sampleSize := 100
	if request.SampleSize > 0 {
		sampleSize = request.SampleSize
	}

	// 推断列类型
	columnTypes, err := parser.DetectColumnTypes(sampleSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "推断列类型失败: " + err.Error(),
		})
		return
	}

	// 返回列类型
	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"columnTypes": columnTypes,
	})
}

// UploadCSVFile 处理CSV文件上传
func (h *DataSourceHandler) UploadCSVFile(c *gin.Context) {
	// 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "获取上传文件失败: " + err.Error(),
		})
		return
	}

	// 生成临时文件路径
	tempPath := "temp/" + file.Filename

	// 保存上传的文件
	if err := c.SaveUploadedFile(file, tempPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "保存上传文件失败: " + err.Error(),
		})
		return
	}

	// 获取选项
	delimiter := c.DefaultPostForm("delimiter", ",")
	hasHeader := c.DefaultPostForm("hasHeader", "true") == "true"
	encoding := c.DefaultPostForm("encoding", "utf-8")

	// 获取MongoDB导入参数
	importToMongo := c.DefaultPostForm("importToMongo", "false") == "true"
	dbName := c.DefaultPostForm("dbName", "")
	collName := c.DefaultPostForm("collName", "")

	// 创建CSV数据源
	csvSource := datasource.NewCSVSource(tempPath)
	csvSource.Delimiter = delimiter
	csvSource.HasHeader = hasHeader
	csvSource.Encoding = encoding

	// 验证数据源
	if err := csvSource.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "数据源验证失败: " + err.Error(),
		})
		return
	}

	// 如果需要导入到MongoDB
	if importToMongo {
		// 创建解析器
		parser := csv.NewCSVParser(csvSource)

		// 解析CSV文件
		csvData, err := parser.Parse()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "解析CSV文件失败: " + err.Error(),
			})
			return
		}

		// 创建转换器
		converter := csv.NewCSVConverter(nil, nil)

		// 转换为统一数据模型
		model, err := converter.ConvertToUnifiedModel(csvSource, csvData)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "转换数据失败: " + err.Error(),
			})
			return
		}

		// 创建MongoDB存储服务
		mongoURI := "mongodb://localhost:27017"
		storage, err := datastorage.NewMongoStorage(mongoURI)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "连接MongoDB失败: " + err.Error(),
			})
			return
		}
		defer storage.Close()

		// 如果未提供数据库名，默认使用csv_文件名
		if dbName == "" {
			fileName := filepath.Base(tempPath)
			fileNameWithoutExt := strings.TrimSuffix(fileName, filepath.Ext(fileName))
			dbName = "csv_" + fileNameWithoutExt
		}

		// 如果未提供集合名，默认使用"data"
		if collName == "" {
			collName = "data"
		}

		// 导入数据到MongoDB
		connInfo, err := storage.ImportCSVToMongoDB(model, dbName, collName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "导入数据到MongoDB失败: " + err.Error(),
			})
			return
		}

		// 获取配置信息的保存路径 - 修改为保存在可执行文件所在目录的data子目录
		// 获取当前工作目录
		// 获取可执行文件所在目录
		exePath, err := os.Executable()
		if err != nil {
			log.Printf("警告: 无法获取可执行文件路径: %v", err)
		} else {
			// 获取可执行文件的目录
			exeDir := filepath.Dir(exePath)

			// 设置 config.json 保存在可执行文件目录的 data 子目录中
			dataDir := filepath.Join(exeDir, "data")
			configPath := filepath.Join(dataDir, "config.json")

			// 确保 data 目录存在
			if err := os.MkdirAll(dataDir, 0755); err != nil {
				log.Printf("警告: 无法创建 data 目录: %v", err)
			} else {
				// 将配置信息保存到 config.json
				configData, err := json.MarshalIndent(connInfo, "", "  ")
				if err != nil {
					log.Printf("警告: 无法序列化配置数据: %v", err)
				} else {
					if err := os.WriteFile(configPath, configData, 0644); err != nil {
						log.Printf("警告: 无法保存配置到 %s: %v", configPath, err)
					} else {
						log.Printf("已将配置信息保存到: %s", configPath)
					}
				}
			}
		}

		// 直接返回连接信息
		c.JSON(http.StatusOK, connInfo)
		return
	}

	// 如果不需要导入到MongoDB，则返回上传成功信息
	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"filePath": tempPath,
		"fileSize": file.Size,
		"message":  "文件上传成功",
	})
}

// ImportCSVToMongoDB 处理将CSV导入MongoDB的请求
func (h *DataSourceHandler) ImportCSVToMongoDB(c *gin.Context) {
	var filePath string
	var dbName string
	var collName string
	var csvSource *datasource.CSVSource

	// 检查内容类型
	contentType := c.GetHeader("Content-Type")

	// 处理multipart/form-data类型 (文件上传)
	if strings.Contains(contentType, "multipart/form-data") {
		// 获取上传的文件
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "获取上传文件失败: " + err.Error(),
			})
			return
		}

		// 生成临时文件路径
		filePath = "temp/" + file.Filename

		// 保存上传的文件
		if err := c.SaveUploadedFile(file, filePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "保存上传文件失败: " + err.Error(),
			})
			return
		}

		// 获取CSV选项
		delimiter := c.DefaultPostForm("delimiter", ",")
		hasHeader := c.DefaultPostForm("hasHeader", "true") == "true"
		encoding := c.DefaultPostForm("encoding", "utf-8")

		// 获取MongoDB选项
		dbName = c.DefaultPostForm("dbName", "")
		collName = c.DefaultPostForm("collName", "")

		// 创建CSV数据源
		csvSource = datasource.NewCSVSource(filePath)
		csvSource.Delimiter = delimiter
		csvSource.HasHeader = hasHeader
		csvSource.Encoding = encoding

		// 记录文件上传
		log.Printf("通过文件上传方式接收CSV: %s, 分隔符: %s, 表头: %v",
			filePath, delimiter, hasHeader)
	} else {
		// 处理application/json类型 (原有逻辑)
		var request struct {
			FilePath string                `json:"filePath" binding:"required"`
			Options  *datasource.CSVSource `json:"options"`
			DbName   string                `json:"dbName"`
			CollName string                `json:"collName"`
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "无效的请求参数: " + err.Error(),
			})
			return
		}

		// 设置参数
		filePath = request.FilePath
		dbName = request.DbName
		collName = request.CollName

		// 创建CSV数据源
		if request.Options != nil {
			csvSource = request.Options
			csvSource.FilePath = filePath
		} else {
			csvSource = datasource.NewCSVSource(filePath)
		}

		// 记录服务器路径方式
		log.Printf("通过服务器路径接收CSV: %s", filePath)
	}

	// 验证数据源
	if err := csvSource.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "数据源验证失败: " + err.Error(),
		})
		return
	}

	// 创建解析器
	parser := csv.NewCSVParser(csvSource)

	// 解析CSV文件
	csvData, err := parser.Parse()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "解析CSV文件失败: " + err.Error(),
		})
		return
	}

	// 创建转换器
	converter := csv.NewCSVConverter(nil, nil)

	// 转换为统一数据模型
	model, err := converter.ConvertToUnifiedModel(csvSource, csvData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "转换数据失败: " + err.Error(),
		})
		return
	}

	// 创建MongoDB存储服务
	mongoURI := "mongodb://localhost:27017"
	storage, err := datastorage.NewMongoStorage(mongoURI)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "连接MongoDB失败: " + err.Error(),
		})
		return
	}
	defer storage.Close()

	// 导入数据到MongoDB
	connInfo, err := storage.ImportCSVToMongoDB(model, dbName, collName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "导入数据到MongoDB失败: " + err.Error(),
		})
		return
	}

	// 获取配置信息的保存路径 - 修改为保存在可执行文件所在目录的data子目录
	// 获取当前工作目录
	wd, err := os.Getwd()
	if err != nil {
		log.Printf("警告: 无法获取当前工作目录: %v", err)
	} else {
		// 设置config.json保存在当前目录的data目录
		dataDir := filepath.Join(wd, "data")
		configPath := filepath.Join(dataDir, "config.json")

		// 确保data目录存在
		if err := os.MkdirAll(dataDir, 0755); err != nil {
			log.Printf("警告: 无法创建data目录: %v", err)
		} else {
			// 将配置信息保存到config.json
			configData, err := json.MarshalIndent(connInfo, "", "  ")
			if err != nil {
				log.Printf("警告: 无法序列化配置数据: %v", err)
			} else {
				if err := os.WriteFile(configPath, configData, 0644); err != nil {
					log.Printf("警告: 无法保存配置到 %s: %v", configPath, err)
				} else {
					log.Printf("已将配置信息保存到: %s", configPath)
				}
			}
		}
	}

	// 直接返回连接信息，不包含success和data包装
	c.JSON(http.StatusOK, connInfo)
}

// ConnectToMongoDB 处理MongoDB连接请求
func (h *DataSourceHandler) ConnectToMongoDB(c *gin.Context) {
	// 支持两种请求格式：1. ConnectionURI + Database, 2. host + port + username + password + database
	var request struct {
		// 首选格式（原有格式）
		ConnectionURI string `json:"ConnectionURI" binding:"omitempty"`
		Database      string `json:"Database" binding:"omitempty"`

		// 备用格式（与MySQL保持一致）
		Host     string      `json:"host" binding:"omitempty"`
		Port     interface{} `json:"port" binding:"omitempty"` // 支持字符串或数字
		Username string      `json:"username" binding:"omitempty"`
		Password string      `json:"password" binding:"omitempty"`
		DbName   string      `json:"database" binding:"omitempty"`
	}

	// 支持字段大小写不敏感
	var rawData map[string]interface{}
	if err := c.ShouldBindJSON(&rawData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的请求格式: " + err.Error(),
		})
		return
	}

	// 尝试从原始数据中提取关键字段（不区分大小写）
	for key, value := range rawData {
		lowerKey := strings.ToLower(key)
		switch lowerKey {
		case "connectionuri", "connection_uri", "connection":
			if strValue, ok := value.(string); ok {
				request.ConnectionURI = strValue
			}
		case "database", "db", "dbname":
			if strValue, ok := value.(string); ok {
				if request.Database == "" { // 优先使用Database字段
					request.Database = strValue
				}
				if request.DbName == "" { // 备用使用database字段
					request.DbName = strValue
				}
			}
		case "host":
			if strValue, ok := value.(string); ok {
				request.Host = strValue
			}
		case "port":
			request.Port = value // 可以是字符串或数字
		case "username", "user":
			if strValue, ok := value.(string); ok {
				request.Username = strValue
			}
		case "password", "pwd", "passwd":
			if strValue, ok := value.(string); ok {
				request.Password = strValue
			}
		}
	}

	// 确保有数据库名
	dbName := request.Database
	if dbName == "" {
		dbName = request.DbName
	}
	if dbName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "缺少必要参数: 数据库名",
		})
		return
	}

	// 构建连接URI
	uri := request.ConnectionURI
	if uri == "" && request.Host != "" {
		// 从独立字段构建连接字符串
		portStr := "27017" // 默认端口

		// 处理端口值（可能是字符串或数字）
		if request.Port != nil {
			switch v := request.Port.(type) {
			case string:
				portStr = v
			case float64:
				portStr = fmt.Sprintf("%d", int(v))
			case int:
				portStr = fmt.Sprintf("%d", v)
			}
		}

		// 构建基本连接字符串
		uri = fmt.Sprintf("mongodb://%s:%s", request.Host, portStr)

		// 如果有用户名和密码，添加到URI
		if request.Username != "" {
			if request.Password != "" {
				uri = fmt.Sprintf("mongodb://%s:%s@%s:%s",
					request.Username, request.Password, request.Host, portStr)
			} else {
				uri = fmt.Sprintf("mongodb://%s@%s:%s",
					request.Username, request.Host, portStr)
			}
		}
	}

	// 设置默认连接URI（如果都未提供）
	if uri == "" {
		uri = "mongodb://localhost:27017"
	}

	// 创建MongoDB连接器
	connector, err := mongodb.NewMongoDBConnector(uri)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "连接MongoDB失败: " + err.Error(),
		})
		return
	}
	defer connector.Close()

	// 提取连接信息
	connInfo, err := connector.ExtractConnectionInfo(dbName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "获取数据库信息失败: " + err.Error(),
		})
		return
	}

	// 获取当前工作目录
	wd, err := os.Getwd()
	if err != nil {
		log.Printf("警告: 无法获取当前工作目录: %v", err)
	} else {
		// 设置config.json保存在当前目录的data目录
		dataDir := filepath.Join(wd, "data")
		configPath := filepath.Join(dataDir, "config.json")

		// 确保data目录存在
		if err := os.MkdirAll(dataDir, 0755); err != nil {
			log.Printf("警告: 无法创建data目录: %v", err)
		} else {
			// 将配置信息保存到config.json，外层包裹{"mysql": ...}
			wrapped := map[string]interface{}{"mysql": connInfo}
			configData, err := json.MarshalIndent(wrapped, "", "  ")
			if err != nil {
				log.Printf("警告: 无法序列化配置数据: %v", err)
			} else {
				if err := os.WriteFile(configPath, configData, 0644); err != nil {
					log.Printf("警告: 无法保存配置到 %s: %v", configPath, err)
				} else {
					log.Printf("已将配置信息保存到: %s", configPath)
				}
			}
		}
	}

	// 直接返回连接信息，外层包裹{"mysql": ...}
	c.JSON(http.StatusOK, gin.H{
		"mysql": connInfo,
	})
}

// ConnectToMySQL 处理MySQL连接请求
func (h *DataSourceHandler) ConnectToMySQL(c *gin.Context) {
	var request struct {
		Host     string `json:"host" binding:"required"`
		Port     int    `json:"port" binding:"required"`
		Username string `json:"username" binding:"required"`
		Password string `json:"password"`
		Database string `json:"database" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		// 检查是否可能是端口类型错误
		var rawData map[string]interface{}
		if bindErr := c.ShouldBindJSON(&rawData); bindErr == nil {
			// 尝试从原始数据中获取端口值并转换
			if portVal, ok := rawData["port"]; ok {
				// 如果是字符串类型，尝试转换为整数
				if portStr, isStr := portVal.(string); isStr {
					if portNum, err := strconv.Atoi(portStr); err == nil {
						request.Port = portNum
						request.Host = rawData["host"].(string)
						request.Username = rawData["username"].(string)
						request.Password = rawData["password"].(string)
						request.Database = rawData["database"].(string)
						// 修正了端口，继续处理
						goto ProcessRequest
					}
				}
			}
		}

		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的请求参数: " + err.Error(),
		})
		return
	}

ProcessRequest:
	// 创建MySQL存储服务
	storage, err := datastorage.NewMySQLStorage(
		request.Host,
		request.Port,
		request.Username,
		request.Password,
		request.Database,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "连接MySQL失败: " + err.Error(),
		})
		return
	}
	defer storage.Close()

	// 获取所有表的连接信息
	connInfo, err := storage.GenerateConnectionInfo()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "获取数据库信息失败: " + err.Error(),
		})
		return
	}

	// 对密码进行加密，使用固定密钥"TokugawaMatsuri"
	encryptedPassword := ""
	if request.Password != "" {
		encryptedPassword = EncryptPassword(request.Password, "TokugawaMatsuri")
	}

	// 将结构体转换为map以便操作和包装
	connInfoMap := map[string]interface{}{
		"host":     request.Host,
		"port":     request.Port,
		"username": request.Username,
		"database": request.Database,
		"tables":   connInfo.Tables,
	}

	// 添加加密后的密码
	if encryptedPassword != "" {
		connInfoMap["password"] = encryptedPassword
	}

	// 将连接信息包装到mysql对象中
	wrappedConnInfo := map[string]interface{}{
		"mysql": connInfoMap,
	}

	// 获取当前工作目录
	wd, err := os.Getwd()
	if err != nil {
		log.Printf("警告: 无法获取当前工作目录: %v", err)
	} else {
		// 设置config.json保存在当前目录的data目录
		dataDir := filepath.Join(wd, "data")
		configPath := filepath.Join(dataDir, "config.json")

		// 确保data目录存在
		if err := os.MkdirAll(dataDir, 0755); err != nil {
			log.Printf("警告: 无法创建data目录: %v", err)
		} else {
			// 将配置信息保存到config.json (包含mysql外层包装)
			configData, err := json.MarshalIndent(wrappedConnInfo, "", "  ")
			if err != nil {
				log.Printf("警告: 无法序列化配置数据: %v", err)
			} else {
				if err := os.WriteFile(configPath, configData, 0644); err != nil {
					log.Printf("警告: 无法保存配置到 %s: %v", configPath, err)
				} else {
					log.Printf("已将配置信息保存到: %s", configPath)
				}
			}
		}
	}

	// 直接返回包装后的连接信息
	c.JSON(http.StatusOK, wrappedConnInfo)
}

// ProcessSQLiteFile 处理本地SQLite文件
func (h *DataSourceHandler) ProcessSQLiteFile(c *gin.Context) {
	var request struct {
		FilePath string `json:"filePath" binding:"required"`
		Table    string `json:"table"` // 可选，指定要处理的表
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的请求参数: " + err.Error(),
		})
		return
	}

	// 创建SQLite数据源
	sqliteSource := datasource.NewSQLiteSource(request.FilePath)

	// 验证数据源
	if err := sqliteSource.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "数据源验证失败: " + err.Error(),
		})
		return
	}

	// 创建SQLite连接器
	connector, err := sqlite.NewSQLiteConnector(request.FilePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "连接SQLite数据库失败: " + err.Error(),
		})
		return
	}
	defer connector.Close()

	// 获取表结构信息
	var connInfo *datastorage.SQLiteConnectionInfo
	if request.Table != "" {
		// 获取指定表的信息
		connInfo, err = connector.ExtractTableConnectionInfo(request.Table)
	} else {
		// 获取所有表的信息
		connInfo, err = connector.ExtractConnectionInfo()
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "获取数据库信息失败: " + err.Error(),
		})
		return
	}

	// 获取当前工作目录
	wd, err := os.Getwd()
	if err != nil {
		log.Printf("警告: 无法获取当前工作目录: %v", err)
	} else {
		// 设置config.json保存在当前目录的data目录
		dataDir := filepath.Join(wd, "data")
		configPath := filepath.Join(dataDir, "config.json")

		// 确保data目录存在
		if err := os.MkdirAll(dataDir, 0755); err != nil {
			log.Printf("警告: 无法创建data目录: %v", err)
		} else {
			// 将配置信息保存到config.json
			configData, err := json.MarshalIndent(connInfo, "", "  ")
			if err != nil {
				log.Printf("警告: 无法序列化配置数据: %v", err)
			} else {
				if err := os.WriteFile(configPath, configData, 0644); err != nil {
					log.Printf("警告: 无法保存配置到 %s: %v", configPath, err)
				} else {
					log.Printf("已将配置信息保存到: %s", configPath)
				}
			}
		}
	}

	// 返回结果
	c.JSON(http.StatusOK, connInfo)
}

// ImportSQLiteToMongoDB 将SQLite数据导入MongoDB
func (h *DataSourceHandler) ImportSQLiteToMongoDB(c *gin.Context) {
	var request struct {
		FilePath       string `json:"filePath" binding:"required"` // SQLite文件路径
		Table          string `json:"table" binding:"required"`    // 要导入的表名
		MongoURI       string `json:"mongoUri"`                    // MongoDB连接URI
		DatabaseName   string `json:"dbName"`                      // MongoDB数据库名
		CollectionName string `json:"collName"`                    // MongoDB集合名
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的请求参数: " + err.Error(),
		})
		return
	}

	// 创建SQLite数据源
	sqliteSource := datasource.NewSQLiteSource(request.FilePath)
	sqliteSource.Table = request.Table

	// 验证数据源
	if err := sqliteSource.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "数据源验证失败: " + err.Error(),
		})
		return
	}

	// 创建SQLite存储服务
	storage, err := datastorage.NewSQLiteStorage(request.FilePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "连接SQLite数据库失败: " + err.Error(),
		})
		return
	}
	defer storage.Close()

	// 设置默认的MongoDB连接参数
	mongoURI := request.MongoURI
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}

	dbName := request.DatabaseName
	if dbName == "" {
		dbName = "sqlite_import"
	}

	collName := request.CollectionName
	if collName == "" {
		collName = request.Table
	}

	// 导入数据到MongoDB
	connInfo, err := storage.ImportSQLiteToMongoDB(
		mongoURI,
		dbName,
		collName,
		request.Table,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "导入数据到MongoDB失败: " + err.Error(),
		})
		return
	}

	// 获取当前工作目录
	wd, err := os.Getwd()
	if err != nil {
		log.Printf("警告: 无法获取当前工作目录: %v", err)
	} else {
		// 设置config.json保存在当前目录的data目录
		dataDir := filepath.Join(wd, "data")
		configPath := filepath.Join(dataDir, "config.json")

		// 确保data目录存在
		if err := os.MkdirAll(dataDir, 0755); err != nil {
			log.Printf("警告: 无法创建data目录: %v", err)
		} else {
			// 将配置信息保存到config.json
			configData, err := json.MarshalIndent(connInfo, "", "  ")
			if err != nil {
				log.Printf("警告: 无法序列化配置数据: %v", err)
			} else {
				if err := os.WriteFile(configPath, configData, 0644); err != nil {
					log.Printf("警告: 无法保存配置到 %s: %v", configPath, err)
				} else {
					log.Printf("已将配置信息保存到: %s", configPath)
				}
			}
		}
	}

	// 返回结果
	c.JSON(http.StatusOK, connInfo)
}
