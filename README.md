# Minds Iolite Backend API 文档

提供了Minds Iolite Backend系统的所有API接口定义

**对于每个会话前端需要在创建时返回一个唯一标识

## 基础信息

- 基础URL: `http://localhost:8080`
- 内容类型: `application/json`

## 跨域(CORS)支持

本API服务已配置跨域资源共享(CORS)支持，允许从不同域名的前端应用程序访问：

- 允许所有来源(`*`)的请求
- 支持的请求方法: `GET`, `POST`, `PUT`, `DELETE`, `OPTIONS`
- 支持的请求头: `Origin`, `Content-Type`, `Content-Length`, `Accept-Encoding`, `X-CSRF-Token`, `Authorization`
- 允许携带凭证(Credentials)
- 预检请求(OPTIONS)缓存时间为12小时

在生产环境部署时，建议将`AllowOrigins`设置为特定的前端域名，以增强安全性。

## API响应格式

所有API返回格式统一如下:

```json
{
  "success": true|false,        // 请求是否成功
  "data|connectionInfo": {...}, // 成功时的数据
  "error": "错误信息"            // 失败时的错误信息
}
```

## 数据源API

### 1. CSV文件处理

CSV相关API提供了多种处理CSV数据的方式，从本地文件读取、文件上传到导入数据库的完整流程。

#### 1.1 处理本地CSV文件

**功能说明**: 处理**已存在于服务器本地文件系统**上的CSV文件，前端只需提供服务器上的文件路径。服务器读取并解析该CSV文件，返回数据预览。


```
POST /api/datasource/csv/process
Content-Type: application/json

请求体:
{
  "filePath": "E:/path/to/your/file.csv",  // 服务器本地文件路径
  "options": {
    "delimiter": ",",                       // 分隔符，默认为逗号
    "hasHeader": true,                      // 是否有表头行
    "encoding": "utf-8"                     // 文件编码
  }
}

响应:
{
  "success": true,
  "data": {
    "totalRows": 1000,                      // 总行数
    "columns": ["id", "name", "age", "email"], // 列名列表
    "previewData": [                        // 数据预览，默认前10行
      {"id": "1", "name": "张三", "age": "30", "email": "zhangsan@example.com"},
      // 更多预览数据...
    ]
  }
}
```

#### 1.2 获取CSV列类型

**功能说明**: 分析CSV文件中每列的数据类型，以便进行更准确的数据处理。

```
POST /api/datasource/csv/column-types
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

#### 1.3 上传CSV文件

**功能说明**: 允许用户从**客户端上传CSV文件到服务器**。文件将保存在服务器的指定目录中，便于后续处理。

**使用场景**: 当用户需要将本地CSV文件上传到服务器以进行分析或导入数据库时使用。

```
POST /api/datasource/csv/upload
Content-Type: multipart/form-data

表单字段:
file: [CSV文件]            // 上传的文件对象，通过表单提交
delimiter: ,              // 可选，CSV分隔符
hasHeader: true           // 可选，是否有表头
importToMongo: false      // 可选，是否自动导入到MongoDB，默认false
dbName: ""                // 可选，MongoDB数据库名，仅importToMongo=true时有效
collName: ""              // 可选，MongoDB集合名，仅importToMongo=true时有效

响应（当importToMongo=false时）:
{
  "success": true,
  "filePath": "E:/uploaded/files/data.csv", // 服务器上保存的文件路径
  "fileSize": 1024,                         // 文件大小(字节)
  "message": "文件上传成功"
}

响应（当importToMongo=true时）: 
// 返回与1.4 API相同的MongoDB连接信息
{
  "host": "localhost",                     // MongoDB主机地址
  "port": 27017,                           // MongoDB端口
  "username": "",
  "password": "",
  "database": "csv_data",                  // 使用的数据库名
  "collections": {
    "customers": {                         // 集合名
      "fields": {                          // 字段类型信息
        "_id": "ObjectId",                 // MongoDB自动生成的ID
        "id": "int",
        "name": "str",
        "age": "int",
        "email": "str"
      },
      "sample_data": "{\"_id\": ObjectId(\"67e50e0900ce029f7ac66046\"), \"id\": 1, \"name\": \"张三\", \"age\": 30, \"email\": \"zhangsan@example.com\"}"
    }
  }
}
```

**功能增强**: 设置`importToMongo=true`时，API会在上传文件后自动将数据导入到MongoDB，省去了先上传再导入的两步操作。

上传成功后返回的`filePath`可直接用于后续API调用，如处理CSV或导入到MongoDB。

#### 1.4 将CSV导入MongoDB

**功能说明**: 将CSV数据导入到MongoDB数据库中，创建集合并存储数据。系统会自动处理类型转换和数据验证。

**使用场景**: 需要持久化存储CSV数据并支持复杂查询时使用，比直接处理CSV文件更灵活和高效。

**支持两种调用方式**:

1. **通过JSON指定服务器路径**:
```
POST /api/datasource/csv/import-to-mongo
Content-Type: application/json

请求体:
{
  "filePath": "E:/path/to/your/file.csv",   // 服务器上的CSV文件路径
  "options": {
    "delimiter": ",",                        // 分隔符
    "hasHeader": true,                       // 是否有表头
    "encoding": "utf-8"                      // 文件编码
  },
  "dbName": "csv_data",                      // 可选，MongoDB数据库名，默认使用csv_文件名
  "collName": "customers"                    // 可选，MongoDB集合名，默认为"data"
}
```

2. **直接上传文件**:
```
POST /api/datasource/csv/import-to-mongo
Content-Type: multipart/form-data

表单字段:
file: [CSV文件]            // 上传的文件对象
delimiter: ,              // 可选，CSV分隔符，默认为逗号
hasHeader: true           // 可选，是否有表头，默认为true
encoding: utf-8           // 可选，文件编码，默认为utf-8
dbName: csv_data          // 可选，MongoDB数据库名
collName: customers       // 可选，MongoDB集合名
```

**响应**:
```
{
  "host": "localhost",                     // MongoDB主机地址
  "port": 27017,                           // MongoDB端口
  "username": "",
  "password": "",
  "database": "csv_data",                  // 使用的数据库名
  "collections": {
    "customers": {                         // 集合名
      "fields": {                          // 字段类型信息
        "_id": "ObjectId",                 // MongoDB自动生成的ID
        "id": "int",
        "name": "str",
        "age": "int",
        "email": "str"
      },
      "sample_data": "{\"_id\": ObjectId(\"67e50e0900ce029f7ac66046\"), \"id\": 1, \"name\": \"张三\", \"age\": 30, \"email\": \"zhangsan@example.com\"}"
    }
  }
}
```

**功能增强**: 支持直接上传CSV文件并导入MongoDB，无需先调用上传API再调用导入API，简化了操作流程。

**注意**: 导入后的数据将保存在本地MongoDB数据库中，可通过MongoDB连接API直接访问数据。

### 2. MongoDB连接

**功能说明**: 连接到现有的MongoDB数据库，获取集合信息和样本数据。可以连接导入后的CSV数据或其他MongoDB数据源。

**使用场景**: 访问已存在的MongoDB数据库，将其中的数据提供给前端展示或分析。

**支持两种连接格式**:

1. **使用完整连接URI**:
```
POST /api/datasource/mongodb/connect
Content-Type: application/json

请求体:
{
  "ConnectionURI": "mongodb://localhost:27017",  // MongoDB连接URI
  "Database": "database_name"                    // 要连接的数据库名
}
```

2. **使用独立的连接参数**:
```
POST /api/datasource/mongodb/connect
Content-Type: application/json

请求体:
{
  "host": "localhost",       // MongoDB服务器主机
  "port": 27017,             // 端口号，可选，默认27017
  "username": "admin",       // 用户名，可选
  "password": "password",    // 密码，可选
  "database": "database_name" // 数据库名
}
```

**响应**:
```
{
  "host": "localhost",                         // 数据库主机
  "port": 27017,                               // 数据库端口
  "username": "",                              // 用户名(如有)
  "password": "",                              // 密码(返回时为空，保护敏感信息)
  "database": "database_name",                 // 数据库名
  "collections": {                             // 集合信息
    "departments": {                           // 集合名
      "fields": {                              // 字段类型信息
        "_id": "ObjectId",                     // MongoDB ID字段
        "名字": "str",                         // 字符串类型字段
        "部门": "str"                          // 字符串类型字段
      },
      "sample_data": "{\"_id\": ObjectId(\"67e50e0900ce029f7ac66046\"), \"名字\": \"孙七\", \"部门\": \"销售部\"}"
    },
    "attendance": {
      "fields": {
        "_id": "ObjectId",
        "姓名": "str",
        "日期": "str",
        "考勤": "str"
      },
      "sample_data": "{\"_id\": ObjectId(\"67e50e146add66f28b6746dc\"), \"姓名\": \"张三\", \"日期\": \"2024-03-25\", \"考勤\": \"出勤\"}"
    }
  }
}
```

**特性**:
- 连接URI支持所有标准MongoDB连接字符串参数，包括认证信息、复制集配置等
- 字段名称不区分大小写，例如`ConnectionURI`/`connectionURI`/`connectionuri`都有效
- 端口号支持字符串和数字格式

### 3. MySQL连接

**功能说明**: 连接到MySQL数据库，获取数据库中所有表的结构和样本数据。

**使用场景**: 访问现有MySQL数据库中的数据，无需导出再导入即可在系统中使用。

```
POST /api/datasource/mysql/connect
Content-Type: application/json

请求体:
{
  "host": "localhost",         // MySQL主机地址
  "port": 3306,                // MySQL端口，默认3306
  "username": "dbuser",        // 数据库用户名
  "password": "dbpassword",    // 数据库密码
  "database": "mydatabase"     // 数据库名
}

响应:
{
  "host": "tarsgo.com",
  "port": 3306,
  "username": "tarsgo",
  "password": "",            // 返回时密码为空，保护敏感信息
  "database": "tarsgo",
  "tables": {
    "Article": {
      "fields": {
        "ID": "int(11) unsigned",
        "title": "longtext",
        "time": "text",
        "classification": "text",
        "nickname": "text",
        "content": "longtext",
        "images": "longtext"
      },
      "sample_data": "{\"ID\": 19, \"title\": \"南岭杏花节\", \"time\": \"2023-04-16\", \"classification\": \"活动\", \"nickname\": \"何佳悦 邢浩泽 赵天培 周昊燃 杨好 宫硕 龚博文 李海齐 吴子豪 陈力进 孙健皓 穆子圣 章紫嫣 ...\", \"content\": \"南岭杏花节活动总结\\n                                负责人:何佳悦\\n...\", \"images\": \"[{\\\"name\\\":\\\"83741112_582256a667c2fe201f616a090b3d2d5...\"}"
    },
    "Data": {
      "fields": {
        "ID": "bigint(20)",
        "nickname": "varchar(255)",
        "IDcard": "longtext",
        "sex": "varchar(255)",
        "age": "text",
        "address": "text",
        "classification": "text",
        "school": "text",
        "subjects": "text",
        "phone": "varchar(255)",
        "email": "varchar(255)",
        "qq": "varchar(255)",
        "wechat": "varchar(255)",
        "webID": "text",
        "jlugroup": "text",
        "study": "text",
        "identity": "text",
        "state": "text",
        "image1": "longtext",
        "image2": "longtext"
      },
      "sample_data": "{\"ID\": 403, \"nickname\": \"才爽\", \"IDcard\": \"https://tarsgo.xf233.com/TARSGO/person/才爽.jpeg\", \"sex\": \"女\", \"age\": \"2020\", \"address\": \"吉林省长春市\", \"classification\": \"40200106\", \"school\": \"前卫南区\", \"subjects\": \"人工智能\", \"phone\": \"13039134133\", \"email\": \"2376749633@qq.com\", \"qq\": \"2376749633\", \"wechat\": \"13039134133\", \"webID\": \"155\", \"jlugroup\": \"视觉组\", \"study\": \"哨兵\", \"identity\": \"正式队员\", \"state\": \"是\", \"image1\": \"\", \"image2\": \"\"}"
    }
    // 更多表...
  }
}
```

**注意**: API将返回数据库中所有表的结构和样本数据，便于前端全面了解数据库信息。

### 4. SQLite数据源

#### 4.1 处理SQLite文件

**功能说明**: 读取服务器上的SQLite数据库文件，获取其中的表结构和数据。

**使用场景**: 当有现成的SQLite数据库文件需要查看或导入时使用。

```
POST /api/datasource/sqlite/process
Content-Type: application/json

请求体:
{
  "filePath": "E:/path/to/your/database.db",  // 服务器上的SQLite文件路径
  "table": "users"                            // 可选，指定要查看的表，不提供则返回所有表信息
}

响应:
{
  "success": true,
  "data": {
    "filePath": "E:/path/to/your/database.db",
    "tables": {
      "users": {                              // 表名
        "fields": {                           // 字段信息
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

**注意**: 支持.db、.sqlite和.sqlite3格式的SQLite数据库文件。

#### 4.2 导入SQLite数据到MongoDB

**功能说明**: 将SQLite数据库中的表数据导入到MongoDB中，实现不同数据库之间的迁移。

**使用场景**: 需要将SQLite数据迁移到MongoDB，以便利用MongoDB的高级特性和查询能力。

```
POST /api/datasource/sqlite/import-to-mongo
Content-Type: application/json

请求体:
{
  "filePath": "E:/path/to/your/database.db",    // SQLite文件路径
  "table": "users",                             // 要导入的SQLite表名
  "mongoUri": "mongodb://localhost:27017",      // 可选，MongoDB连接URI
  "dbName": "sqlite_database",                  // 可选，MongoDB数据库名
  "collName": "users"                           // 可选，MongoDB集合名
}

响应:
{
  "success": true,
  "connectionInfo": {
    "host": "localhost",
    "port": 27017,
    "username": "",
    "password": "",
    "database": "sqlite_database",
    "collections": {
      "users": {
        "fields": {
          "_id": "ObjectId",                    // MongoDB自动生成的唯一ID
          "id": "int",                          // 原SQLite表中的ID
          "name": "str",
          "email": "str",
          "created_at": "date"
        },
        "sample_data": "{\"_id\":\"67f7c9e79f2d90b8bcdfa47c\",\"id\":1,\"name\":\"张三\",\"email\":\"zhangsan@example.com\",\"created_at\":\"2024-03-26T10:30:00Z\"}"
      }
    }
  },
  "message": "SQLite数据已成功导入到MongoDB"
}
```

**注意**:
- 如未指定dbName，默认使用"sqlite_文件名"作为数据库名
- 如未指定collName，默认使用表名作为集合名
- 导入完成后，数据存储在本地MongoDB服务中，可通过MongoDB连接API访问

## 数据类型映射

所有数据源API统一使用以下数据类型表示:

| 原始类型 | API返回类型 | 说明 |
|---------|------------|------|
| 整数类型 | `int` | 包括int, integer, bigint等 |
| 浮点类型 | `float` | 包括float, double, decimal等 |
| 字符串类型 | `str` | 包括varchar, text, char等 |
| 布尔类型 | `bool` | 包括boolean, tinyint(1)等 |
| 日期时间类型 | `date` | 包括date, datetime, timestamp等 |
| 二进制类型 | `binary` | 包括blob, binary等 |
| MongoDB ObjectId | `ObjectId` | MongoDB的唯一标识符 |
| 未知类型 | `unknown` | 无法识别的类型 |

## 错误处理

所有API在遇到错误时会返回相应的HTTP状态码和错误信息:

```json
{
  "success": false,
  "error": "错误信息描述"
}
```

## 重要说明

### config.json文件生成

所有数据源连接API（包括MongoDB连接、MySQL连接、SQLite处理和CSV导入）都会在应用程序当前目录的`data`子目录下自动生成`config.json`文件。该文件包含最近一次成功连接的数据源信息，可用于：

1. 快速重连到上次使用的数据源
2. 在前端界面中显示连接信息
3. 作为不同数据源之间的配置传递媒介

文件路径：`./data/config.json`

示例内容（MongoDB连接）：
```json
{
  "host": "localhost",
  "port": 27017,
  "username": "",
  "password": "",
  "database": "test_db",
  "collections": {
    "users": {
      "fields": {
        "_id": "ObjectId",
        "name": "str",
        "age": "int"
      },
      "sample_data": "{\"_id\": \"60d21b4667d0d8992e610c85\", \"name\": \"张三\", \"age\": 30}"
    }
  }
}
```

示例内容（MySQL连接）：
```json
{
  "host": "localhost",
  "port": 3306,
  "username": "root",
  "password": "",
  "database": "test_db",
  "tables": {
    "users": {
      "fields": {
        "id": "int",
        "name": "str",
        "age": "int"
      },
      "sample_data": "{\"id\": 1, \"name\": \"张三\", \"age\": 30}"
    }
  }
}
```