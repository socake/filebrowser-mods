# lumen — LumenBrowser CLI

LumenBrowser 的命令行客户端。把服务端的 REST API 封装成终端命令，便于**脚本化、CI 集成、批量上传与定时备份**。

> 服务端**无需改动**：所有命令都走既有 HTTP 接口。

## 与服务端的关系

本目录是仓库内的**独立 Go 模块**（自己的 `go.mod`），通过仓库根的 `go.work` 与服务端主模块连在一起：

```
filebrowser-mods/        # 一个 git 仓库
├── go.work              # 工作区：连接下面两个模块
├── go.mod               # 服务端主模块 github.com/filebrowser/filebrowser/v2
└── client/              # ← 你在这里
    └── go.mod           # CLI 模块 github.com/socake/filebrowser-mods/client
```

独立 `go.mod` 的目的：CLI 只背自己需要的轻依赖（第一版**零外部依赖**），不扛服务端的 SQLite / 图像处理等重依赖，编出来的 `lumen` 是最小单二进制。

## 构建

需要 Go 1.25+（`go.work` 已配好）：

```bash
# 在仓库根目录
go build -o lumen ./client

# 或进 client 目录
cd client && go build -o lumen .
```

## 命令

| 命令 | 作用 | 背后 API | 状态 |
|---|---|---|---|
| `lumen login <url>` | 登录并保存连接配置 | `POST /api/login` | 可用 |
| `lumen ls <远程路径>` | 列出远程目录 | `GET /api/resources` | 可用 |
| `lumen put <本地> <远程>` | 上传文件（远程以 `/` 结尾则拼本地文件名） | `POST /api/resources?override=true` | 可用 |
| `lumen get <远程> [本地]` | 下载文件（本地省略则用远程文件名） | `GET /api/raw` | 可用 |
| `lumen rm <远程>` | 删除文件或目录（目录递归） | `DELETE /api/resources` | 可用 |
| `lumen mv <源> <目标>` | 重命名/移动 | `PATCH /api/resources?action=rename&destination=` | 可用 |
| `lumen mkdir <远程目录>` | 新建目录 | `POST /api/resources/<path>/` | 可用 |
| `lumen share <远程>` | 生成分享链（`-password`/`-expires`/`-unit`） | `POST /api/share` | 可用 |
| `lumen shares` | 列出我的分享 | `GET /api/shares` | 可用 |
| `lumen search [-in 目录] <关键词>` | 搜索文件（流式 NDJSON） | `GET /api/search` | 可用 |
| `lumen drop <文件>` | 免登录投递，上传即得永久链 | `POST /api/public/upload` | 可用 |

> `put` 第一版用 `POST /api/resources?override=true` 直传（流式，不读入内存），尚未走 `/api/tus` 分块续传；超大文件续传是后续工作。

## 目录结构

```
client/
├── main.go            # 入口 + 子命令分发
├── login.go           # lumen login
├── ls.go              # lumen ls
├── put.go get.go      # lumen put / get
├── rm.go mv.go mkdir.go  # lumen rm / mv / mkdir
├── share.go shares.go    # lumen share / shares
├── search.go          # lumen search
├── drop.go            # lumen drop
├── internal/
│   ├── config/        # ~/.lumen/config.json 读写（profile / token）
│   └── api/           # REST 封装（login / list / upload / put / get / rm / mv / mkdir / share / search ...）
└── README.md
```

## 配置文件

登录后连接信息保存在 `~/.lumen/config.json`：

```json
{
  "server": "https://files.example.com",
  "username": "alice",
  "token": "<jwt>",
  "expires": "2026-06-22T10:00:00Z"
}
```

## 实施路线（第一阶段）

最小验证链路：先打通 `login` → `ls` → `drop` 三条命令，确认 REST 封装可行，再扩展其余命令。

1. `internal/config`：配置读写
2. `internal/api`：login 拿 token、list、public upload
3. 接通 `login` / `ls` / `drop`
4. 连本地服务端联调验证
