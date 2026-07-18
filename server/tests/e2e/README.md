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

## 完整备份与恢复容器 E2E

`make test-backup-e2e` 会构建正式 server 镜像，并启动隔离的 PostgreSQL 17、Redis 和 server 容器。测试实际执行以下往返，不会复用本机数据库或存储：

- 注册管理员，写入数据库探针和真实上传卷文件
- 运行 `pg_dump`，轮询后台任务，使用签名链接下载并检查 `tar.gz` manifest
- 到期触发计划备份并固定归档
- 篡改数据库与上传卷，从历史备份触发服务重启和离线 `pg_restore`
- 验证数据库值、原上传文件和额外文件删除均已恢复
- 清空业务 schema 模拟全新安装，再通过初始化恢复接口上传归档并验证管理员回归

测试结束后会删除专属 Compose 项目及临时 volume。需要本机已安装 Docker Compose、curl、jq 和 tar。
