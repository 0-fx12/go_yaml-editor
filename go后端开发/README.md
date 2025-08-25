# VNFs Config Manager (Gin + GORM + MongoDB)

一个支持双数据库存储的VNF配置管理系统，能够智能解析YAML文档并提取表单项。

## 功能特性

- 🚀 **双数据库支持**: 同时使用MySQL和MongoDB存储数据
- 📄 **智能YAML解析**: 自动识别和提取YAML中的表单项配置
- 🔍 **动态字段映射**: 支持不同格式的YAML配置项
- 📊 **灵活存储策略**: MySQL存储结构化数据，MongoDB存储完整配置
- 🎯 **表单项管理**: 自动生成和管理配置表单
- 🔄 **数据同步**: 在MySQL和MongoDB之间保持数据一致性

## 系统要求

- Go 1.22+
- MySQL 8+
- MongoDB 6+

## 快速开始

### 1. 环境准备

```bash
# 创建MySQL数据库
mysql -u root -p -e "CREATE DATABASE vnf_config CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"

# 启动MongoDB服务
mongod --dbpath /data/db
```

### 2. 配置环境变量

复制 `.env.example` 到 `.env` 并调整配置：

```bash
# 应用配置
APP_ENV=development
APP_PORT=8080

# MySQL配置
MYSQL_DSN=root:password@tcp(127.0.0.1:3306)/vnf_config?charset=utf8mb4&parseTime=True&loc=Local

# MongoDB配置
MONGO_URI=mongodb://localhost:27017
MONGO_DATABASE=vnf_config
```

### 3. 安装依赖并运行

```bash
go mod tidy
go run cmd/server/main.go
```

## API端点

### 上传管理
- `POST /api/v1/uploads` - 上传包含YAML的ZIP文件
- `GET /api/v1/vnfs/:id/form-fields` - 获取表单项
- `GET /api/v1/vnfs/:id/yaml-config` - 获取YAML配置

### VNF实例管理
- `GET /api/v1/vnfs` - 列出VNF实例（分页）
- `GET /api/v1/vnfs/:id` - 获取VNF实例详情
- `DELETE /api/v1/vnfs/:id` - 删除VNF实例

### VNF定义管理
- `GET /api/v1/vnfs/:id/definitions` - 列出参数定义（分页，支持修改过滤）
- `POST /api/v1/vnfs/:id/definitions` - 创建参数
- `PUT /api/v1/vnfs/:id/definitions/:defId` - 更新参数
- `DELETE /api/v1/vnfs/:id/definitions/:defId` - 删除参数

## YAML解析特性

### 支持的字段类型
- **基础类型**: string, number, boolean
- **复杂类型**: array, object
- **自动类型推断**: 根据默认值自动确定字段类型

### 表单项属性
- `type` - 字段类型
- `default` - 默认值
- `description` - 描述信息
- `required` - 是否必需
- `hidden` - 是否隐藏
- `validation` - 验证规则
- `options` - 选项列表
- `group` - 分组信息
- `order` - 排序顺序

### 智能解析
- 自动识别配置节点和表单项
- 支持嵌套结构解析
- 智能字段映射和类型推断
- 配置验证和错误提示

## 数据库架构

### MySQL (结构化数据)
- `vnf_instances` - VNF实例基本信息
- `vnf_definitions` - VNF参数定义

### MongoDB (完整配置)
- `vnf_instances` - 完整的VNF实例配置
- `vnf_definitions` - 详细的参数定义
- 存储原始YAML配置和表单项数据

## 前端界面

访问 `http://localhost:8080` 使用简单的Web界面。

## 开发说明

### 项目结构
```
├── cmd/server/          # 主程序入口
├── internal/            # 内部包
│   ├── controller/      # 控制器层
│   ├── service/         # 业务逻辑层
│   ├── model/           # 数据模型
│   ├── dto/             # 数据传输对象
│   ├── infra/db/        # 数据库基础设施
│   └── router/          # 路由配置
├── web/                 # 前端静态文件
└── data/                # 数据目录
    ├── uploads/         # 上传文件
    └── extracts/        # 解压文件
```

### 核心服务
- `YAMLParserService` - YAML解析和表单项提取
- `DualStorageService` - 双数据库存储管理
- `UploadService` - 文件上传和处理

## 部署说明

### Docker部署
```bash
# 构建镜像
docker build -t vnf-config .

# 运行容器
docker run -p 8080:8080 --env-file .env vnf-config
```

### 生产环境
- 设置 `APP_ENV=production`
- 配置适当的数据库连接池大小
- 启用HTTPS和身份验证
- 配置日志记录和监控

## 故障排除

### 常见问题
1. **数据库连接失败**: 检查数据库服务状态和连接字符串
2. **YAML解析错误**: 验证YAML文件格式和语法
3. **存储失败**: 检查数据库权限和磁盘空间

### 日志查看
```bash
# 查看应用日志
tail -f logs/app.log

# 查看数据库日志
tail -f logs/db.log
```

## 贡献指南

1. Fork 项目
2. 创建功能分支
3. 提交更改
4. 推送到分支
5. 创建 Pull Request

## 许可证

MIT License


