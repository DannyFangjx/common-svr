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
2. 在 `.env` 中设置本地 PostgreSQL 的 `DATABASE_DSN`。
3. 执行 `make run`。

`DATABASE_DSN` 使用 GORM PostgreSQL DSN，例如：

```text
host=<host> user=<user> password=<password> dbname=<database> port=5432 sslmode=disable TimeZone=Asia/Shanghai
```

当前 Demo 在 `database.auto_migrate=true` 时启动自动创建或更新 `users` 表。这只是尚未引入 Migration 工具前的临时方案，不建议直接用于生产环境。

User CRUD 当前未增加鉴权，目的是完整展示分层调用链；将脚手架用于真实服务前，需要在路由层接入项目统一认证与授权中间件。

## 验证

```bash
make test
make build
```

## 配置约定

- `main.go` 是唯一服务入口，启动时加载 `.env` 和 `config/app.yaml`。
- `.env.example` 只描述环境变量名称和安全默认值。
- `.env`、密钥、Token、密码及运行日志禁止提交。
- `config/` 可以存放真实的非敏感 YAML/JSON 配置。

完整目录说明见 [docs/directory-structure.md](docs/directory-structure.md)。
