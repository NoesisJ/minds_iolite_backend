package handlers

import (
	"net/http"

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

	// 创建CSV数据源
	csvSource := datasource.NewCSVSource(tempPath)
	csvSource.Delimiter = delimiter
	csvSource.HasHeader = hasHeader

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

	// 返回处理结果
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    model,
	})
}

// ImportCSVToMongoDB 处理将CSV导入MongoDB的请求
func (h *DataSourceHandler) ImportCSVToMongoDB(c *gin.Context) {
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
	connInfo, err := storage.ImportCSVToMongoDB(model, request.DbName, request.CollName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "导入数据到MongoDB失败: " + err.Error(),
		})
		return
	}

	// 直接返回连接信息，不包含success和data包装
	c.JSON(http.StatusOK, connInfo)
}

// ConnectToMongoDB 处理MongoDB连接请求
func (h *DataSourceHandler) ConnectToMongoDB(c *gin.Context) {
	var request struct {
		ConnectionURI string `json:"ConnectionURI"`
		Database      string `json:"Database" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的请求参数: " + err.Error(),
		})
		return
	}

	// 设置默认连接URI（如果未提供）
	uri := request.ConnectionURI
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
	connInfo, err := connector.ExtractConnectionInfo(request.Database)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "获取数据库信息失败: " + err.Error(),
		})
		return
	}

	// 返回连接信息
	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"connectionInfo": connInfo,
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
		Table    string `json:"table" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的请求参数: " + err.Error(),
		})
		return
	}

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

	// 生成指定表的连接信息
	connInfo, err := storage.GenerateConnectionInfoForTable(request.Table)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "获取数据库信息失败: " + err.Error(),
		})
		return
	}

	// 返回连接信息
	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"connectionInfo": connInfo,
	})
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

	// 返回结果
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    connInfo,
	})
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

	// 设置默认MongoDB连接URI
	mongoURI := request.MongoURI
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}

	// 导入数据到MongoDB
	connInfo, err := storage.ImportSQLiteToMongoDB(
		request.Table,
		request.DatabaseName,
		request.CollectionName,
		mongoURI,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "导入数据到MongoDB失败: " + err.Error(),
		})
		return
	}

	// 返回结果
	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"connectionInfo": connInfo,
		"message":        "SQLite数据已成功导入到MongoDB",
	})
}
