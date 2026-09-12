# grtblog CLI

grtblog 博客系统的命令行管理工具。使用**管理员令牌（admin token，`gt_` 开头）**连接站点，在终端里管理文章、手记、思考、评论、媒体文件与设置项。

适合这些场景：

- 用本地编辑器写 Markdown，保存即发布
- `grtblog thinking "一句话"` 快速发思考
- 上传图片后直接拿到 Markdown 链接
- 在终端里处理待审评论、批量发布/置顶/删除
- 脚本化/CI 中调用（`--json` 输出）

复杂配置（联合协议、邮件、AI、OAuth 等）仍建议在管理后台操作。

## 安装

需要 Go 1.25+。

```bash
# 从源码安装（版本随项目 tag，例如 v2.2.0）
go install github.com/grtsinry43/grtblog/cli/v2/cmd/grtblog@v2.2.0

# 或安装最新版本
go install github.com/grtsinry43/grtblog/cli/v2/cmd/grtblog@latest

# 或在 cli/ 目录本地构建
make build          # 生成 ./grtblog，版本取项目最新 tag
make install        # 安装到 $GOBIN
```

> CLI 的版本与项目版本保持一致（`grtblog version` 会显示 `v2.2.0`）。
> 项目发版时会同时打 `vX.Y.Z` 和 `cli/vX.Y.Z` 两个 tag，后者供 `go install` 使用。

## 快速开始

1. 在管理后台 **设置 → 令牌** 创建一个管理员令牌（`gt_` 开头），复制保存。
2. 登录：

```bash
grtblog auth login --server https://your.blog --token gt_xxx
```

3. 开始使用：

```bash
grtblog article ls
grtblog thinking "今天天气不错"
grtblog upload ./cover.png --markdown
```

## 认证

CLI 直接使用管理员令牌作为 `Authorization` 头（`gt_xxx` 原样发送），无需用户名密码，也不经过人机校验。

```bash
grtblog auth login   # 交互式输入 server 和 token，并校验令牌有效性
grtblog auth status  # 查看当前站点、用户与令牌状态
grtblog auth logout  # 删除本地凭据（不会吊销服务端令牌）
```

### 多站点（profile）

默认写入 `default` profile，可用 `--as` 指定其他名称：

```bash
grtblog auth login --server http://127.0.0.1:8080 --token gt_dev --as dev
grtblog article ls --profile dev        # 使用 dev 站点
grtblog --profile dev status
```

### 环境变量

| 变量 | 说明 |
|------|------|
| `GRTBLOG_SERVER` | 服务器地址 |
| `GRTBLOG_TOKEN` | 管理员令牌 |
| `GRTBLOG_PROFILE` | 使用的 profile |
| `GRTBLOG_CONFIG` | 配置文件路径 |
| `GRTBLOG_EDITOR` | 编辑器（其次 `VISUAL` / `EDITOR` / `vi`） |
| `NO_COLOR` | 设置后禁用颜色 |

优先级：命令行 flag > 环境变量 > 配置文件。

### 配置文件

默认路径 `~/.config/grtblog/config.yaml`（权限 0600）：

```yaml
current: default
profiles:
  default:
    server: https://your.blog
    token: gt_xxx
  dev:
    server: http://127.0.0.1:8080
    token: gt_dev
```

## 命令一览

```
auth        login / logout / status
article     文章：ls / view / new / edit / rm / publish / unpublish / top / untop
moment      手记：ls / view / new / edit / rm / publish / unpublish / top / untop
thinking    思考：ls / new / edit / rm（也可直接 grtblog thinking "内容"）
upload      上传文件，输出 URL 或 Markdown 链接
file        媒体库：ls / rename / rm / download / sync
comment     评论：ls / reply / approve / reject / block / rm / viewed
config      设置项：ls / get / set
category    分类：ls / new / rm
column      专栏：ls / new / rm
tag         标签：ls / new / rm
status      查看服务器运行状态
version     版本信息
completion  生成 shell 补全脚本
```

查看任意命令帮助：`grtblog <命令> --help`。

## 编辑器工作流

`article new` / `article edit` / `moment new` / `moment edit` 会打开 `$EDITOR`，
文件以 YAML front-matter + Markdown 正文组成，保存后自动提交。

```markdown
---
title: 我的新文章
summary: ""
category: 技术          # 分类名称或 ID
tags: [Go, 随笔]        # 标签名称或 ID
shortUrl: ""            # 留空由服务端生成
published: false        # 默认草稿
top: false
original: true
allowComment: true
---

正文（Markdown）……
```

手记的 front-matter 使用 `column`（专栏）、`topics`（话题）和 `images`（图片 URL 列表）。

也可以完全跳过编辑器：

```bash
grtblog article new -f post.md        # 从文件创建（可含 front-matter）
grtblog article new -f - < post.md    # 从标准输入创建
grtblog article new --publish         # 撰写后直接发布
grtblog article edit 42               # 编辑已有文章
```

## 常用示例

```bash
# 文章
grtblog article ls --search Go --published
grtblog article view 42
grtblog article publish 42 43
grtblog article top 42
grtblog article rm 43 -y

# 思考（最快路径）
grtblog thinking "刚看到一个很有意思的观点……"
echo "多行内容" | grtblog thinking

# 媒体
grtblog upload cover.png a.png --markdown
grtblog file ls
grtblog file sync

# 评论审核
grtblog comment ls --status pending
grtblog comment approve c_xxx
grtblog comment reply c_xxx "谢谢支持！"

# 设置项
grtblog config ls --group site
grtblog config get site.title
grtblog config set site.title "我的博客"

# 运维
grtblog status
```

## 输出与脚本

- 默认输出对齐表格；`--json` 输出结构化 JSON，便于 `jq` 处理
- 非 TTY 环境自动禁用颜色；`--no-color` 可强制关闭
- 破坏性操作（删除）默认二次确认，`-y/--yes` 跳过
- 出错时退出码为 `1`，认证失败会提示重新 `auth login`

```bash
grtblog article ls --json | jq '.items[] | {id, title}'
```

## 开发

```bash
cd cli
make test      # go test ./...
make vet       # go vet ./...
make build     # 构建到 ./grtblog
make fmt       # gofmt
```

目录结构：

```
cli/
├── cmd/grtblog/         # 入口
└── internal/
    ├── client/          # API 客户端（envelope、鉴权、上传下载）
    ├── commands/        # 各子命令
    ├── config/          # 多 profile 配置
    ├── editor/          # $EDITOR 工作流
    ├── frontmatter/     # YAML front-matter 解析
    └── output/          # 表格 / JSON / 颜色 / 确认
```

CLI 是纯 HTTP 客户端，不依赖 `server/` 的任何内部包，可独立构建与发布。
