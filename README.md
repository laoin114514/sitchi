# Sitchi 开发规范

本文档用于统一后端（Go）与前端（WebUI）开发方式，降低沟通成本，保证代码可维护性。

## 1. 项目结构约定

### 1.1 后端（Go）

- `cmd/manage/`
  - 启动与管理命令入口（迁移、初始化、启动服务）
- `configs/`
  - 配置加载、数据库连接、迁移脚本
- `internal/`
  - 业务代码（按模块拆分）
  - `module/`：模块管理
  - `user/`：用户管理与登录
  - `middleware/`：中间件
  - `common/`：公共模型、JWT 等

### 1.2 前端（webui）

- `src/components/` 公共组件
- `src/views/` 页面组件
- `src/store/` Pinia 状态
- `src/router/` 路由
- `src/utils/` 工具函数（如请求封装）
- `src/api/` 接口定义与调用
- `src/styles/` 全局样式与主题变量

---

## 2. 分层职责规范（后端）

### 2.0 统一代码组织风格

- 除了简单 model/DTO，`controller/service/repository/dao` 全部使用结构体封装。
- 统一提供 `NewXxx(...)` 构造函数进行依赖注入。
- 禁止新增包级“业务入口函数”作为主调用方式（兼容旧代码的 wrapper 可保留但不再扩展）。

推荐形态：

```go
type Service struct {
    db   *sql.DB
    repo *Repository
}

func NewService(db *sql.DB, repo *Repository) *Service { ... }
func (s *Service) CreateXxx(...) (...) { ... }
```

### 2.1 controller
- 只处理 HTTP 入参/出参与状态码
- 不写复杂业务逻辑
- 统一返回 `internal/common/model/api_response.go`
- 使用结构体封装（如 `type Controller struct { service *Service }`）
- 路由注册使用实例方法（如 `userController.Login`），避免直接绑定包级函数

### 2.2 service
- 负责业务编排、事务边界、规则校验
- 只调用 repository/dao，不直接写 SQL
- 使用结构体封装（`type Service struct {...}`）
- 事务应在 service 层开启与提交，repository/dao 不自行开启事务

### 2.3 repository
- 只负责模块内数据库访问与 SQL
- 使用结构体封装（`type Repository struct {...}`）
- 方法命名清晰：`Get/Insert/Update/Delete/Rebuild...`
- 尽量参数化查询，禁止字符串拼接 SQL

### 2.4 dao
- 放“跨模块复用”的通用数据库操作，不放模块专属复杂业务 SQL
- 使用结构体封装（如 `type ACLDAO struct{}`）
- 典型场景：
  - 通用编码构造（`module_code:role_code`）
  - 通用查询（`GetModuleIDByCode`）
  - 通用关系写入（`BindUserRole`）
  - 通用缓存回填（`RebuildUserPermResByUser`）
- 原则：模块专属逻辑留在模块 repository，通用能力再下沉到 dao

---

## 3. 数据与权限模型规范

### 3.1 命名规范
- 角色编码：`module_code:role_code`
  - 例：`admin:admin`、`order:normal`、`admin:super`
- 资源编码：`module_code:resource_code`
  - 例：`admin:users`、`order:module`
- 权限编码（模块内唯一）：`view/plus/change/delete/bind/unbind`

### 3.2 鉴权规则
- 运行时鉴权统一查 `user_perm_res`（最终权限缓存）
- 禁止在高频鉴权中使用字符串模糊匹配（LIKE）
- 允许在管理层将 code 解析为 id，再走 id 查询

### 3.3 角色规则
- 每个模块默认创建：`<module_code>:admin`、`<module_code>:normal`
- `super` 角色用于全局放行场景（由初始化流程创建）

---

## 4. JWT 规范

- 使用 `internal/common/jwt_auth_service.go`
- Token 类型：
  - `access`：接口鉴权
  - `refresh`：刷新令牌
- 鉴权中间件仅验证 access token
- refresh token 仅用于刷新接口，不进入常规业务中间件

---

## 5. 错误与响应规范（重点）

### 5.1 统一错误模型（AppError）

- 统一使用：`internal/common/model/app_error.go`
- 错误结构字段：
  - `Code`：业务错误码（如 `401001`）
  - `Type`：错误类型（如 `UNAUTHORIZED`）
  - `Message`：默认错误文案（建议英文）
  - `I18nKey`：前端国际化键（如 `error.unauthorized`）
  - `Detail`：调试细节（生产环境谨慎返回）

#### 错误码分段约定

- `400xxx`：参数/请求错误
- `401xxx`：未认证
- `403xxx`：无权限
- `404xxx`：资源不存在
- `409xxx`：资源冲突
- `500xxx`：系统内部错误

#### HTTP 状态码映射

- 必须通过 `AppError.HTTPStatus()` 或 `HTTPStatusFromError(err)` 统一映射
- 禁止在各 controller 中手写零散状态码判断

### 5.2 统一 API 返回模型（ApiResponse）

- 统一使用：`internal/common/model/api_response.go`
- 标准返回结构：
  - `success`：是否成功
  - `code`：业务码
  - `message`：说明
  - `data`：业务数据（成功时）
  - `error`：错误详情（失败时）
  - `timestamp`：毫秒时间戳

#### 成功返回

- 统一调用：`ApiSuccessResponse(data)`
- 业务成功时，HTTP 状态码通常为 `200`

#### 失败返回

- 统一调用：`ApiErrorResponse(code, message, err)`
- HTTP 状态码使用 `appErr.HTTPStatus()` 计算，不手写 magic number

### 5.3 Controller 错误处理范式

```go
if err != nil {
    appErr := model.ErrInternal.WithDetail(err.Error())
    c.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
    return
}

c.JSON(http.StatusOK, model.ApiSuccessResponse(data))
```

### 5.4 Service 与 Repository 错误边界

- repository：返回底层错误（`sql.ErrNoRows` 等）
- service：做错误语义转换（映射为业务错误）
- controller：只做最终输出封装（`ApiResponse` + HTTP 状态）

### 5.5 安全要求

- 禁止将密码、密钥、SQL 原文直接放入 `Detail` 返回给前端
- 生产环境可通过开关裁剪 `Detail`

---

## 6. 前端开发规范（Vue3 + TS + Element Plus）

1. 必须使用 `<script setup lang="ts">`
2. 命名：变量/函数 `camelCase`，组件 `PascalCase`
3. 组件使用 `defineProps` + TS 类型、`defineEmits`
4. 样式默认 `scoped`
5. 颜色禁止硬编码，统一走 `src/styles/theme.scss` 变量
6. 请求统一经过 `src/utils/request.ts`（axios）
7. 页面逻辑中不直接写裸 `fetch`

---

## 7. 主题规范

- 使用 `data-theme` 机制切换主题
- 深色变量放 `:root`
- 浅色变量放 `[data-theme='light']`
- 新增颜色时必须先在主题变量中定义，再在组件内使用

---

## 8. 开发流程规范

### 8.1 本地开发推荐顺序
1. 执行数据库迁移
2. 初始化系统（admin 模块 + admin 用户）
3. 启动后端服务
4. 启动 webui

### 8.2 后端常用命令

```bash
go run ./cmd/manage migrate
go run ./cmd/manage initSystem [admin_password]
go run ./cmd/manage runServer
```

### 8.3 前端常用命令

```bash
cd webui
npm install
npm run dev
```

---

## 9. 代码评审（CR）检查项

- 是否遵守分层职责（controller/service/repository）
- 是否复用统一错误与响应模型
- 是否使用事务保护关键写操作
- 是否避免循环依赖
- 是否遵守角色/资源编码规范
- 是否将颜色抽为主题变量
- 是否在新增接口时补齐鉴权与错误处理

---

## 10. 禁止事项

- controller 直接写 SQL
- service 中拼接 SQL
- 鉴权逻辑散落在业务代码中（应集中中间件/权限层）
- 前端组件硬编码主题颜色
- 绕过 `request.ts` 直接在页面写请求

---

## 11. 后续建议（可选）

- 增加 `auth/refresh` 接口，完整利用 refresh token
- 增加统一 `request_id` 中间件（前后端链路追踪）
- 增加权限变更审计日志
- 补充 E2E：登录、鉴权、角色切换场景

---

若团队后续新增规范，请直接更新本文件并在 PR 中说明“规范变更点”。
