# Minds Iolite 

Minds Iolite Backend 是一个基于Golang的低代码平台后端实现，使用Gin作为Web框架，GORM进行数据操作，MongoDB作为数据库。该项目旨在为前端低代码平台提供灵活、高效的后端支持，允许用户通过拖拽方式构建信息管理系统，同时无需编写大量代码即可实现数据库操作。

## 已实现功能

### 核心功能
- ✅ 元数据模型定义和管理
- ✅ 动态API自动生成
- ✅ 数据验证和类型检查
- ✅ 高级查询、排序和分页
- ✅ MongoDB数据存储集成
- ✅ CSV数据源处理

### API功能
- ✅ 元数据管理API (`/metadata/*`)
- ✅ 动态生成的数据API (`/api/*`)
- ✅ 完整CRUD操作支持
- ✅ 过滤、排序和分页
- ✅ 数据源处理API (`/api/datasource/*`)

## 快速开始

### 依赖项
- Go 1.18+
- MongoDB 5.0+

### 安装和运行
1. 克隆仓库
2. 配置MongoDB连接 (见 config/config.yaml)
3. 运行 `go run cmd/server/main.go`
4. 服务器默认监听 `:8080`

### 使用示例
1. 创建模型定义:
   ```
   POST /metadata/models
   {
     "name": "product",
     "displayName": "产品",
     "fields": [...]
   }
   ```

2. 使用自动生成的API:
   ```
   POST /api/product
   GET /api/product
   GET /api/product/:id
   PUT /api/product/:id
   DELETE /api/product/:id
   ```

3. 使用高级查询:
   ```
   GET /api/product?category=electronics&sort=price:desc&page=1&pageSize=10
   ```

## CSV数据源处理

我们已经实现了CSV数据源处理功能，使系统能够从本地CSV文件中读取、解析和转换数据，并提供统一的数据模型供Agent处理。

### 功能特点
- 自动检测CSV文件列数据类型
- 支持不同分隔符和编码格式
- 灵活的参数配置
- 安全的文件路径处理
- 流式处理大文件
- 统一的数据验证和错误报告

### 组件结构

#### 1. 数据源模型 (internal/models/datasource)
- **CSVSource**：定义CSV数据源配置，包括文件路径、分隔符等
- **UnifiedDataModel**：统一数据模型，用于在不同数据源和Agent间传递数据
- **Column/ColumnType**：数据列定义和类型系统
- **ValidationError**：数据验证错误表示

#### 2. CSV解析和转换 (internal/datasource/providers/csv)
- **CSVParser**：负责读取和解析CSV文件，自动检测数据类型
- **CSVConverter**：将CSV数据转换为统一数据模型
- **各种辅助函数**：数据类型判断、格式转换等

#### 3. API处理器 (internal/api/handlers)
- **DataSourceHandler**：处理CSV相关HTTP请求
- **处理本地CSV文件**：通过路径访问和处理CSV
- **支持文件上传**：接收上传的CSV文件并处理

### API接口

1. **处理本地CSV文件**：
   ```
   URL: http://localhost:8080/api/datasource/csv/process
   方法: POST
   Content-Type: application/json

   请求体:
   {
     "filePath": "E:/path/to/your/file.csv",
     "options": {
       "delimiter": ",",
       "hasHeader": true,
       "encoding": "utf-8"
     }
   }

   响应:
   {
     "success": true,
     "data": {
       "totalRows": 1000,
       "columns": ["id", "name", "age", "email"],
       "previewData": [
         {"id": "1", "name": "张三", "age": "30", "email": "zhangsan@example.com"},
         // 更多预览数据...
       ]
     }
   }
   ```

2. **获取CSV列类型**：
   ```
   URL: http://localhost:8080/api/datasource/csv/column-types
   方法: POST
   Content-Type: application/json

   请求体:
   {
     "filePath": "E:/path/to/your/file.csv",
     "delimiter": ",",
     "hasHeader": true,
     "sampleSize": 100
   }

   响应:
   {
     "success": true,
     "columnTypes": {
       "id": "integer",
       "name": "string",
       "age": "integer",
       "email": "string"
     }
   }
   ```

3. **上传CSV文件**：
   ```
   URL: http://localhost:8080/api/datasource/csv/upload
   方法: POST
   Content-Type: multipart/form-data

   表单字段:
   file: [CSV文件]
   delimiter: ,
   hasHeader: true

   响应:
   {
     "success": true,
     "filePath": "E:/uploaded/files/data.csv",
     "fileSize": 1024,
     "message": "文件上传成功"
   }
   ```

4. **将CSV导入MongoDB**：
   ```
   URL: http://localhost:8080/api/datasource/csv/import-to-mongo
   方法: POST
   Content-Type: application/json

   请求体:
   {
     "filePath": "E:/path/to/your/file.csv",
     "options": {
       "delimiter": ",",
       "hasHeader": true,
       "encoding": "utf-8"
     },
     "dbName": "可选的数据库名",
     "collName": "可选的集合名"
   }

   响应:
   {
     "success": true,
     "connectionInfo": {
       "host": "localhost",
       "port": 27017,
       "database": "csv_data",
       "collections": {
         "data": {
           "fields": {
             "id": "int",
             "name": "str",
             "age": "int",
             "email": "str"
           },
           "sampleData": "{\"_id\":\"...\",\"id\":1,\"name\":\"张三\",\"age\":30,\"email\":\"zhangsan@example.com\"}"
         }
       }
     },
     "message": "CSV数据已成功导入到MongoDB"
   }
   ```

## 组件功能详解

### 数据源模型 (internal/models/datasource)

#### csv_source.go
- **CSVSource** - 定义CSV数据源配置
  - `FilePath` - CSV文件路径
  - `Delimiter` - 分隔符，默认为逗号
  - `HasHeader` - 是否有表头
  - `SkipRows` - 跳过起始行数
  - `Encoding` - 文件编码
  - `ColumnTypes` - 列数据类型映射
- **NewCSVSource()** - 创建带默认值的CSV数据源
- **Validate()** - 验证配置有效性和文件可访问性
- **GetDelimiterRune()** - 获取分隔符的rune表示

#### data_model.go
- **ColumnType** - 数据列类型枚举（字符串、整数、浮点等）
- **Column** - 数据列定义结构
- **DataMetadata** - 数据集元数据信息
- **ValidationError** - 数据验证错误结构
- **UnifiedDataModel** - 统一数据模型结构
  - `Metadata` - 元数据信息
  - `Columns` - 列定义
  - `Records` - 数据记录
  - `TotalRecords` - 总记录数
  - `Errors` - 验证错误列表
- **NewUnifiedDataModel()** - 创建新的统一数据模型

### CSV解析和转换 (internal/datasource/providers/csv)

#### parser.go
- **CSVParser** - CSV文件解析器
- **CSVData** - 解析后的CSV数据结构
- **NewCSVParser()** - 创建新的解析器
- **Parse()** - 解析整个CSV文件
- **ParseStream()** - 流式解析大文件
- **DetectColumnTypes()** - 推断列数据类型
- **inferColumnTypes()** - 从数据推断列类型
- **validateFilePath()** - 验证文件路径安全性
- **normalizeHeader()** - 规范化列标题
- **isInteger()/isFloat()/isBoolean()/isDate()** - 类型判断辅助函数

#### converter.go
- **CSVConverter** - CSV数据转换器
- **NewCSVConverter()** - 创建新的转换器
- **ConvertToUnifiedModel()** - 将CSV数据转换为统一模型
- **ValidateData()** - 验证CSV数据
- **convertValue()** - 根据类型转换值
- **validateValue()** - 验证值是否符合类型要求
- **getDisplayName()** - 生成友好的显示名称
- **parseBool()/parseDate()** - 辅助函数

### API处理器 (internal/api/handlers)

#### datasource_handler.go
- **DataSourceHandler** - 数据源处理器
- **NewDataSourceHandler()** - 创建数据源处理器
- **ProcessCSVFile()** - 处理本地CSV文件路径
- **GetColumnTypes()** - 获取CSV列类型
- **UploadCSVFile()** - 处理文件上传

### 路由 (internal/routes)

#### datasource_routes.go
- **SetupDataSourceRoutes()** - 注册数据源相关路由
  - `/api/datasource/csv/process` - 处理CSV
  - `/api/datasource/csv/column-types` - 获取列类型
  - `/api/datasource/csv/upload` - 上传CSV

## 后续开发计划
- [ ] 用户认证和权限控制
- [ ] 模型关系和联合查询
- [ ] 自定义业务逻辑挂钩
- [ ] 前端界面开发
- [ ] 数据可视化和报表
- [ ] 其他数据源集成 (MongoDB、MySQL)

## 数据源集成计划

### 待实现的数据源
- [x] CSV文件导入与解析
- [ ] MongoDB数据库连接与查询
- [ ] MySQL数据库连接与查询

## MongoDB和MySQL数据源集成

不同于CSV处理直接使用本地文件路径，MongoDB和MySQL数据源需要通过连接URL进行访问。以下是对这两种数据源的实现规划：

### MongoDB数据源

#### 连接方式
MongoDB使用URI连接字符串格式：
```
mongodb://[username:password@]host[:port][/database][?options]
```

#### 实现功能
- 连接验证和测试
- 数据库和集合列表获取
- 集合结构和样本数据提取
- 数据查询和转换
- 本地MongoDB镜像创建（可选）

#### API设计
```
POST /api/datasource/mongodb/connect
Content-Type: application/json

{
  "ConnectionURI": "mongodb://localhost:27017",
  "Database": "database_name"
}
```

#### 返回格式
```json
{
  "success": true,
  "connectionInfo": {
    "host": "localhost",
    "port": 27017,
    "username": "",
    "password": "",
    "database": "database_name",
    "collections": {
      "collection_name": {
        "fields": {
          "_id": "ObjectId",
          "field1": "str",
          "field2": "int"
        },
        "sample_data": "{\"_id\": \"...\", \"field1\": \"value\", ...}"
      }
    }
  }
}
```

### 使用示例

连接到MongoDB数据库:
```
curl -X POST http://localhost:8080/api/datasource/mongodb/connect \
  -H "Content-Type: application/json" \
  -d '{
    "ConnectionURI": "mongodb://localhost:27017",
    "Database": "csv_customers"
  }'
```

响应示例:
```json
{
  "success": true,
  "connectionInfo": {
    "host": "localhost",
    "port": 27017,
    "username": "",
    "password": "",
    "database": "csv_customers",
    "collections": {
      "data": {
        "fields": {
          "_id": "ObjectId",
          "age": "int",
          "city": "str",
          "credit_score": "int",
          "email": "str",
          "id": "int",
          "is_active": "bool",
          "name": "str",
          "registration_date": "date"
        },
        "sample_data": "{\"_id\":\"67f7c9e79f2d90b8bcdfa47c\",\"age\":35,\"city\":\"北京\",\"credit_score\":720,\"email\":\"zhangsan@example.com\",\"id\":1,\"is_active\":true,\"name\":\"张三\",\"registration_date\":\"2022-01-15T00:00:00Z\"}"
      }
    }
  }
}
```

### MySQL数据源

#### 连接方式
MySQL使用DSN (Data Source Name) 连接字符串格式：
```
username:password@tcp(host:port)/database?param=value
```

#### 实现功能
- 数据库连接和验证
- 表结构和关系提取
- SQL查询支持
- 表数据样本获取
- 本地数据镜像（可选）

#### API设计
```
POST /api/datasource/mysql/connect
Content-Type: application/json

{
  "host": "localhost",
  "port": 3306,
  "username": "dbuser",
  "password": "dbpassword",
  "database": "mydatabase",
  "table": "tablename"
}
```

#### 返回格式
```json
{
  "success": true,
  "connectionInfo": {
    "host": "localhost",
    "port": 3306,
    "username": "dbuser",
    "password": "",
    "database": "mydatabase",
    "tables": {
      "tablename": {
        "fields": {
          "id": "int",
          "name": "str",
          "email": "str",
          "created_at": "date"
        },
        "sample_data": "{\"id\": 1, \"name\": \"张三\", \"email\": \"zhangsan@example.com\", \"created_at\": \"2024-03-26T10:30:00Z\"}"
      }
    }
  }
}
```

### 使用示例

连接到MySQL数据库特定表:
```bash
curl -X POST http://localhost:8080/api/datasource/mysql/connect \
  -H "Content-Type: application/json" \
  -d '{
    "host": "tarsgo.com",
    "port": 3306,
    "username": "tarsgo",
    "password": "xf210398444@",
    "database": "tarsgo",
    "table": "members"
  }'
```

响应示例:
```json
{
  "success": true,
  "connectionInfo": {
    "host": "tarsgo.com",
    "port": 3306,
    "username": "tarsgo",
    "database": "tarsgo",
    "tables": {
      "members": {
        "fields": {
          "id": "int",
          "name": "str",
          "email": "str",
          "phone": "str",
          "gender": "str",
          "grade": "str",
          "major": "str",
          "campus": "str",
          "branch": "str",
          "group": "str",
          "identity": "str",
          "qq": "str",
          "we_chat": "str",
          "created_at": "date",
          "updated_at": "date"
        },
        "sample_data": "{\"id\": 1, \"name\": \"张三\", \"email\": \"zhangsan@example.com\", ...}"
      }
    }
  }
}
```

注意事项：
1. `table` 参数现在是必需的，用于指定要连接的具体表
2. 返回的数据只包含指定表的结构和样本数据
3. 样本数据会返回表中的一条实际记录（如果存在）
4. 所有字段类型都会被映射为统一的类型表示（int, str, date, float, bool, binary）

## agent连接信息示例
{
  "host": "localhost",
  "port": 27017,
  "username": "",
  "password": "",
  "database": "company",
  "collections": {
    "departments": {
      "fields": {
        "_id": "ObjectId",
        "名字": "str",
        "部门": "str"
      },
      "sample_data":"{\"_id\": ObjectId(\"67e50e0900ce029f7ac66046\"), \"名字\": \"孙七\", \"部门\": \"销售部\"}"
    },
    "attendance": {
      "fields": {
        "_id": "ObjectId",
        "姓名": "str",
        "日期": "str",
        "考勤": "str"
      },
      "sample_data":"{\"_id\": ObjectId(\"67e50e146add66f28b6746dc\"), \"姓名\": \"张三\", \"日期\": \"2024-03-25\", \"考勤\": \"出勤\"}"
    }
  }
}

### 安全考虑

1. **敏感信息保护**：
   - 不在日志中记录完整连接字符串
   - 在返回结构中隐藏密码信息
   - 支持加密存储连接信息

2. **连接限制**：
   - 增加连接超时设置
   - 限制并发连接数
   - 支持只读模式连接

3. **权限管理**：
   - 验证用户拥有足够的数据库权限
   - 建议使用最小权限原则配置的账户
   - 提供连接授权验证机制 