# LumenBrowser 产品设计文档

> Slogan：**你的文件之光 / Bring your files to light**
>
> 本文档基于 `filebrowser` v2.62.2 二改分支（`my-mods`）真实源码撰写，描述将该分支包装为
> 面向个人与小团队的自托管产品「LumenBrowser」的产品设计。技术实现细节见
> [`IMPLEMENTATION.md`](./IMPLEMENTATION.md)，接口契约见 [`API-SPEC.md`](./API-SPEC.md)，
> 已知缺口与路线见 [`ROADMAP.md`](./ROADMAP.md)。

---

## 1. 产品定位

LumenBrowser 是一个**自托管的文件存储/分享 + 轻量文档协作门户**。它把一台服务器（或一个目录）变成
干净、私有、完全归你掌控的文件库：对内是带登录的多用户文件管理器，对外是免登录的文档门户与投递箱。

- **形态**：单二进制 + 一个数据库文件（前端 dist 已内嵌进可执行文件），Docker 一行启动。
- **底座**：经典 filebrowser 的文件管理内核（资源 CRUD、断点续传、预览、搜索、分享、多用户权限），
  在其之上二改补齐了「对外协作层」（公开文档门户、免登录投递箱、增强分享页、Markdown 优先）。
- **数据主权**：数据始终在你自己的服务器上，不经过任何第三方。

### 一句话价值主张

> 在经典文件管理器之上，补齐**对外分享与文档协作**的关键一环 —— 既能私有管理，也能优雅地对世界发布。

---

## 2. 目标用户与使用场景

| 目标用户 | 典型场景 |
|---|---|
| **个人开发者 / 极客** | 把家用服务器或 VPS 上的目录变成私有云盘；对外发布一份在线文档库（README、设计稿、drawio 架构图） |
| **小团队 / 工作室** | 内部共享资料，对外用分享链接发交付物；用文档门户做轻量知识库 |
| **内容收集方（老师 / 编辑 / 活动主办）** | 用「免登录投递箱」收作业、收外部投稿，投递者无需注册即可上传并拿到链接 |
| **需要脚本化 / 自动化的人** | 用规划中的 `lumen` CLI 做批量上传、定时备份、CI 集成 |

> ⚠️ 当前对外协作层（`/docs`、`/api/public/upload`）是**匿名开放**的，定位适合
> **可信内网或临时场景**；裸暴露公网前须处理 ROADMAP 中的安全项。

---

## 3. 信息架构（功能模块）

```
LumenBrowser
├── 私有文件管理（需登录，多用户）
│   ├── 文件浏览 / 列目录          GET    /api/resources
│   ├── 新建目录 / 上传（小文件）   POST   /api/resources
│   ├── 保存 / 覆盖                 PUT    /api/resources
│   ├── 复制 / 重命名 / 移动        PATCH  /api/resources
│   ├── 删除                        DELETE /api/resources
│   ├── 断点续传（大文件，tus 协议）/api/tus
│   ├── 原始下载 / 打包下载         GET    /api/raw
│   ├── 缩略图 / 预览图             GET    /api/preview
│   ├── 搜索（流式）               GET    /api/search
│   ├── 磁盘用量                    GET    /api/usage
│   ├── 字幕                        GET    /api/subtitle
│   └── 命令执行（WebSocket）       GET    /api/command
├── 账户与权限
│   ├── 登录 / 续期 / 注册          /api/login /api/renew /api/signup
│   ├── 用户管理（管理员）          /api/users
│   └── 全局设置                    /api/settings
├── 分享
│   ├── 创建分享（可设密码/过期）   POST   /api/share
│   ├── 查询某路径的分享            GET    /api/share
│   ├── 删除分享                    DELETE /api/share
│   └── 我的分享列表                GET    /api/shares
└── ★ 对外协作层（二改新增，匿名免登录）
    ├── 公开文档门户               GET /docs, /docs/list, /docs/content, /docs/download
    ├── 免登录投递箱               POST /api/public/upload
    ├── 公开分享内容预览           GET  /api/public/share/{hash}
    └── 公开分享下载               GET  /api/public/dl/{hash}
```

### 二改差异化：对外协作层

经典 filebrowser 全部接口都在 `/api` 鉴权链路上（私有）。LumenBrowser 的差异化在于
新增了一组**不在鉴权链路上、匿名可访问**的对外接口（注册在根 router，而非 `/api` 子路由）：

1. **公开文档门户 `/docs`** —— 把服务器 root 目录当成一份在线文档库只读对外展示。
   左侧目录树 + 右侧预览，支持 Markdown 渲染、代码高亮（highlight.js）、drawio 图（diagrams.net viewer）。
   列表/预览自动跳过点开头文件与敏感名文件（`credentials / secrets / password / .env / id_rsa / id_ed25519 / token`）。
2. **免登录投递箱 `POST /api/public/upload`** —— 任何人 POST 一个文件即可拿到永久分享链接，
   文件落到 `/uploads/` 并加时间戳前缀防冲突。适合收集作业、外部投稿。
3. **增强分享页** —— 分享链接支持 Markdown / 文本在线预览与语法高亮，密码分享流程更顺，不再只能盲下载。

---

## 4. UI 交互设计

### 4.1 视觉规范（Design Tokens）

整体走**纯白 To C 极简风**：只用「白 + 近黑」一个强调色，零彩色、零渐变，把焦点压在产品本身。
以下 token 取自落地页/功能页 mock，作为产品 UI 的统一规范。**该规范已落地**：通过独立覆盖层
`frontend/src/css/lumen-theme.css` 叠加到实际后台界面（侧栏、当前项高亮、存储卡片、文件树面板等），
不改 filebrowser 原始 `base.css`，便于回退与跟踪上游 diff。

| Token | 值 | 用途 |
|---|---|---|
| `--ink` | `#141414` | 主文字 / 强调（**唯一的「重」色**，按钮底色、图标描边、强调文本） |
| `--muted` | `#6b6b6b` | 次要文字 |
| `--faint` | `#9a9a9a` | 更弱文字（面包屑、占位、脚注） |
| `--line` | `#eaeaea` | 分隔线 / 边框 |
| `--panel` | `#fafafa` | 浅面板 / hover 底 |
| `--bg` | `#ffffff` | 纯白背景 |

**排版**
- 字体族：`Inter, -apple-system, BlinkMacSystemFont, "Segoe UI", "PingFang SC", "Microsoft YaHei", system-ui, sans-serif`
- 等宽（代码块）：`"SF Mono", ui-monospace, Menlo, Consolas, monospace`
- 标题：大字重（800），紧字距（`letter-spacing:-.02em ~ -.03em`），H1 用 `clamp(40px,6vw,68px)`
- 正文：`line-height:1.6`，正文 `15~18px`，次要 `13~14.5px`
- 开启字体平滑 `-webkit-font-smoothing:antialiased`

**间距 / 圆角 / 阴影**
- 容器：`max-width:1080px; padding:0 24px`
- 圆角层级：按钮 `10px`、卡片/通用 `14px`（`--radius`）、产品截图框 `16px`、图标容器 `11px`、胶囊 `999px`
- 区块纵向间距：`section` 上下 `80px`，hero `96px / 64px`
- 阴影克制：仅产品截图用一层 `0 30px 60px -30px rgba(20,20,20,.18)`；卡片默认无阴影，hover 仅描边变深 + `translateY(-2px)`
- 导航：`height:64px`，sticky 吸顶，半透明白底 + `backdrop-filter:blur(12px)`，底部 1px 分隔线

**按钮**
- 主按钮 `.btn-primary`：`#141414` 底、白字、`font-weight:600`、hover 上移 1px + 透明度 .92
- 次按钮 `.btn-ghost`：白底、`#eaeaea` 边、hover 边框变 `#141414`
- 交互一律微动（`transition:.15s`），不做花哨动画

**图标**
- 产品图标 `branding/icon.svg`（前端 logo 为 `frontend/public/img/logo.svg`）：圆角方形容器内一个发光光点 +
  透出的光芒，纯单色 `#141414` 描边，呼应「文件之光」。
- 界面 chrome 图标（侧栏、顶栏、面包屑）仍用单色 material-icons，保持极简。
- **文件类型图标（已落地）**：列表内的文件改用一套 **16 个彩色 PNG**
  （`frontend/public/img/file-icons/`，按扩展名映射，详见 [`FEATURES.md`](./FEATURES.md) §6.4），
  让 PDF / Word / 图片 / 代码 等一眼可辨 —— 这是对早期「界面内不引入彩色 icon set」设想的有意调整。

### 4.2 主要界面与交互流程

#### A. 文件浏览（私有，主界面）

- **布局**：左侧固定侧栏 + 右侧主区。侧栏顶部为**用户卡片**，下方为导航项（圆角、hover 浅灰高亮），
  **当前项随路由自动高亮**（`Sidebar.vue` 按 `$route.path` 判断，高亮态为近黑底白字），底部为**卡片化存储用量**。
  （已落地，见 [`FEATURES.md`](./FEATURES.md) §6.2）
- **主区**：顶部面包屑（`root / documents`，当前段用 `--ink` 加粗），下方文件网格（卡片：细边框、
  `10px` 圆角、缩略图/文件类型图标占位 + 文件名单行省略）。两种视图模式：列表 `list` / 网格 `mosaic`（对应用户 `viewMode`）。
- **可选右侧文件树面板**（`FileTree.vue`，已落地）：顶栏一键开关（默认关，状态记忆到 `localStorage`），
  懒加载子目录、进入当前路径自动展开一层、`userTouched` 干预态、每个文件夹显示项目数 badge（折叠也显示）。
  详见 [`FEATURES.md`](./FEATURES.md) §6.3。
- **交互**：
  - 列目录走 `GET /api/resources/<path>`，返回 `Listing`（含 `items / numDirs / numFiles / sorting`）。
  - 单击进入目录 / 打开文件（受用户 `singleClick` 偏好影响）。
  - 排序按 name / size / modified，升降序由用户 `sorting` 决定，前端可切换。
  - 文件按扩展名显示彩色类型图标（`ListingItem.vue` + `file-icons/`，已落地）。
  - **缩略图可关**：设置里 per-user 开关 `disableThumbnails`，关闭后网格用文件图标占位、不请求缩略图（已落地，§6.5）。
  - 文本文件可在内置编辑器中查看/编辑（`Editor.vue`，带 Markdown 预览）。

#### B. 上传

- **小文件**：直接 `POST /api/resources/<path>` 写入（带 `?override=true` 可覆盖）。
- **大文件 / 断点续传**：走 **tus 协议**（`/api/tus`）。
  1. `POST /api/tus/<path>` 带 `Upload-Length` 头创建上传，服务端返回 `201` + `Location`。
  2. `HEAD /api/tus/<path>` 查询已传 `Upload-Offset`。
  3. `PATCH /api/tus/<path>` 分块续传（`Content-Type: application/offset+octet-stream` + `Upload-Offset`）。
  4. 传完自动落盘并触发 upload hook。
  - 交互上表现为带进度条、可暂停/断点续传的上传，刷新/断网后能续。

#### C. 分享生成

- **入口**：在文件/目录上「生成分享」，可选设置密码与有效期。
- **流程**：`POST /api/share/<path>` 带 `{password, expires, unit}`。
  - 服务端生成 **24 字节** 随机 hash（base64-url），抗遍历猜测。
  - 设了密码则额外生成一个 96 字节 `token`（用于带 token 的免密访问）。
  - 返回 `share.Link`（含 `hash / path / expire / userID`）。
- **结果**：得到分享链接 `/share/{hash}`（带预览页）与直链下载。
- **管理**：`GET /api/shares` 看「我的分享」（管理员看全部），`DELETE /api/share/{hash}` 撤销。
- **访问端**：访问者打开分享页，文本/Markdown 内联预览（语法高亮），密码分享先输密码（`X-SHARE-PASSWORD` 头校验）。

#### D. 公开文档门户浏览（对外，匿名）

- **入口**：`GET /docs`，返回一个零依赖单页（HTML 内嵌进二进制）。
- **布局**：左侧目录树 + 右侧内容预览。
- **交互**：
  - 目录树点击 → `GET /docs/list?path=/sub` 拉子目录（JSON，自动过滤敏感文件）。
  - 文件点击 → `GET /docs/content?path=/a.md`（≤10MB）取原始内容 → 前端按类型渲染：
    Markdown 渲染、代码高亮、`.drawio` 用 diagrams.net viewer 嵌入。
  - 「下载」按钮 → `GET /docs/download?path=...` 触发附件下载。
- **气质**：对外只读、干净、无登录摩擦，像一份「在线文档站」而非文件管理器。

#### E. 免登录投递（对外，匿名）

- **入口**：投递箱页面 / 直接 `curl -F "file=@x" /api/public/upload`。
- **流程**：单文件 `multipart/form-data`（字段名 `file`，上限 1GB）→ 落到 `/uploads/{时间戳}_{文件名}` →
  自动建一条**永久**分享 → 返回 JSON：`filename / size / path / share_hash / download_url / browse_url`。
- **交互**：投递者上传后立刻看到「下载链接 + 浏览链接」，零账号、零等待。

---

## 5. CLI 客户端（部分实现）

`lumen` 是 LumenBrowser 的命令行客户端（仓库内独立 Go 模块 `github.com/socake/filebrowser-mods/client`，
第一版刻意零外部依赖、纯标准库，编出最小单二进制）。服务端无需改动，全部走既有 REST API。

| 命令 | 作用 | 背后 API | 状态 |
|---|---|---|---|
| `lumen login` | 登录并保存配置到 `~/.lumen/config.json` | `POST /api/login`（续期 `/api/renew`） | 已实现 |
| `lumen ls` | 列出远程目录 | `GET /api/resources` | 已实现 |
| `lumen put` | 上传（大文件断点续传） | `/api/tus` | 规划 |
| `lumen get` | 下载文件 | `GET /api/raw` | 规划 |
| `lumen rm / mv / mkdir` | 删除 / 移动 / 新建目录 | `DELETE / PATCH / POST /api/resources` | 规划 |
| `lumen share` | 生成分享链 | `POST /api/share` | 规划 |
| `lumen shares` | 列出我的分享 | `GET /api/shares` | 规划 |
| `lumen search` | 搜索文件 | `GET /api/search` | 规划 |
| `lumen drop` | 免登录投递，上传即得永久链 | `POST /api/public/upload` | 已实现 |

> 注：`internal/config`（`~/.lumen/config.json` 读写）与 `internal/api`（login / list /
> public-upload 封装，token 走 `X-Auth` 头）已实现并编译通过，`login / ls / drop` 三命令已接通可用；
> 其余命令（`put/get/rm/mv/share/shares/search`）仍在规划中。命令到 API 的字段映射详见
> [`API-SPEC.md`](./API-SPEC.md) 的 CLI 一节。

---

## 6. 产品路线图概要

结合 [`ROADMAP.md`](./ROADMAP.md)，当前状态为「能用，但对外协作层匿名开放，仅适合可信内网/临时场景」。
上公网前的优先级如下：

### 🔴 必修（安全）

1. **`/docs/download` 未接敏感过滤** —— list/preview 会过滤 `.env`、`id_rsa`、`credentials` 等，
   但 `docsDownloadHandler` 没调 `hiddenOrSensitive`，知道路径即可直接下载敏感文件。
2. **路径越界的 sibling-prefix bug** —— `strings.HasPrefix(realPath, realRoot)` 会把
   `/home/ubuntu-secret` 误判为在 `/home/ubuntu` 内；三个 docs handler 都受影响。
3. **投递箱无鉴权 / 无限速 / 无类型校验** —— `POST /api/public/upload` 单文件 1GB 且生成永久分享，
   任何人可填满磁盘、传恶意文件、刷爆分享表。需加上传凭证 / 限速 / IP 配额 / 类型白名单 / 默认过期。
4. **CSP 引入 `unsafe-eval` + 外链 CDN** —— 为 diagrams.net viewer 放开了 `unsafe-eval` 并信任
   cdnjs / viewer.diagrams.net，有 XSS 面。建议把渲染库本地化打包、收紧 CSP。

### 🟡 健壮性

5. 硬编码 admin = user 1（所有匿名 handler 用 `Users.Get(root, 1)` 取管理员），应改成可配置的「门户暴露用户」。
6. 永久分享无清理（投递产生 `Expire:0` 分享 + `/uploads` 只增不减），需加 TTL + 定期清理。
7. 门户 UI 内嵌成数百行 Go 字符串，难维护，应改 `//go:embed` 模板文件。

### 🟢 想做

- 门户全文搜索、目录级对外权限（配置驱动而非硬编码黑名单）、更多文件类型在线预览（PDF/图库/diff）、
  投递后自动转 Markdown 预览，以及把二改整理成可向上游提 PR 的形态。
</content>
</invoke>
