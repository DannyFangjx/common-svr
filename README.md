# common-svr

`common-svr` 是一个面向 Go HTTP 服务的通用脚手架，目录组织参考 Eve `chat-control`，并增加独立的 `bootstrap` 装配层。

仓库包含一个可运行 Demo：基础探活接口和基于 PostgreSQL 的 User CRUD。

## 核心调用链

```text
Router -> Middleware -> Handler -> Service -> Common/DB
```

- `Handler` 负责协议解析、参数校验和响应转换。
- `Service` 负责业务逻辑和流程编排。
- `Common/DB` 负责数据库连接、事务和数据持久化。
- `Bootstrap` 负责依赖初始化、对象装配、启动顺序和资源释放。

## Demo 接口

```text
GET    /ping
GET    /test
GET    /health
GET    /ready

GET  /api/v1/users/list
GET  /api/v1/users/get
POST /api/v1/users/create
POST /api/v1/users/update
POST /api/v1/users/delete
```

接口采用 RPC 风格：路径表达动作，每个接口都有明确的 Request/Response Struct。GET 请求字段通过 Query 传输，POST 请求字段通过 JSON Body 传输。

查询示例：

```text
GET /api/v1/users/list?limit=20&offset=0
GET /api/v1/users/get?user_id=<uuid>
```

创建请求示例：

```json
{
  "name": "Alice",
  "email": "alice@example.com"
}
```

更新和删除请求会把 `user_id` 放在 JSON Body 中，不使用 Path Parameter。

## 本地启动

1. 复制环境变量模板：`cp .env.example .env`。
2. 将 `.env` 中的 `COMMON_PG_PASSWORD` 替换为本地开发密码。
3. 执行 `docker compose up -d common-pg`，启动项目独立的 PostgreSQL。
4. 执行 `make run`。

本地数据库实例：

```text
Docker container: common-pg
PostgreSQL user: common_server
PostgreSQL database: common_server_db
Host port: 55433
Docker volume: common-pg-data
```

`DATABASE_DSN` 使用 GORM PostgreSQL DSN，例如：

```text
host=<host> user=<user> password=<password> dbname=<database> port=5432 sslmode=disable TimeZone=Asia/Shanghai
```

当前 Demo 在 `database.auto_migrate=true` 时启动自动创建或更新 `users` 表。这只是尚未引入 Migration 工具前的临时方案，不建议直接用于生产环境。

User CRUD 当前未增加鉴权，目的是完整展示分层调用链；将脚手架用于真实服务前，需要在路由层接入项目统一认证与授权中间件。

## OpenTelemetry

服务为每个 HTTP 请求创建 OpenTelemetry Span，并为 GORM/PostgreSQL 操作创建子 Span。日志保持 JSON 格式，处于有效 Span 上下文中的日志会在顶层自动增加 `trace_id` 和 `span_id`；`X-Request-ID` 对应的字段独立记录为 `request_id`。

`OTEL_EXPORTER_OTLP_ENDPOINT` 为 OTLP gRPC 地址，例如 `localhost:4317`。地址为空时仍生成本地 Trace/Span ID，但不创建 Exporter、不向外上报。

## 验证

```bash
make test
make test-integration
make build
```

- `make test`：单元测试，不依赖 Docker 或已启动的服务。
- `make test-integration`：通过 Testcontainers 启动一次性 PostgreSQL，验证完整的 Handler → Service → Repository → PostgreSQL 链路；不会复用或修改本地 `common-pg`。
- `make test-e2e`：对已启动的服务执行 User CRUD。默认地址为 `http://127.0.0.1:8080`，可通过 `BASE_URL` 覆盖：

```bash
BASE_URL=http://127.0.0.1:8080 make test-e2e
```

## 配置约定

- `main.go` 是唯一服务入口，启动时通过结构体标签一次性加载环境变量；本地开发额外自动加载 `.env`。
- `.env.example` 列出全部运行参数及安全默认值，测试和生产环境由部署系统注入对应值。
- `.env`、密钥、Token、密码及运行日志禁止提交。
- `config/` 只存放证书、CA、规则和模板等静态资源；环境变量保存所需资源的文件路径。
- OpenTelemetry OTLP 地址为空时不上传链路数据。

完整目录说明见 [docs/directory-structure.md](docs/directory-structure.md)。
