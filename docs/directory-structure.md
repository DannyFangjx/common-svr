# common-svr 目录结构

## 设计目标

采用传统三层结构作为主链路，同时把路由、中间件、基础设施和进程装配职责分离：

```text
Router -> Middleware -> Handler -> Service -> Common/DB
```

`bootstrap` 是额外的装配层，只负责初始化和连接各组件，不承载业务逻辑。

当前 Demo 已实现基础探活接口以及 User CRUD。用户请求依次经过 Router、Middleware、Handler、Service、User Repository，最终访问 PostgreSQL。

## 目标结构

```text
common-svr/
├── main.go
├── go.mod
├── go.sum
├── Makefile
├── Dockerfile
├── .gitignore
├── .env
├── .env.example
│
├── config/
│   ├── app.yaml
│   ├── database.yaml
│   └── rules.yaml
│
├── internal/
│   ├── router/
│   │   ├── router.go
│   │   ├── route_health.go
│   │   ├── route_user.go
│   │   └── middleware/
│   │
│   ├── handler/
│   │   ├── health/
│   │   └── user/
│   │
│   ├── service/
│   │   └── user/
│   │
│   ├── model/
│   │
│   ├── common/
│   │   ├── config/
│   │   ├── db/
│   │   │   ├── postgres/
│   │   │   ├── mongo/
│   │   │   └── user/
│   │   ├── cache/
│   │   ├── mq/
│   │   ├── client/
│   │   ├── logger/
│   │   ├── errors/
│   │   ├── response/
│   │   └── utils/
│   │
│   ├── process/
│   └── bootstrap/
│
├── scripts/
├── tools/
├── test/
│   ├── integration/
│   └── e2e/
├── build/
└── logs/
```

## 目录职责

### 根目录

- `main.go`：加载 `.env` 和文件配置，调用 `bootstrap`，启动服务并处理优雅退出。
- `.env.example`：环境变量示例，不包含任何真实凭据。
- `config/`：真实但非敏感的 YAML/JSON 配置文件。
- `Makefile`：统一开发、测试、构建入口。
- `Dockerfile`：服务镜像构建入口。

### `internal/router`

集中创建 HTTP Engine、注册业务路由和路由分组。`middleware/` 放鉴权、Trace ID、访问日志、Recovery、CORS 等通用中间件。

### `internal/handler`

HTTP 协议层，按业务域建子目录。负责：

- 解析 Path、Query、Header 和 Body；
- 参数校验；
- 调用 Service；
- 将业务结果或错误转换为 HTTP 响应。

Handler 不直接访问数据库，也不实现核心业务规则。

### `internal/service`

业务逻辑层，按业务域建子目录。负责业务校验、流程编排、事务组织，并调用 `common/db`、缓存、消息队列或外部 Client。

### `internal/model`

存放跨 Handler、Service 和 DB 层使用的业务实体及通用数据结构。HTTP 专属 Request/Response 优先留在对应 Handler 内。

### `internal/common`

存放跨业务模块复用的基础能力：

- `config/`：`.env` 与 `config/` 文件的加载、解析和校验；
- `db/`：数据库连接、事务和业务数据访问实现；
- `cache/`：Redis 等缓存能力；
- `mq/`：Kafka 等消息生产与消费基础能力；
- `client/`：外部 HTTP/RPC Client；
- `logger/`：结构化日志；
- `errors/`：统一业务错误；
- `response/`：统一 HTTP 响应结构；
- `utils/`：无业务语义的小型工具函数。

`common` 不能成为杂物目录。只有明确跨模块复用、且没有具体业务归属的代码才能放入。

### `internal/process`

可选的非 HTTP 入口，例如 Kafka Consumer、定时任务和后台任务。它们进入系统后仍应调用 Service，不直接实现业务或操作数据库。

### `internal/bootstrap`

负责：

- 初始化配置、日志、数据库、缓存和外部 Client；
- 构造 Repository、Service 和 Handler；
- 注册 HTTP 路由及后台任务；
- 控制启动顺序、健康状态、优雅退出和资源释放。

避免在 `main.go` 中堆积大量初始化逻辑，也避免依赖全局变量和隐式 `init()`。

### 其他目录

- `scripts/`：可重复执行的开发、构建和运维脚本；
- `tools/`：配置检查、数据修复等独立工具；
- `test/integration/`：数据库、Redis、Kafka 等集成测试；
- `test/e2e/`：从 HTTP 入口验证完整业务链路；
- `build/`：二进制和镜像构建脚本；
- `logs/`：本地日志输出目录，实际日志不提交。

## 当前数据库初始化

Demo 暂未引入版本化 Migration。`config/app.yaml` 中的 `database.auto_migrate` 默认为 `true`，启动时通过 GORM `AutoMigrate` 创建或更新 `users` 表。生产化前应替换为独立的版本化 Migration 流程，并关闭自动迁移。

## 分层约束

```text
main
  -> bootstrap
    -> router
      -> middleware
      -> handler
        -> service
          -> common/db | cache | mq | client
```

- Router 不写业务逻辑。
- Handler 不直接访问 DB。
- Service 不依赖 Gin/Fiber 等 HTTP 框架类型。
- DB 层不依赖 Handler 或 Router。
- 后台 Process 和 HTTP Handler 复用同一套 Service。

## 示例请求链路

```text
GET /api/v1/users/get?user_id=<uuid>
  -> internal/router/route_user.go
  -> internal/router/middleware/*
  -> internal/handler/user/handler.go
  -> internal/service/user/service.go
  -> internal/common/db/user/repository.go
  -> PostgreSQL
```

User API 使用 RPC 风格动作路径：

```text
GET  /api/v1/users/list
GET  /api/v1/users/get
POST /api/v1/users/create
POST /api/v1/users/update
POST /api/v1/users/delete
```

GET 接口通过 Query 传输字段，但 Handler 仍绑定到独立 Request Struct；不使用兼容性较差的 GET Body。所有接口都有明确的 Response Struct。
