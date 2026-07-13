# 双实例联合端到端测试

这套测试用于 CI，不访问线上站点，也不需要真实域名、管理员 Token 或人工准备数据。

测试在两个已运行全部正式 migration 的临时 PostgreSQL 数据库上创建 A、B 两个完整 Fiber 应用。两个应用使用保留的公网文档地址作为逻辑 origin，跨站请求通过测试文件内的白名单 `RoundTripper` 直接进入另一个 Fiber 应用：生产 HTTP Signature、URL 安全校验、路由、中间件、数据库和回调逻辑都会执行，但不会监听端口或访问外网。

## 覆盖范围

- A、B 双向 manifest、公钥和 endpoints 发现
- 未签名引用请求拒绝且不落库
- A → B、B → A 的签名引用、人工审核、结果回调
- 引用重复审核不重复发送回调
- A → B、B → A 的签名提及、用户解析、审核和回调
- A → B、B → A 的联合友链申请、验签、审批和回调
- 友链审批后的双方实例激活、友链创建和首次时间线缓存
- 每个实例独立管理员、JWT、文章、配置和 RSA 密钥

## 本地运行

准备两个已经执行 `server/migrations` 全部 migration 的临时 PostgreSQL 数据库：

```bash
export FEDERATION_E2E_DB_DSN_A='postgres://postgres:postgres@127.0.0.1:5432/grtblog_e2e_a?sslmode=disable'
export FEDERATION_E2E_DB_DSN_B='postgres://postgres:postgres@127.0.0.1:5432/grtblog_e2e_b?sslmode=disable'

make test-federation-e2e
```

没有设置这两个 DSN 时，测试会安全跳过。数据库必须是临时空库；测试会向其中写入 fixture。GitHub Actions 的 `server` job 会自动创建两个数据库、运行 Goose migration、执行测试，任务结束后 PostgreSQL service 自动销毁。

测试使用 build tag `federation_e2e`，不会混入普通 `go test ./...`。
