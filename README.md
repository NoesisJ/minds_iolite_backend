# Minds Iolite 

Minds Iolite Backend 是一个基于Golang的低代码平台后端实现，使用Gin作为Web框架，GORM进行数据操作，MongoDB作为数据库。该项目旨在为前端低代码平台提供灵活、高效的后端支持，允许用户通过拖拽方式构建信息管理系统，同时无需编写大量代码即可实现数据库操作。

## 已实现功能

### 核心功能
- ✅ 元数据模型定义和管理
- ✅ 动态API自动生成
- ✅ 数据验证和类型检查
- ✅ 高级查询、排序和分页
- ✅ MongoDB数据存储集成

### API功能
- ✅ 元数据管理API (`/metadata/*`)
- ✅ 动态生成的数据API (`/api/*`)
- ✅ 完整CRUD操作支持
- ✅ 过滤、排序和分页

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

## 后续开发计划
- [ ] 用户认证和权限控制
- [ ] 模型关系和联合查询
- [ ] 自定义业务逻辑挂钩
- [ ] 前端界面开发
- [ ] 数据可视化和报表 