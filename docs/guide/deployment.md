# 快速部署

GrtBlog 使用 Docker Compose 进行一键部署，无需手动配置复杂的运行环境。

## 环境要求

- 一台 Linux 服务器（推荐 1 核 1G 以上配置）
- 安装 Docker 和 Docker Compose
- 一个域名（可选，但推荐）

## 部署步骤

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

### 3. 启动服务

```bash
docker compose up -d --build
```

首次启动会自动完成数据库迁移，无需额外操作。

### 4. 验证部署

```bash
# 检查服务健康状态
curl -f http://localhost:80/healthz

# 检查后端活性
curl -f http://localhost:80/health/liveness
```

看到正常响应后，即可通过浏览器访问：

- **博客首页**: `http://your-server-ip`
- **管理后台**: `http://your-server-ip/admin/`

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

生产环境建议在前面加一层反向代理（如 Caddy 或 Nginx）来处理 HTTPS：

```nginx
# Nginx 示例
server {
    listen 443 ssl http2;
    server_name yourdomain.com;

    ssl_certificate     /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;

    location / {
        proxy_pass http://127.0.0.1:80;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # WebSocket 支持
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}
```

## 数据备份

重要数据存储在 Docker volumes 中：

| Volume | 内容 |
|--------|------|
| `postgres_data` | 数据库（文章、评论等所有数据） |
| `redis_data` | 缓存数据 |
| `html_data` | 生成的静态页面 |
| `uploads_data` | 上传的图片和文件 |

建议定期备份数据库：

```bash
docker compose exec postgres pg_dump -U postgres grtblog > backup.sql
```

## 升级

```bash
cd deploy
git pull
docker compose up -d --build
```

数据库迁移会在启动时自动执行。

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
docker compose run --rm migrate sh -lc \
  'goose -table public.goose_db_version -dir /app/migrations postgres "$DB_DSN" status'
```

如需回滚最近一次迁移：
```bash
docker compose run --rm migrate sh -lc \
  'goose -table public.goose_db_version -dir /app/migrations postgres "$DB_DSN" down'
```
