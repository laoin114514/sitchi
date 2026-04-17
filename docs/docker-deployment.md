# Sitchi Docker 部署文档

## 目录

1. [设计思路](#设计思路)
2. [前后端合并部署原理](#前后端合并部署原理)
3. [项目结构](#项目结构)
4. [使用方法](#使用方法)
5. [配置说明](#配置说明)
6. [常见问题](#常见问题)

---

## 设计思路

### 1. 单容器部署

本项目采用**单容器部署**方案，将后端 API 服务和前端 WebUI 打包到同一个 Docker 镜像中：

- **简化部署**：只需管理一个容器，降低运维复杂度
- **端口统一**：只暴露一个端口（5050），同时提供 API 和页面服务
- **环境一致**：开发、测试、生产环境使用相同的部署方式
- **配置集中**：通过 `configs/.env` 统一管理环境配置

### 2. 多阶段构建

Dockerfile 采用**多阶段构建**策略，优化镜像体积和构建效率：

```
阶段1: backend-builder (golang:1.25.0-alpine)
   └── 编译 Go 后端程序

阶段2: frontend-builder (node:22-alpine)
   └── 构建 Vue 前端项目

阶段3: runtime (alpine:latest)
   └── 合并前后端产物，运行服务
```

**构建优化点**：
- 前端构建时先复制 `package.json`，利用 Docker 缓存机制
- 只要依赖不变，`npm install` 会直接使用缓存层
- 最终镜像只包含运行所需的二进制文件和静态资源

### 3. 环境隔离

通过 `configs/.env` 文件区分不同环境：

```bash
MODE=dev      # 环境标识：dev/prod/自定义
PORT=5050     # 服务暴露端口
```

- **开发环境**：热重载、详细日志、本地数据库
- **生产环境**：性能优化、错误监控、生产数据库

---

## 前后端合并部署原理

### 1. 路由分发机制

后端使用 Gin 框架提供统一入口，根据请求路径智能分发：

```
用户请求 localhost:5050
        │
        ▼
┌─────────────────┐
│   Gin Router    │
└─────────────────┘
        │
    ┌───┴───┐
    ▼       ▼
┌──────┐  ┌──────────┐
│ /api │  │ 其他路径  │
│ 路由 │  │ (NoRoute)│
└──┬───┘  └────┬─────┘
   │           │
   ▼           ▼
后端处理    返回 index.html
(API逻辑)   (Vue Router接管)
```

### 2. 代码实现

**后端路由配置** (`internal/router.go`):

```go
func InitRouter(r *gin.Engine) {
    // 1. API 路由 - 后端接口
    webuiApiGroup := r.Group("/api/webui")
    {
        webuiApiGroup.POST("/login", userController.Login)
    }
    
    // 2. 静态文件服务 - 前端资源
    r.Static("/assets", "./webui/dist/assets")
    
    // 3. 前端路由支持 - 所有非API请求返回index.html
    r.NoRoute(func(c *gin.Context) {
        if strings.HasPrefix(c.Request.URL.Path, "/api") {
            c.JSON(404, gin.H{"error": "API not found"})
            return
        }
        c.File("./webui/dist/index.html")
    })
}
```

### 3. 请求处理流程

| 请求路径 | 处理方式 | 响应内容 |
|---------|---------|---------|
| `GET /health` | 后端路由 | `{"status": "ok"}` |
| `POST /api/webui/login` | 后端路由 | 登录验证逻辑 |
| `GET /assets/*` | 静态文件 | JS/CSS/图片资源 |
| `GET /login` | NoRoute | `index.html` |
| `GET /` | NoRoute | `index.html` |

**SPA 路由支持**：
- 前端使用 Vue Router 进行客户端路由
- 后端将所有非API请求指向 `index.html`
- 浏览器加载页面后，Vue Router 根据路径渲染对应组件

---

## 项目结构

```
sitchi/
├── cmd/manage/           # Go 后端入口
├── internal/             # 后端业务逻辑
│   ├── router.go        # 路由配置（含静态文件服务）
│   └── ...
├── webui/               # Vue 前端项目
│   ├── src/
│   ├── dist/            # 构建产物（Docker中生成）
│   └── package.json
├── configs/             # 配置文件
│   ├── .env            # 环境变量（MODE, PORT）
│   ├── configs.dev.yml # 开发环境配置
│   └── configs.prod.yml # 生产环境配置
├── script/
│   └── docker.sh       # Docker 管理脚本
├── Dockerfile          # 多阶段构建配置
├── docker-compose.yml  # 容器编排配置
└── docs/
    └── docker-deployment.md  # 本文档
```

---

## 使用方法

### 1. 环境准备

确保已安装：
- Docker Engine 20.10+
- Docker Compose 2.0+
- Git Bash / Linux / macOS 终端

### 2. 配置环境

编辑 `configs/.env` 文件：

```bash
# 开发环境
MODE=dev
PORT=5050

# 或生产环境
MODE=prod
PORT=5050
```

### 3. 使用管理脚本

项目提供交互式管理脚本 `script/docker.sh`：

```bash
# 进入交互式菜单
./script/docker.sh
```

**菜单选项**：

```
╔══════════════════════════════════════════════════════════╗
║           Sitchi Docker 管理脚本                       ║
╚══════════════════════════════════════════════════════════╝

当前环境配置 (从 configs/.env 读取):
  环境(MODE): dev
  项目名:     sitchi-dev
  端口(PORT): 5050

请选择操作:

  1) build    - 构建镜像
  2) start    - 启动容器
  3) stop     - 停止容器
  4) restart  - 重启容器
  5) status   - 查看容器状态
  6) logs     - 查看容器日志
  7) clean    - 清理容器和镜像
  8) shell    - 进入容器内部
  9) 刷新配置 - 重新读取.env文件
  0) 退出
```

### 4. 命令行模式

```bash
# 构建镜像
./script/docker.sh build

# 启动服务
./script/docker.sh start

# 查看日志
./script/docker.sh logs

# 停止服务
./script/docker.sh stop

# 查看帮助
./script/docker.sh --help
```

### 5. 访问服务

启动成功后访问：

- **前端页面**: http://localhost:5050
- **登录页面**: http://localhost:5050/login
- **健康检查**: http://localhost:5050/health
- **API 接口**: http://localhost:5050/api/webui/login

---

## 配置说明

### 环境变量 (`configs/.env`)

| 变量 | 说明 | 示例 |
|-----|------|------|
| `MODE` | 运行模式 | `dev`, `prod` |
| `PORT` | 服务端口 | `5050` |

### 配置文件 (`configs/configs.{MODE}.yml`)

```yaml
server:
  port: 5051          # 后端监听端口（需与Dockerfile一致）
  host: 0.0.0.0       # 监听地址
  allow_origins:
    - "*"             # CORS配置

db:
  host: 172.28.219.70 # 数据库地址
  port: 5432          # 数据库端口
  user: postgres      # 数据库用户
  password: xxx       # 数据库密码
  dbname: sitchi      # 数据库名
  sslmode: disable    # SSL模式

auth:
  jwt_secret: "xxx"   # JWT密钥
  access_token_ttl_minutes: 60
  refresh_token_ttl_minutes: 10080

dev: true             # 开发模式标识
```

### 端口映射关系

```
宿主机端口 (PORT)  →  容器端口 (5050)  →  后端服务 (configs.yml中的port)
     5050                5050                  5051
```

**注意**：
- `docker-compose.yml` 中的 `PORT` 是宿主机映射端口
- `Dockerfile` 暴露的 `5050` 是容器内部端口
- `configs.yml` 中的 `port` 是后端实际监听端口（需与Dockerfile一致）

---

## 常见问题

### Q1: 前端代码修改后如何更新？

需要重新构建镜像：

```bash
./script/docker.sh build
```

或完整重建：

```bash
./script/docker.sh clean
./script/docker.sh build
```

### Q2: 如何查看容器内部文件？

```bash
./script/docker.sh shell
# 进入容器后
ls -la /app/webui/dist  # 查看前端文件
ls -la /app/configs     # 查看配置文件
```

### Q3: 如何切换环境？

1. 修改 `configs/.env` 文件中的 `MODE`
2. 运行 `./script/docker.sh`，选择 `9) 刷新配置`
3. 或直接重启：`./script/docker.sh restart`

### Q4: 构建很慢怎么办？

- 首次构建需要下载依赖，时间较长
- 后续构建会利用 Docker 缓存
- 确保网络连接正常（使用了国内镜像源）

### Q5: 前端页面显示 404？

检查：
1. 容器是否正常运行：`./script/docker.sh status`
2. 前端是否构建成功：查看构建日志
3. 访问路径是否正确：使用 `/` 而非 `/index.html`

---

## 总结

本项目的 Docker 部署方案具有以下特点：

1. **单容器架构**：前后端合并，简化部署
2. **多阶段构建**：优化镜像体积和构建速度
3. **配置驱动**：通过 `.env` 文件灵活切换环境
4. **SPA 支持**：完美支持 Vue Router 前端路由
5. **统一管理**：提供交互式脚本，简化日常操作
