# LumenBrowser 私有化部署文档

> 本文是 LumenBrowser 最详细的逐步部署指南，涵盖：构建（含交叉编译与静态编译坑）、Docker 部署、镜像仓库分发（ACR / Docker Hub）、Nginx 反向代理、certbot HTTPS 证书、首次登录与安全、数据备份。
>
> 想要最短上手路径，请看仓库根 [`README.md`](../README.md) 的「快速开始」。在线演示：**[file.vishine.top](https://file.vishine.top)**。

---

## 目录

1. [架构与产物](#1-架构与产物)
2. [前置依赖](#2-前置依赖)
3. [一键构建（scripts/build.sh）](#3-一键构建scriptsbuildsh)
4. [手动构建与关键坑：CGO_ENABLED=0](#4-手动构建与关键坑cgo_enabled0)
5. [Docker 部署](#5-docker-部署)
6. [镜像仓库分发（ACR / Docker Hub）](#6-镜像仓库分发acr--docker-hub)
7. [Nginx 反向代理](#7-nginx-反向代理)
8. [HTTPS：certbot 自动签发证书](#8-httpscertbot-自动签发证书)
9. [首次登录与安全](#9-首次登录与安全)
10. [数据备份与恢复](#10-数据备份与恢复)
11. [常见问题（FAQ）](#11-常见问题faq)

---

## 1. 架构与产物

LumenBrowser 最终是**单个静态 Go 二进制 + 一个数据库文件**：

- **前端**（Vue 3 + Vite）`pnpm build` 产出的 dist，通过 `go:embed` 打进二进制。
- **MCP 二进制**（`lumen-mcp`，各平台预编译）也被嵌入二进制，供设置页直接下载，用户无需自行构建。
- **后端**编译为单个可执行文件 `./filebrowser`，运行时只需一个 SQLite 数据库文件。

也就是说，构建机需要 Go + Node 工具链；**运行机什么都不用装**，拷过去一个二进制即可跑（或用 Docker）。

运行时挂载点：

| 路径 | 含义 | 建议挂载 |
| --- | --- | --- |
| `/srv` | 对外暴露的文件根目录（`-r`） | 你的数据盘目录 |
| `/db/filebrowser.db` | 数据库（用户 / 设置 / 分享，`-d`） | 独立持久卷 |

---

## 2. 前置依赖

构建机需要：

- **Go 1.25+**。本仓库用 `go.work` 连了三个模块（主程序 `.`、`./client`、`./mcp`）。
  - 项目约定 Go SDK 装在 `~/go-sdk`，`build.sh` 会自动 `export PATH=$HOME/go-sdk/bin:$PATH GOTOOLCHAIN=local`。如果你的 Go 在别处，改这一行或确保 `go` 在 `PATH` 即可。
- **Node.js + pnpm**（前端构建）。没有 pnpm：`npm i -g pnpm` 或 `corepack enable`。

运行机需要：

- 裸机运行：无（静态二进制）。
- Docker 运行：Docker（镜像基于 `alpine:3.23`，已装 `ca-certificates mailcap tzdata`）。

---

## 3. 一键构建（scripts/build.sh）

仓库提供 [`scripts/build.sh`](../scripts/build.sh)，一条命令完成全部构建：

```bash
./scripts/build.sh
```

它做三件事：

1. **交叉编译 `lumen-mcp` 各平台二进制** → 输出到 `http/mcpbin/`（不入 git），供 MCP 设置页直接下载：
   - `windows/amd64`、`darwin/arm64`、`darwin/amd64`、`linux/amd64`
   - 用 `GOOS=... GOARCH=... GOWORK=off go build -ldflags="-s -w"` 逐个编出。
2. **构建前端**：`cd frontend && pnpm install（首次）&& vite build`。
3. **构建主程序**：`CGO_ENABLED=0 GOWORK=off go build -o filebrowser .`。

产物为仓库根的 `./filebrowser`，**已内嵌前端 dist + 各平台 lumen-mcp**。

> 构建后可直接本机验证：
> ```bash
> ./filebrowser -r /tmp/lumen-data -d /tmp/lumen.db -a 0.0.0.0 -p 8080
> ```

---

## 4. 手动构建与关键坑：CGO_ENABLED=0

如果你不用脚本、手动分步构建：

```bash
# 1) 交叉编译 MCP（可选，仅当需要设置页下载功能）
cd mcp
for t in windows/amd64 darwin/arm64 darwin/amd64 linux/amd64; do
  os=${t%/*}; arch=${t#*/}; ext=""; [ "$os" = windows ] && ext=.exe
  GOOS=$os GOARCH=$arch GOWORK=off go build -ldflags="-s -w" \
    -o "../http/mcpbin/lumen-mcp-$os-$arch$ext" .
done
cd ..

# 2) 前端
cd frontend && pnpm install && pnpm build && cd ..

# 3) 后端（静态编译）
CGO_ENABLED=0 GOWORK=off go build -o filebrowser .
```

> ### ⚠️ 关键坑：必须 `CGO_ENABLED=0` 静态编译
>
> 如果用默认（CGO 开启）编译，二进制会动态链接 glibc。一旦放到 **alpine / musl** 这类没有 glibc 的环境（包括本项目的 `alpine:3.23` 镜像）运行，会直接报：
>
> ```
> exec /bin/filebrowser: no such file or directory
> ```
>
> 这个报错极具迷惑性 —— 文件明明在，却说找不到，实际是动态链接器找不到。**解决办法就是 `CGO_ENABLED=0` 做静态编译**。`scripts/build.sh` 已默认带上，手动构建务必自己加。

单独构建 `lumen` CLI：

```bash
go build -o lumen ./client
```

---

## 5. Docker 部署

二进制由本地构建（含前端 + MCP embed），[`Dockerfile.demo`](../Dockerfile.demo) 只负责打包：

```dockerfile
FROM alpine:3.23
RUN apk add --no-cache ca-certificates mailcap tzdata && update-ca-certificates
COPY filebrowser /bin/filebrowser
COPY deploy/demo-data /srv          # 内置演示数据，自有部署可换成空目录
RUN mkdir -p /db && chmod +x /bin/filebrowser
EXPOSE 80
VOLUME /db
ENTRYPOINT ["/bin/filebrowser", "-d", "/db/filebrowser.db", "-r", "/srv", "-a", "0.0.0.0", "-p", "80"]
```

构建并运行：

```bash
# 先构建二进制
./scripts/build.sh

# 打镜像
docker build -f Dockerfile.demo -t lumenbrowser .

# 运行（把 /srv 和 /db 挂到宿主机持久化）
docker run -d --name lumenbrowser \
  -p 80:80 \
  -v /your/data:/srv \
  -v /your/db:/db \
  lumenbrowser
```

说明：

- `-v /your/data:/srv`：对外暴露的文件目录，换成你真实的数据目录。
- `-v /your/db:/db`：数据库持久化（`/db/filebrowser.db`）。**不挂这个卷，容器重建后用户 / 设置 / 分享会全丢**。
- 生产环境建议**不要**用 `Dockerfile.demo` 内置的 `deploy/demo-data`，改为挂载空目录或自己的数据，避免演示文件混入。可以自己写一个不 COPY demo-data 的 Dockerfile，或运行后清空 `/srv`。
- 端口：示例直接 `80:80`。放到 Nginx 后面时建议改成 `127.0.0.1:8080:80`，只让本机的 Nginx 反代访问（见第 7 节）。

查看日志 / 进容器：

```bash
docker logs -f lumenbrowser
docker exec -it lumenbrowser sh
```

---

## 6. 镜像仓库分发（ACR / Docker Hub）

适合**本地（或 CI）构建好镜像 → 服务器只负责拉取运行**的场景，尤其是弱配置服务器（没有 Go / Node 工具链、内存小编不动）特别合适。

### 6.1 推到阿里云 ACR

```bash
# 登录（用 ACR 控制台给的用户名 / 密码）
docker login --username=<你的账号> registry.cn-hangzhou.aliyuncs.com

# 打 tag
docker tag lumenbrowser \
  registry.cn-hangzhou.aliyuncs.com/<命名空间>/lumenbrowser:latest

# 推送
docker push registry.cn-hangzhou.aliyuncs.com/<命名空间>/lumenbrowser:latest
```

服务器上拉取并运行：

```bash
docker login --username=<你的账号> registry.cn-hangzhou.aliyuncs.com
docker pull registry.cn-hangzhou.aliyuncs.com/<命名空间>/lumenbrowser:latest

docker run -d --name lumenbrowser \
  -p 127.0.0.1:8080:80 \
  -v /your/data:/srv \
  -v /your/db:/db \
  registry.cn-hangzhou.aliyuncs.com/<命名空间>/lumenbrowser:latest
```

### 6.2 推到 Docker Hub / 任意 registry

```bash
docker login                                   # Docker Hub
docker tag lumenbrowser <用户名>/lumenbrowser:latest
docker push <用户名>/lumenbrowser:latest

# 服务器
docker pull <用户名>/lumenbrowser:latest
docker run -d --name lumenbrowser -p 127.0.0.1:8080:80 \
  -v /your/data:/srv -v /your/db:/db <用户名>/lumenbrowser:latest
```

> 任意私有 registry（Harbor 等）同理：把镜像名前缀换成 `<registry 地址>/<项目>/lumenbrowser:tag`。
>
> 跨架构提示：示例只编了 `linux/amd64`。如果服务器是 ARM（如部分云主机 / 树莓派），需要 `docker buildx` 构建对应架构镜像，且后端二进制也要 `GOARCH=arm64 CGO_ENABLED=0` 重新编。

---

## 7. Nginx 反向代理

让 Nginx 监听 80/443，反代到本机容器（如 `127.0.0.1:8080`）。**`client_max_body_size 0` 是重点** —— 不设的话大文件上传会被 Nginx 默认 1MB 限制截断。

`/etc/nginx/conf.d/lumenbrowser.conf`：

```nginx
server {
    listen 80;
    server_name your.domain.com;

    # 关键：不限制上传体积，否则大文件 / 断点续传会被截断
    client_max_body_size 0;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;

        proxy_set_header Host              $host;
        proxy_set_header X-Real-IP         $remote_addr;
        proxy_set_header X-Forwarded-For   $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # WebSocket / 长连接（预览、进度等）
        proxy_set_header Upgrade    $http_upgrade;
        proxy_set_header Connection "upgrade";

        # 大文件传输别过早超时
        proxy_read_timeout  3600s;
        proxy_send_timeout  3600s;
        proxy_request_buffering off;   # 上传边收边转，省内存
    }
}
```

校验并重载：

```bash
nginx -t && systemctl reload nginx
```

> 对应地，容器请用 `-p 127.0.0.1:8080:80` 只绑本机，避免 8080 直接暴露公网绕过 Nginx。

---

## 8. HTTPS：certbot 自动签发证书

确保第 7 节的 HTTP server 已生效、域名已解析到本机，然后：

```bash
# 安装 certbot 与 nginx 插件（Debian/Ubuntu）
apt update && apt install -y certbot python3-certbot-nginx

# 自动签发并改写 Nginx 配置为 HTTPS（含 80→443 跳转）
certbot --nginx -d your.domain.com

# 续期是自动的（systemd timer / cron），可手动演练：
certbot renew --dry-run
```

certbot 会自动在上面的 server 块里加上 `listen 443 ssl`、证书路径，并新增一个 80→443 的跳转 server。`client_max_body_size 0` 等自定义指令会被保留。

---

## 9. 首次登录与安全

首次启动（无已存在配置 / 数据库）时自动创建第一个管理员：

| 项目 | 默认值 |
| --- | --- |
| 用户名 | `admin` |
| 密码 | `admin` |
| 语言 | 简体中文（zh-cn） |

> ### ⚠️ 部署后立即改密码
>
> 默认 `admin / admin` 是**公开已知凭据**。把服务暴露公网却不改密码，等于把整台服务器的文件读写权交给所有人，存在严重安全风险。
>
> 两种改法（二选一）：
>
> - **Web 界面**：右上角 *设置 → 个人设置（Profile Settings）* → 修改密码。
> - **CLI**（需先停服，避免数据库占用）：
>   ```bash
>   filebrowser users update admin --password <新密码> --database /db/filebrowser.db
>   # 容器内：docker exec -it lumenbrowser filebrowser users update admin --password <新密码> -d /db/filebrowser.db
>   ```

其他安全建议：

- 用第 7、8 节把服务放到 Nginx + HTTPS 后面，容器只绑 `127.0.0.1`。
- 在 *设置 → 角色权限* 配置好新用户默认角色，按最小权限原则发放。
- MCP 访问令牌按需选择有效期（永久 / 7d / 30d / 90d / 1y），长期不用及时回收。

---

## 10. 数据备份与恢复

要备份的只有两样东西：

1. **数据库** `/db/filebrowser.db` —— 用户、设置、分享、角色权限全在这里。
2. **数据目录** `/srv`（或你 `-r` 指向的目录）—— 实际文件。

### 冷备（最稳，先停服）

```bash
docker stop lumenbrowser
tar czf lumen-backup-$(date +%Y%m%d).tar.gz /your/db /your/data
docker start lumenbrowser
```

### 热备数据库（SQLite 在线安全备份）

不停服时，直接 `cp` 数据库可能拷到不一致状态，建议用 SQLite 的 `.backup`：

```bash
# 宿主机装了 sqlite3 的话
sqlite3 /your/db/filebrowser.db ".backup '/backup/filebrowser-$(date +%Y%m%d).db'"
```

数据目录 `/srv` 用 `rsync` 增量同步即可：

```bash
rsync -a --delete /your/data/ /backup/lumen-srv/
```

### 恢复

把备份的 `filebrowser.db` 放回 `/db/`、文件放回 `/srv/`，重启容器即可：

```bash
docker restart lumenbrowser
```

> 建议把上面命令塞进 cron 做定时备份，并把归档同步到异地 / 对象存储。

---

## 11. 常见问题（FAQ）

**Q：运行报 `exec /bin/filebrowser: no such file or directory`，文件明明在。**
A：CGO 编译进了 glibc 动态链接，alpine/musl 跑不起来。用 `CGO_ENABLED=0` 重新静态编译（见第 4 节）。

**Q：上传大文件失败 / 被截断。**
A：Nginx 没设 `client_max_body_size 0`（见第 7 节）。LumenBrowser 本身支持 tus 断点续传，但反代层会先卡住。

**Q：容器重建后用户和设置全没了。**
A：没挂 `/db` 卷，数据库随容器销毁。务必 `-v /your/db:/db`。

**Q：设置页下载的 lumen-mcp 是空的 / 404。**
A：构建时没交叉编译 MCP 二进制到 `http/mcpbin/`。用 `scripts/build.sh` 完整构建，或手动执行第 4 节第 1 步。

**Q：弱配置服务器编不动。**
A：用第 6 节，本地 / CI 构建镜像推到 registry，服务器只 `docker pull` + `run`。

**Q：忘了 admin 密码。**
A：停服后 `filebrowser users update admin --password <新密码> -d /db/filebrowser.db`（见第 9 节）。
