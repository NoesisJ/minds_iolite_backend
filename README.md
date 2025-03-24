# Minds Iolite Backend

Minds Iolite后端是一个基于Gin和GORM的Go语言Web服务，为财务管理系统提供API支持。

## 技术栈

- **Go**: 1.21+
- **Web框架**: Gin
- **ORM**: GORM
- **数据库**: MySQL

## 项目结构 
minds_iolite_backend/
├── cmd/ # 应用程序入口
│ └── main.go # 主程序
├── pkg/ # 包目录
│ ├── config/ # 配置相关
│ ├── database/ # 数据库连接及操作
│ ├── handlers/ # HTTP处理器
│ └── models/ # 数据模型
├── .env # 环境变量配置
├── go.mod # Go模块定义
└── go.sum # 依赖版本锁定
```

## API路由

后端提供以下API路由：

### 健康检查API

- `GET /health`: 检查服务器状态

### 人员管据API

- `GET /api/data`: 获取所有人员数据

### 财务数据API

- `GET /api/financial`: 获取所有财务记录
- `GET /api/financial/:id`: 获取单个财务记录
- `POST /api/financial`: 创建新的财务记录
- `PUT /api/financial/:id`: 更新财务记录
- `DELETE /api/financial/:id`: 删除单个财务记录
- `DELETE /api/financial/batch`: 批量删除财务记录
- `GET /api/financial/export`: 导出财务数据为CSV
- `GET /api/health`: 检查API健康状态

## 数据模型

### Financial (财务模型)

财务数据存储在`Financial_Log`表中，主要字段包括：

| 字段名 | 类型 | 描述 |
|-------|------|------|
| ID | uint | 主键 |
| Name | string | 物资名称 |
| Model | string | 型号 |
| Quantity | string | 数量 |
| Unit | string | 单位 |
| Price | string | 价格 |
| ExtraPrice | string | 额外费用(如运费) |
| PurchaseLink | string | 购买链接 |
| PostDate | string | 发布日期 |
| Purchaser | string | 购买人 |
| Campus | string | 校区 |
| GroupName | string | 组名 |
| TroopTypeProject | string | 兵种和项目 |
| Remarks | string | 备注 |

### Data (人员模型)

人员数据存储在`Data`表中，主要字段包括用户信息、联系方式等。

## 如何运行

1. 确保已安装Go 1.21+
2. 配置数据库连接（在`.env`文件中）
3. 运行以下命令：

```bash
# 进入项目目录
cd minds_iolite_backend

# 运行项目
go run cmd/main.go
```

默认情况下，服务器将在`:8080`端口启动。

## 跨域支持

后端已配置CORS中间件，支持跨域请求，允许来自任何源的API访问。

## 错误处理

系统使用HTTP状态码和统一的错误响应格式处理错误：

- 400: 请求参数错误
- 404: 资源未找到
- 500: 服务器内部错误

## 开发指南

1. **添加新API**:
   - 在`pkg/handlers/`中创建新处理函数
   - 在`pkg/handlers/routes.go`中注册路由

2. **修改数据模型**:
   - 在`pkg/models/`中更新模型定义
   - 根据需要更新数据库迁移逻辑

3. **环境配置**:
   - 开发环境配置在`.env`文件中
   - 生产环境建议使用环境变量

## API请求/响应格式

### 标准响应格式

成功响应：
```json:minds_iolite_backend/README.md
{
  "success": true,
  "error": null,
  "data": {...}
}
```

错误响应：
```json
{
  "success": false,
  "error": "错误描述",
  "data": null
}
```

### 财务记录创建示例

请求：
```
POST /api/financial
Content-Type: application/json

{
  "name": "测试物资",
  "model": "test-001",
  "quantity": "5",
  "unit": "个",
  "price": "100.00",
  "extra_price": "10.00",
  "purchase_link": "http://example.com",
  "post_date": "2023-07-15",
  "purchaser": "张三",
  "campus": "前卫南区",
  "group_name": "机械组",
  "troop_type": "步兵",
  "project": "测试项目",
  "remarks": "这是一条测试记录"
}
```

响应：
```json
{
  "success": true,
  "error": null,
  "data": {
    "id": 1,
    "name": "测试物资",
    "model": "test-001",
    "quantity": "5",
    "unit": "个",
    "price": "100.00",
    "extra_price": "10.00",
    "purchase_link": "http://example.com",
    "post_date": "2023-07-15",
    "purchaser": "张三",
    "campus": "前卫南区",
    "group_name": "机械组",
    "troop_type": "步兵",
    "project": "测试项目",
    "remarks": "这是一条测试记录"
  }
}
```