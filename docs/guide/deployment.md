# 快速部署

GrtBlog 使用 Docker Compose 进行一键部署，无需手动配置复杂的运行环境。

## 环境要求

- 一台 Linux 服务器（推荐 1 核 1G 以上配置，支持 amd64 / arm64）
- 安装 Docker 和 Docker Compose
- 一个域名（可选，但推荐）

## 方式一：使用预构建镜像（推荐）

每次发布新版本时，GitHub Actions 会自动构建并推送多架构镜像到 GHCR。这是最简单的部署方式，无需克隆完整代码仓库，也无需本地构建。

### 1. 准备部署目录

只需下载 `deploy` 目录和 Nginx 配置：

```bash
# 创建部署目录
mkdir -p grtblog && cd grtblog

# 下载所需文件
BASE_URL="https://raw.githubusercontent.com/grtsinry43/grtblog-v2/main"
curl -fsSL "$BASE_URL/deploy/docker-compose.yml" -o docker-compose.yml
curl -fsSL "$BASE_URL/deploy/.env.example"       -o .env

mkdir -p nginx
curl -fsSL "$BASE_URL/deploy/nginx/nginx.conf"    -o nginx/nginx.conf
```

### 2. 配置环境变量

编辑 `.env`，设置镜像来源和版本：

```ini
# 使用 GHCR 预构建镜像
IMAGE_REPO_PREFIX=ghcr.io/grtsinry43/
APP_VERSION=1.0.0
APP_UPDATE_CHANNEL=stable

# 数据库密码（请设置为强密码）
POSTGRES_PASSWORD=your-secure-password

# 认证密钥（请使用随机字符串）
AUTH_SECRET=your-random-secret-string
```

::: tip 生成随机密钥
可以使用以下命令生成：
```bash
openssl rand -hex 32
```
:::

::: tip 查看可用版本
所有版本可在 GitHub Releases 页面查看：

`https://github.com/grtsinry43/grtblog-v2/releases`

也可通过 GHCR 查看所有镜像标签：

`https://github.com/grtsinry43/grtblog-v2/pkgs/container/grtblog-server`
:::

### 3. 启动服务

```bash
mkdir -p storage/html storage/uploads storage/geoip
docker compose up -d
```

首次启动会自动拉取镜像并完成数据库迁移，无需额外操作。

### 4. 升级

```bash
# 修改 .env 中的 APP_VERSION 为新版本号，然后：
docker compose pull server renderer
docker compose up -d server renderer
```

Nginx 无需重启，会通过 Docker DNS 自动发现新容器。

---

## 方式二：本地构建

如果你需要自定义构建参数（如修改管理面板配置），或者无法访问 GHCR，可以克隆仓库本地构建。

### 1. 获取代码

```bash
git clone https://github.com/grtsinry43/grtblog-v2.git
cd grtblog-v2/deploy
```

### 2. 配置环境变量

```bash
cp .env.example .env
```

编辑 `.env` 文件，至少修改以下配置：

```ini
# 本地构建不需要设置 IMAGE_REPO_PREFIX
IMAGE_REPO_PREFIX=
APP_VERSION=local

# 数据库密码（请设置为强密码）
POSTGRES_PASSWORD=your-secure-password

# 认证密钥（请使用随机字符串）
AUTH_SECRET=your-random-secret-string
```

### 3. 启动服务

```bash
mkdir -p storage/html storage/uploads storage/geoip
docker compose up -d --build
```

首次启动会自动完成数据库迁移，无需额外操作。

### 4. 升级

```bash
git pull
docker compose up -d --build
```

---

## 验证部署

```bash
# 检查服务健康状态
curl -f http://localhost:80/healthz

# 检查后端活性
curl -f http://localhost:80/health/liveness
```

看到正常响应后，即可通过浏览器访问：

- **博客首页**: `http://your-server-ip`
- **管理后台**: `http://your-server-ip/admin/`

## 镜像标签说明

| Tag 类型 | 示例 | 说明 |
|----------|------|------|
| 完整版本号 | `1.2.3` | 精确版本，推荐生产使用 |
| 主次版本号 | `1.2` | 自动跟随该次版本的最新补丁 |
| `latest` | `latest` | 最新 stable 版本 |
| 预发布版本 | `2.0.0-alpha.1` | 测试版本，不附带 `latest` 标签 |

预发布版本（含 `-alpha` / `-beta` / `-rc` 后缀）不会更新 `latest` 和主次版本标签，适合提前测试。

更新检查通道与镜像标签建议配套使用：

| 用途 | 推荐 `APP_VERSION` | 推荐 `APP_UPDATE_CHANNEL` |
|----------|------|------|
| 生产稳定 | `1.2.3` | `stable` |
| 预发布验证 | `2.1.0-beta.1` | `preview` |
| 跟随 stable 滚动 | `stable` | `stable` |
| 跟随 preview 滚动 | `preview` | `preview` |

其中 `APP_UPDATE_CHANNEL=stable` 会读取 GitHub Releases，`APP_UPDATE_CHANNEL=preview` 会读取 Git tags，并默认只提示当前 major 内的预发布版本。

## 可选配置

### Cloudflare Turnstile 验证

如果需要防止垃圾评论，可以启用 Turnstile 人机验证：

```ini
TURNSTILE_ENABLED=true
TURNSTILE_SECRET=your-turnstile-secret
```

### GeoIP 地理定位

启用后可以在管理后台查看访客的地理分布：

```ini
GEOIP_DB_URL=https://github.com/P3TERX/GeoLite.mmdb/raw/download/GeoLite2-City.mmdb
GEOIP_ASN_DB_URL=https://github.com/P3TERX/GeoLite.mmdb/raw/download/GeoLite2-ASN.mmdb
```

### 自定义端口

默认使用 80 端口，可以通过 `NGINX_PORT` 修改：

```ini
NGINX_PORT=8080
```

## 反向代理与 HTTPS

生产环境建议在前面加一层反向代理（如 Caddy 或 Nginx）来处理 HTTPS。内层 Nginx 已经处理了路由分发，外层只需透传即可，但有几个路径需要特殊配置。

### Nginx 示例

```nginx
server {
    listen 80;
    server_name yourdomain.com;
    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl http2;
    server_name yourdomain.com;

    ssl_certificate     /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;

    # 与内层保持一致，否则大文件上传会被外层拦截
    client_max_body_size 200M;

    # 透传真实 IP（内层通过 X-Real-IP 识别客户端 IP）
    proxy_set_header Host              $host;
    proxy_set_header X-Real-IP         $remote_addr;
    proxy_set_header X-Forwarded-For   $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;

    # WebSocket（实时通知推送）
    location /api/v2/ws/ {
        proxy_pass http://127.0.0.1:80;
        proxy_http_version 1.1;
        proxy_set_header Upgrade    $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_read_timeout 86400s;
    }

    # SSE 流式接口（AI 重写 / 摘要生成）
    location ~ ^/api/v2/admin/ai/.+/stream$ {
        proxy_pass http://127.0.0.1:80;
        proxy_http_version 1.1;
        proxy_set_header Connection "";
        proxy_buffering off;
        proxy_request_buffering off;
        proxy_cache off;
        proxy_read_timeout 3600s;
        proxy_send_timeout 3600s;
        add_header X-Accel-Buffering no;
    }

    # 默认转发
    location / {
        proxy_pass http://127.0.0.1:80;
    }
}
```

### Caddy 示例

Caddy 默认透传 Host 和 X-Forwarded-* 头，配置更简洁：

```
yourdomain.com {
    reverse_proxy localhost:80
}
```

::: warning Caddy 注意事项
Caddy 默认的反代超时可能不够长，WebSocket 和 SSE 流式接口需要额外配置超时时间，否则长连接会被提前断开。
:::

### 关键注意事项

| 项目 | 说明 |
|------|------|
| `X-Real-IP` | **必须设置**。内层 Nginx 通过此头获取真实客户端 IP，用于评论显示、访问日志等 |
| `Host` 头 | **必须透传**。后端依赖它生成 ActivityPub Actor URL、RSS 链接等 |
| WebSocket | `/api/v2/ws/` 需透传 `Upgrade` + `Connection` 头，否则实时通知无法工作 |
| SSE 流式 | AI 相关的流式接口使用 SSE，外层必须关闭 `proxy_buffering`，否则响应会被缓冲导致流式效果失效 |
| `client_max_body_size` | 内层限制 200M，外层应 ≥ 200M |
| ActivityPub | `/.well-known/`、`/ap/`、`/nodeinfo/` 等联邦路径无需特殊处理，普通转发即可 |

## 数据备份

重要数据存储在 Docker volumes 中：

| Volume | 内容 |
|--------|------|
| `postgres_data` | 数据库（文章、评论等所有数据） |
| `redis_data` | 缓存数据 |
| `./storage/html` | 生成的静态页面与客户端资源 |
| `./storage/uploads` | 上传的图片和文件 |
| `./storage/geoip` | GeoIP 数据库缓存 |

建议定期备份数据库：

```bash
docker compose exec postgres pg_dump -U postgres grtblog > backup.sql
```

## 故障排查

### 服务无法启动

```bash
# 查看容器日志
docker compose logs server
docker compose logs renderer
```

### 页面显示异常

检查静态文件生成状态：
```bash
docker compose logs server | grep -i "isr\|render\|html"
```

### 数据库迁移问题

查看迁移状态：
```bash
docker compose exec server goose -table public.goose_db_version -dir /app/migrations postgres "$DB_DSN" status
```

如需回滚最近一次迁移：
```bash
docker compose exec server goose -table public.goose_db_version -dir /app/migrations postgres "$DB_DSN" down
```
