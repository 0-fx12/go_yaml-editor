# YAML配置管理器

一个基于Go + Gin框架的Web应用，用于管理和编辑YAML配置文件，支持前后端一体化部署，提供直观的Web界面进行YAML配置的查看、编辑和历史记录管理。

## 🚀 项目特性

- **📄 YAML文件管理**: 支持读取、解析和编辑YAML配置文件
- **🌐 Web界面**: 提供现代化的响应式Web界面
- **📝 在线编辑**: 支持字段级别的在线编辑，保留原始文件格式
- **📊 分页显示**: 大型配置文件的分页展示
- **🔍 搜索过滤**: 支持按路径和值进行实时搜索
- **💾 历史记录**: MongoDB存储配置变更历史
- **🐳 容器化**: 完整的Docker和Kubernetes支持
- **🔄 健康检查**: 内置服务健康检查和监控
- **🎯 无状态设计**: 前端完全无状态，支持水平扩容

## 🏗️ 技术架构

### 后端技术栈

- **Go 1.22**: 主要编程语言
- **Gin**: Web框架
- **MongoDB**: 数据存储和历史记录
- **YAML**: 配置文件格式支持

### 前端技术栈

- **原生HTML/CSS/JavaScript**: 轻量级前端实现
- **响应式设计**: 支持移动端和桌面端

### 基础设施

- **Docker**: 容器化部署
- **Kubernetes**: 容器编排
- **Nginx Ingress**: 负载均衡和路由

## 📁 项目结构

```
simple_version/
├── main.go.go              # 主应用程序入口
├── mongo/                  # MongoDB操作模块
│   └── mongo.go           # 数据库连接和操作
├── web/                   # 前端静态资源
│   └── index.html         # Web界面
├── Dockerfile             # Docker镜像构建文件
├── docker-compose.yml     # 本地开发环境配置
├── deploy.sh              # Kubernetes部署脚本
├── k8s/                   # K8s配置文件目录
│   ├── k8s-namespace.yaml     # K8s命名空间配置
│   ├── k8s-mongodb.yaml       # MongoDB部署配置
│   ├── k8s-configmap.yaml     # 应用环境配置
│   ├── k8s-app-deployment.yaml # 应用部署配置
│   ├── k8s-ingress.yaml       # Ingress路由配置
│   └── k8s-app-config-files.yaml # 配置文件ConfigMap
├── sample_config.yaml     # 示例配置文件
├── go.mod                 # Go模块依赖
├── go.sum                 # 依赖校验文件
└── README.md              # 项目文档
```

## 🛠️ 快速开始

### 前置要求

- **Go 1.22+**: [下载安装](https://golang.org/dl/)
- **Docker**: [下载安装](https://www.docker.com/get-started)
- **MongoDB**: 本地安装或使用Docker
- **Kubernetes集群** (可选): 用于生产环境部署

### 本地开发环境

#### 1. 克隆项目

```bash
git clone <repository-url>
cd simple_version
```

#### 2. 安装依赖

```bash
go mod download
```

#### 3. 配置环境变量

创建 `.env` 文件：

```env
APP_PORT=8080
MONGO_URI=mongodb://localhost:27017
MONGO_DATABASE=vnf_config
GIN_MODE=debug
```

#### 4. 启动MongoDB

```bash
# 使用Docker启动MongoDB
docker run -d \
  --name mongodb \
  -p 27017:27017 \
  -e MONGO_INITDB_ROOT_USERNAME=admin \
  -e MONGO_INITDB_ROOT_PASSWORD=password123 \
  mongo:7.0
```

#### 5. 运行应用

```bash
go run main.go.go
```

访问 http://localhost:8080 查看Web界面。

### 使用Docker Compose

这是推荐的本地开发方式，会自动启动所有依赖服务：

```bash
# 构建并启动所有服务
docker-compose up --build

# 后台运行
docker-compose up -d --build

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down
```

## 🐳 Docker部署

### 构建镜像

```bash
# 构建应用镜像
docker build -t yaml-config-manager:latest .

# 查看镜像
docker images | grep yaml-config-manager
```

### 运行容器

```bash
# 创建网络
docker network create yaml-config-network

# 启动MongoDB
docker run -d \
  --name mongodb \
  --network yaml-config-network \
  -e MONGO_INITDB_ROOT_USERNAME=admin \
  -e MONGO_INITDB_ROOT_PASSWORD=password123 \
  -p 27017:27017 \
  mongo:7.0

# 启动应用
docker run -d \
  --name yaml-config-app \
  --network yaml-config-network \
  -e MONGO_URI=mongodb://admin:password123@mongodb:27017 \
  -e MONGO_DATABASE=vnf_config \
  -p 8080:8080 \
  yaml-config-manager:latest
```

## ☸️ Kubernetes部署

### 自动化部署

使用提供的部署脚本进行一键部署：

```bash
# 给脚本添加执行权限 (Linux/Mac)
chmod +x deploy.sh

# 构建镜像并部署到K8s
./deploy.sh --build --deploy

# 查看部署状态
./deploy.sh --status

# 清理部署
./deploy.sh --cleanup
```

### 手动部署

#### 1. 创建命名空间

```bash
kubectl apply -f k8s/k8s-namespace.yaml
```

#### 2. 部署MongoDB

```bash
kubectl apply -f k8s/k8s-mongodb.yaml

# 等待MongoDB就绪
kubectl wait --for=condition=ready pod -l app=mongodb -n yaml-config-manager --timeout=300s
```

#### 3. 创建配置

```bash
kubectl apply -f k8s/k8s-configmap.yaml
kubectl apply -f k8s/k8s-app-config-files.yaml
```

#### 4. 部署应用

```bash
kubectl apply -f k8s/k8s-app-deployment.yaml

# 等待应用就绪
kubectl wait --for=condition=ready pod -l app=yaml-config-manager -n yaml-config-manager --timeout=300s
```

#### 5. 配置Ingress

```bash
kubectl apply -f k8s/k8s-ingress.yaml
```

#### 6. 验证部署

```bash
# 查看Pod状态
kubectl get pods -n yaml-config-manager

# 查看服务状态
kubectl get svc -n yaml-config-manager

# 查看应用日志
kubectl logs -l app=yaml-config-manager -n yaml-config-manager
```

### 访问应用

部署完成后，可以通过以下方式访问：

- **本地端口转发**:

  ```bash
  kubectl port-forward service/yaml-config-app 8080:8080 -n yaml-config-manager
  ```

  然后访问 http://localhost:8080
- **通过Ingress** (需要配置域名解析):

  - http://yaml-config.local
  - http://api.yaml-config.local/api/v1/

## 🔧 配置说明

### 环境变量

| 变量名             | 默认值                        | 说明              |
| ------------------ | ----------------------------- | ----------------- |
| `APP_PORT`       | `8080`                      | 应用监听端口      |
| `MONGO_URI`      | `mongodb://localhost:27017` | MongoDB连接字符串 |
| `MONGO_DATABASE` | `vnf_config`                | MongoDB数据库名称 |
| `GIN_MODE`       | `release`                   | Gin框架运行模式   |

### YAML配置文件

应用会按以下优先级查找YAML配置文件：

1. `config.yaml`
2. `sample_config.yaml`
3. `test.yaml`

### 示例配置文件

```yaml
app:
  name: "YAML Config Manager"
  version: "1.0.0"
  description: "A web-based YAML configuration manager"

server:
  port: 8080
  host: "0.0.0.0"
  cors:
    enabled: true
    origins: ["*"]

database:
  type: "mongodb"
  connection_string: "mongodb://admin:password123@mongodb:27017"
  database_name: "vnf_config"

features:
  yaml_editor: true
  config_history: true
  api_access: true

logging:
  level: "info"
  format: "json"
```

## 📚 API文档

### 获取YAML数据 (分页)

```http
GET /api/v1/yaml?page=1&size=20
```

**响应示例:**

```json
{
  "data": {
    "fields": [
      {
        "path": "app.name",
        "value": "YAML Config Manager",
        "type": "string",
        "description": ""
      }
    ],
    "total": 10,
    "page": 1,
    "pageSize": 20,
    "totalPage": 1
  },
  "message": "success"
}
```

### 获取原始YAML内容

```http
GET /api/v1/yaml/raw
```

### 保存YAML修改

```http
POST /api/v1/yaml
Content-Type: application/json

{
  "updates": {
    "app.name": "New App Name",
    "server.port": 9090
  }
}
```

## 🎯 功能特性详解

### 1. YAML文件解析

- **保留原始格式**: 使用Go的yaml.Node保留键的原始顺序
- **类型识别**: 自动识别字符串、数字、布尔值、数组、对象类型
- **路径映射**: 将嵌套结构映射为点分隔的路径格式

### 2. Web界面功能

- **响应式设计**: 自适应桌面端和移动端
- **实时搜索**: 支持按字段路径或值进行实时过滤
- **在线编辑**: 不同数据类型提供对应的编辑控件
- **变更标识**: 高亮显示已修改但未保存的字段
- **批量保存**: 支持多个字段同时保存

### 3. 数据存储

- **历史记录**: 每次读取和修改都会保存到MongoDB
- **最新快照**: 维护每个文件的最新状态快照
- **操作审计**: 记录所有配置变更的时间和内容

### 4. 容器化特性

- **多阶段构建**: 使用Docker多阶段构建优化镜像大小
- **非root运行**: 容器内使用非特权用户运行
- **健康检查**: 内置HTTP健康检查端点
- **优雅关闭**: 支持信号处理和优雅关闭

### 5. Kubernetes集成

- **StatefulSet**: MongoDB使用StatefulSet确保数据持久性
- **ConfigMap**: 配置文件和环境变量通过ConfigMap管理
- **Service**: 提供稳定的服务发现和负载均衡
- **Ingress**: 支持外部访问和路由规则
- **资源限制**: 设置CPU和内存限制防止资源滥用

## 🔍 监控和调试

### 查看应用日志

**Docker Compose:**

```bash
docker-compose logs -f app
```

**Kubernetes:**

```bash
kubectl logs -l app=yaml-config-manager -n yaml-config-manager -f
```

### 健康检查

应用提供以下健康检查端点：

- **HTTP检查**: `GET /api/v1/yaml`
- **容器检查**: `wget --spider http://localhost:8080/api/v1/yaml`

### 数据库连接检查

```bash
# 进入MongoDB容器
kubectl exec -it <mongodb-pod-name> -n yaml-config-manager -- mongosh

# 或使用Docker
docker exec -it mongodb mongosh

# 连接并检查数据
use vnf_config
db.yaml_latest.find()
```

## 🚨 故障排除

### 常见问题

#### 1. 应用启动失败

**症状**: 容器启动后立即退出
**解决方案**:

- 检查MongoDB连接配置
- 确认YAML配置文件存在
- 查看应用日志获取详细错误信息

#### 2. MongoDB连接失败

**症状**: 应用日志显示MongoDB连接错误
**解决方案**:

```bash
# 检查MongoDB服务状态
kubectl get pods -l app=mongodb -n yaml-config-manager

# 检查网络连接
kubectl exec -it <app-pod> -n yaml-config-manager -- nslookup mongodb
```

#### 3. 无法访问Web界面

**症状**: 浏览器无法打开应用页面
**解决方案**:

- 检查Ingress配置和域名解析
- 使用port-forward进行本地访问测试
- 检查Service和Pod的状态

#### 4. YAML文件解析失败

**症状**: Web界面显示"未找到可用的YAML文件"
**解决方案**:

- 确认配置文件已正确挂载到容器
- 检查文件权限和路径
- 验证YAML文件格式是否正确

### 调试命令

```bash
# 查看Pod详细信息
kubectl describe pod <pod-name> -n yaml-config-manager

# 进入应用容器调试
kubectl exec -it <app-pod> -n yaml-config-manager -- /bin/sh

# 检查配置文件挂载
kubectl exec -it <app-pod> -n yaml-config-manager -- ls -la /app/

# 检查环境变量
kubectl exec -it <app-pod> -n yaml-config-manager -- env
```

## 🔧 开发指南

### 本地开发环境设置

1. **设置IDE**: 推荐使用VS Code + Go扩展
2. **代码规范**: 使用gofmt格式化代码
3. **依赖管理**: 使用Go modules管理依赖

### 添加新功能

1. **后端API**: 在 `main.go.go`中添加新的路由处理
2. **前端界面**: 修改 `web/index.html`中的JavaScript
3. **数据库操作**: 在 `mongo/mongo.go`中添加新的数据库操作

### 测试

```bash
# 运行单元测试
go test ./...

# 构建验证
go build -o test-binary main.go.go

# 容器构建测试
docker build -t test-image .
```

## 📦 部署最佳实践

### 生产环境建议

1. **资源规划**: 根据配置文件大小和访问频率调整资源限制
2. **数据备份**: 定期备份MongoDB数据
3. **监控报警**: 集成Prometheus + Grafana监控
4. **日志收集**: 集成ELK或类似的日志系统
5. **安全配置**: 使用HTTPS和访问控制

### 扩容策略

- **水平扩容**: 应用层支持多副本部署
- **数据库**: MongoDB可以配置副本集提高可用性
- **存储**: 使用持久卷存储MongoDB数据

## 🤝 贡献指南

1. Fork本项目
2. 创建功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启Pull Request

## 📄 许可证

本项目采用MIT许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 📞 支持

如果您在使用过程中遇到问题或有建议，请：

1. 查看本文档的故障排除部分
2. 搜索已有的GitHub Issues
3. 创建新的Issue描述问题
4. 联系项目维护者

---

**注意**: 这是一个演示项目，用于学习Go Web开发和Kubernetes部署。在生产环境中使用前，请确保进行充分的安全评估和测试。
