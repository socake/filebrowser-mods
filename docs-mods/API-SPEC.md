# LumenBrowser 接口设计规范

> 本规范基于二改分支 `my-mods`（filebrowser v2.62.2 衍生）真实源码整理。路由总表见
> [`http/http.go`](../http/http.go)，鉴权见 [`http/auth.go`](../http/auth.go)。
> 字段一律对照源码；个别未在源码中确证的点标注「待确认」。

---

## 1. 通用约定

### 1.1 Base URL

```
http(s)://<host>[:<port>]<BaseURL>
```

- 默认 `BaseURL` 为空，即 API 根为 `/api`。若部署时配置了 `server.BaseURL`，所有路径前缀会带上它
  （服务端用 `stripPrefix(server.BaseURL, r)` 统一剥离）。
- 私有 API 都在 `/api/*` 下；对外协作层有两类：根级 `/docs*` 与 `/api/public/*`。

### 1.2 鉴权方式（重要）

**鉴权用 JWT，token 通过 HTTP 头 `X-Auth` 传递。**（源码：`http/auth.go` 的
`extractor.ExtractToken` 使用 `request.HeaderExtractor{"X-Auth"}`。）

- 标准用法：在请求头加 `X-Auth: <jwt>`。
- 兼容用法：对 **GET** 请求，也接受名为 `auth` 的 Cookie 携带同一 JWT（仅 GET，供浏览器直链场景）。
- token 必须是包含两个 `.` 的合法 JWT（HS256 签名），否则视为无 token。
- **登录响应不是 JSON**：`POST /api/login` 成功时直接把签名后的 JWT 以
  `Content-Type: text/plain` 写进 **响应 body**（源码 `printToken`：`w.Write([]byte(signed))`）。
  客户端应把整个响应体当作 token 字符串，而非解析 JSON 取字段。

> JWT claims 内含用户信息（见 [§3.1 登录响应](#31-登录)），但 LumenBrowser 把它放在 token 的
> payload 里，HTTP 响应体本身只有 token 文本。

### 1.3 Token 有效期与续期

- 默认有效期 `DefaultTokenExpirationTime = 2h`（可被 `server.GetTokenExpirationTime` 覆盖）。
- 当 token 距过期 < 1 小时，或用户信息在签发后被更新过，受保护接口会在响应里带上
  `X-Renew-Token: true` 头，提示客户端去续期。
- 续期：`POST /api/renew`（需带当前有效 token），返回新 token（同样是 text/plain 的 JWT body），
  并设 `X-Renew-Token: false`。

### 1.4 错误码约定

服务端 handler 返回 `(statusCode, error)`，框架据此写状态码。常见语义：

| 状态码 | 含义 |
|---|---|
| `200 OK` | 成功（多数 GET / 操作成功） |
| `201 Created` | 资源创建成功（如新建用户、tus 创建上传） |
| `204 No Content` | 成功且无响应体（删除、PATCH 移动、tus PATCH 分块） |
| `400 Bad Request` | 请求参数 / body 非法 |
| `401 Unauthorized` | 未登录 / token 无效；或分享密码错误 |
| `403 Forbidden` | 权限不足 / 访问被拒（如越权、敏感路径） |
| `404 Not Found` | 资源不存在 |
| `405 Method Not Allowed` | 方法不允许（如对目录 PUT、注册关闭时 signup） |
| `409 Conflict` | 冲突（文件已存在且未 override、用户名已存在、tus offset 不匹配） |
| `413 Request Entity Too Large` | 文档门户预览文件超过 10MB |
| `415 Unsupported Media Type` | tus PATCH 的 Content-Type 不对 |
| `500 Internal Server Error` | 服务端错误 |

### 1.5 公开 / 二改标注

- 🔓 **公开（免鉴权）**：`/docs*`、`/api/public/*`、`/api/login`、`/api/signup`、`/health`。
- ★ **二改新增**：`/docs*`、`/api/public/upload`（投递箱）。分享 hash 熵提升、分享内容回传为前端预览
  也属二改增强。其余为上游 filebrowser 既有接口。

---

## 2. 路由总表

| 方法 | 路径 | 鉴权 | 说明 |
|---|---|---|---|
| GET | `/health` | 🔓 | 健康检查 |
| POST | `/api/login` | 🔓 | 登录，返回 JWT（text/plain） |
| POST | `/api/signup` | 🔓 | 注册（需服务端开启 Signup） |
| POST | `/api/renew` | ✅ | 续期 token |
| GET | `/api/users` | ✅(admin) | 用户列表 |
| POST | `/api/users` | ✅(admin) | 新建用户 |
| GET/PUT/DELETE | `/api/users/{id}` | ✅(self/admin) | 用户读 / 改 / 删 |
| GET | `/api/resources/<path>` | ✅ | 列目录 / 取文件信息 |
| POST | `/api/resources/<path>` | ✅ | 新建目录（路径以 `/` 结尾）/ 上传文件 |
| PUT | `/api/resources/<path>` | ✅ | 保存（覆盖已存在文件） |
| PATCH | `/api/resources/<path>` | ✅ | 复制 / 重命名 / 移动 |
| DELETE | `/api/resources/<path>` | ✅ | 删除 |
| POST/HEAD/GET/PATCH/DELETE | `/api/tus/<path>` | ✅ | 断点续传（tus 协议） |
| GET | `/api/raw/<path>` | ✅ | 原始下载 / 打包下载 |
| GET | `/api/preview/{size}/{path}` | ✅ | 缩略图 / 预览图（size=thumb\|big） |
| GET | `/api/search/<path>?query=` | ✅ | 搜索（流式 NDJSON） |
| GET | `/api/usage/<path>` | ✅ | 磁盘用量 |
| GET | `/api/subtitle/<path>` | ✅ | 字幕 |
| GET | `/api/command/...` | ✅ | 命令执行（WebSocket） |
| GET | `/api/shares` | ✅ | 我的分享列表（admin 看全部） |
| GET | `/api/share/<path>` | ✅ | 某路径的分享 |
| POST | `/api/share/<path>` | ✅ | 创建分享 |
| DELETE | `/api/share/{hash}` | ✅ | 删除分享 |
| GET/PUT | `/api/settings` | ✅ | 全局设置读 / 写 |
| GET | `/docs` | 🔓 ★ | 文档门户页面（HTML） |
| GET | `/docs/list?path=` | 🔓 ★ | 门户列目录（JSON） |
| GET | `/docs/content?path=` | 🔓 ★ | 门户取文件内容（≤10MB） |
| GET | `/docs/download?path=` | 🔓 ★ | 门户下载 |
| POST | `/api/public/upload` | 🔓 ★ | 免登录投递箱 |
| GET | `/api/public/share/{hash}[/file]` | 🔓 | 公开分享内容预览（JSON） |
| GET | `/api/public/dl/{hash}[/file]` | 🔓 | 公开分享下载 |

---

## 3. 账户与鉴权

### 3.1 登录

```
POST /api/login
```

**Body**（JSON，由配置的 auth 后端解析；JSON auth 下为）：

```json
{ "username": "alice", "password": "secret" }
```

**响应**：`200 OK`，`Content-Type: text/plain`，body 为 JWT 字符串本身：

```
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyIjp7ImlkIjox...
```

该 JWT 的 payload（解码后）含用户信息，字段如下（源码 `userInfo`）：

```json
{
  "user": {
    "id": 1,
    "locale": "zh-cn",
    "viewMode": "list",
    "singleClick": false,
    "redirectAfterCopyMove": false,
    "perm": { "admin": true, "execute": true, "create": true, "rename": true,
              "modify": true, "delete": true, "share": true, "download": true },
    "commands": [],
    "lockPassword": false,
    "hideDotfiles": false,
    "dateFormat": false,
    "username": "alice",
    "aceEditorTheme": ""
  },
  "iss": "File Browser",
  "iat": 1718900000,
  "exp": 1718907200
}
```

- 失败：密码错误 → `403`；其它 → `500`。

### 3.2 续期

```
POST /api/renew
X-Auth: <当前 JWT>
```

响应同登录：`200` + text/plain 新 JWT，并设响应头 `X-Renew-Token: false`。

### 3.3 注册（可选）

```
POST /api/signup
Content-Type: application/json

{ "username": "bob", "password": "secret" }
```

- 仅当服务端 `settings.Signup` 开启时可用，否则 `405`。
- 自注册用户强制非管理员、无执行权限、`commands` 为空，并自动建 home 目录。
- 成功 `200`；用户名已存在 `409`；参数缺失 `400`。

### 3.4 用户管理

| 方法 | 路径 | 权限 | 说明 |
|---|---|---|---|
| GET | `/api/users` | admin | 列出所有用户（密码字段清空） |
| POST | `/api/users` | admin | 新建用户，成功 `201` + `Location: /settings/users/{id}` |
| GET | `/api/users/{id}` | 本人或 admin | 取用户（非 admin 看不到 scope） |
| PUT | `/api/users/{id}` | 本人或 admin | 改用户，body 见下 |
| DELETE | `/api/users/{id}` | 本人或 admin | 删用户 |

PUT/POST body 结构（`modifyUserRequest`）：

```json
{
  "what": "user",
  "which": ["password"],
  "current_password": "...",
  "data": { "id": 2, "username": "bob", "password": "...", "perm": { } }
}
```

- `which` 指定要改的字段（空数组或 `["all"]` 表示整体替换，需 admin）。
- JSON auth 下改敏感字段（all/username/password/scope/lockPassword/commands/perm）需带 `current_password`。
- 非 admin 不可改：`Username / Scope / LockPassword / Perm / Commands / Rules`。

---

## 4. 资源（文件管理）

资源路径直接拼在 URL 上，例如 `/api/resources/documents/季度报告.md`（需 URL 编码）。
全部需鉴权，且受用户权限位（`create/modify/rename/delete/download/share`）与 rules 约束。

### 4.1 列目录 / 取文件信息

```
GET /api/resources/<path>
```

- **目录** → 返回 `FileInfo`（含内嵌 `Listing`）：

```json
{
  "path": "/documents",
  "name": "documents",
  "size": 4096,
  "extension": "",
  "modified": "2026-06-12T10:30:00Z",
  "mode": 2147484141,
  "isDir": true,
  "isSymlink": false,
  "type": "",
  "items": [
    {
      "path": "/documents/季度报告.md",
      "name": "季度报告.md",
      "size": 10240,
      "extension": ".md",
      "modified": "2026-06-12T10:30:00Z",
      "mode": 420,
      "isDir": false,
      "isSymlink": false,
      "type": "text"
    }
  ],
  "numDirs": 1,
  "numFiles": 5,
  "sorting": { "by": "name", "asc": true }
}
```

- **文件** → 返回单个 `FileInfo`（不含 `items`）。文本文件会带 `content` 字段（当 `Content:true`）。
  - 可选 `?checksum=md5|sha1|sha256|sha512` → 返回 `checksums` 映射，并清空 `content`。
  - 请求头 `X-Encoding: true` 且文件为文本类型 → 直接返回原始字节（`application/octet-stream`），不包 JSON。

> `FileInfo` 完整字段（源码 `files/file.go`）：`path, name, size, extension, modified, mode, isDir,
> isSymlink, type, subtitles?, content?, checksums?, token?, resolution?` + 目录时的 `items, numDirs,
> numFiles, sorting`。

### 4.2 新建目录 / 上传小文件

```
POST /api/resources/<path>
```

- 路径以 `/` 结尾 → `MkdirAll` 创建目录。
- 否则把请求 body 当文件内容写入；文件已存在需 `?override=true`（且有 modify 权限），否则 `409`。
- 成功响应头带 `ETag`。需 `create` 权限。

### 4.3 保存（覆盖文件）

```
PUT /api/resources/<path>
```

- 仅对已存在文件；目录 `405`；文件不存在 `404`。需 `modify` 权限。响应带 `ETag`。

### 4.4 复制 / 重命名 / 移动

```
PATCH /api/resources/<src>?action=<copy|rename>&destination=<dst>[&override=true|&rename=true]
```

- `action=copy` 需 `create`；`action=rename`（即移动/改名）需 `rename`。
- 目标已存在且未带 `override`/`rename` → `409`；`rename=true` 时自动加 `(n)` 后缀避免覆盖。
- 成功 `204`。

### 4.5 删除

```
DELETE /api/resources/<path>
```

- 需 `delete` 权限；不能删根。会一并删除该路径前缀下的分享记录与缩略图。成功 `204`。

### 4.6 磁盘用量

```
GET /api/usage/<path>
```

```json
{ "total": 500107862016, "used": 123456789012 }
```

- 仅对目录有意义；非目录返回 `{ "total":0, "used":0 }`。

---

## 5. tus 断点续传

大文件走 tus 协议，端点 `/api/tus/<path>`，需 `create` 权限。

| 步骤 | 方法 | 关键头 | 响应 |
|---|---|---|---|
| 创建上传 | POST | `Upload-Length: <总字节>` | `201` + `Location: <base>/api/tus/<path>` |
| 查询进度 | HEAD | — | `200` + `Upload-Offset` / `Upload-Length` |
| 续传分块 | PATCH | `Content-Type: application/offset+octet-stream`、`Upload-Offset: <当前偏移>`，body 为该块字节 | `204` + 新 `Upload-Offset` |
| 取消 | DELETE | — | `204` |

- 文件已存在需 `?override=true`（否则 POST 返回 `409`）；override 还需 `modify` 权限。
- PATCH 的 `Upload-Offset` 必须等于当前文件大小，否则 `409`；Content-Type 不符 `415`。
- 当累计 offset ≥ `Upload-Length`，服务端标记完成并触发 upload hook。

---

## 6. 下载 / 预览

### 6.1 原始下载 / 打包下载

```
GET /api/raw/<path>?[files=a,b]&[algo=zip|tar|targz|...]&[inline=true]
```

- 需 `download` 权限。
- 单文件 → 直接回传文件流，`Content-Disposition: attachment`（`inline=true` 则 `inline`）。
- 目录或多文件 → 打包下载。`algo` 支持：`zip`(默认)、`tar`、`targz`、`tarbz2`、`tarxz`、`tarlz4`、
  `tarsz`、`tarbr`、`tarzst`。`files=` 用逗号分隔相对该路径的多个条目。

### 6.2 预览图 / 缩略图

```
GET /api/preview/{size}/{path}
```

- `size`：`thumb` | `big`。返回图片二进制。受服务端 `EnableThumbnails / ResizePreview` 控制。

### 6.3 字幕

```
GET /api/subtitle/<path>
```

---

## 7. 搜索

```
GET /api/search/<path>?query=<关键词>
```

- 需鉴权。**响应是流式 NDJSON**（不是单个 JSON 数组）：每命中一项写一行 JSON，行间以 `\n` 分隔，
  长连接期间每 5 秒发一个空心跳包保活。

每行结构：

```json
{"dir": false, "path": "documents/季度报告.md"}
```

- 客户端需按行解析，读到 EOF 即结束；空行为心跳，应忽略。
- query 语法支持上游 filebrowser 的类型过滤（如 `type:image`）等，详见 search 包（此处略，**待确认**具体语法面）。

---

## 8. 分享

### 8.1 创建分享 ★（hash 熵增强）

```
POST /api/share/<path>
Content-Type: application/json

{ "password": "可选", "expires": "7", "unit": "days" }
```

- 需 `share` 且 `download` 权限。
- `expires` 为数字字符串，`unit` ∈ `seconds|minutes|hours(默认)|days`；不传则永久（`expire=0`）。
- hash 为 **24 字节** 随机值（base64-url），显著抗遍历。设了密码则额外生成 96 字节 `token`。

**响应**（`share.Link`）：

```json
{
  "hash": "Ab3xYz...（base64url，约32字符）",
  "path": "/documents/季度报告.md",
  "userID": 1,
  "expire": 1719500000,
  "password_hash": "$2a$...（设密码时）",
  "token": "...（设密码时，96字节base64url）"
}
```

### 8.2 查询某路径的分享

```
GET /api/share/<path>
```

返回该路径下属于当前用户的 `share.Link` 数组（无则 `[]`）。

### 8.3 我的分享列表

```
GET /api/shares
```

- 需 `share`+`download` 权限。admin 返回全部分享，普通用户仅自己的。
- 过期分享在读取时会被惰性清理。按 `userID`、`expire` 排序。

### 8.4 删除分享

```
DELETE /api/share/{hash}
```

- 仅分享所有者或 admin 可删。成功按 `errToStatus`（通常 `200`）。

---

## 9. 对外协作层（二改新增 ★，匿名免登录）

### 9.1 公开文档门户

> 这组路由注册在**根 router**，不在 `/api` 鉴权链上，**匿名可访问**。内部用硬编码的 admin
> （`Users.Get(root, 1)`）取文件系统句柄。

#### 9.1.1 门户页面

```
GET /docs
```

返回内嵌的单页 HTML（`Content-Type: text/html`）。

#### 9.1.2 列目录

```
GET /docs/list?path=/sub
```

```json
{
  "path": "/sub",
  "entries": [
    { "name": "guide.md", "path": "/sub/guide.md", "isDir": false,
      "size": 2048, "modTime": "2026-06-12T10:30:00Z", "ext": ".md" }
  ]
}
```

- `path` 缺省为 `/`。目录优先、再按名排序。
- **自动跳过**点开头文件与敏感名文件（名字含 `credentials/secrets/password/.env/id_rsa/id_ed25519/token`）。
- 路径越界 → `403`；非目录 → `400`；不存在 → `404`。

#### 9.1.3 取文件内容（预览）

```
GET /docs/content?path=/a.md
```

- 返回原始文件内容（`text/plain`），附响应头 `X-File-Name / X-File-Size / X-File-ModTime`。
- 同样过滤敏感文件（命中 → `403`）；目录 → `400`；> 10MB → `413`。

#### 9.1.4 下载

```
GET /docs/download?path=/a.md
```

- `Content-Disposition: attachment` 触发下载。
- ⚠️ **已知缺口**：此 handler **未**调用敏感文件过滤，知道路径即可下载 `.env` 等（见 ROADMAP 必修项 1）。

### 9.2 免登录投递箱 ★

```
POST /api/public/upload
Content-Type: multipart/form-data
```

- 表单字段名 `file`，单文件上限 **1GB**。匿名、无限速、无类型校验。
- 文件落到 `/uploads/{时间戳}_{原文件名}`（碰撞再加 `_1` 后缀），并自动建一条**永久**分享。

**响应**（`200`，`application/json`）：

```json
{
  "filename": "report.pdf",
  "size": 102400,
  "path": "/uploads/20260612-103000_report.pdf",
  "share_hash": "k3Jq8vR2mNpX7wZ1aB4cD6eF9gH0iJ5L",
  "download_url": "/api/public/dl/k3Jq8vR2mNpX7wZ1aB4cD6eF9gH0iJ5L",
  "browse_url": "/share/k3Jq8vR2mNpX7wZ1aB4cD6eF9gH0iJ5L"
}
```

```bash
curl -F "file=@./report.pdf" http://<host>/api/public/upload
```

> 投递箱生成的分享 hash 为 **24 字节**（base64url 约 32 字符，`http/public_upload.go`），
> 已与 `POST /api/share` 对齐，早期 6 字节熵偏低的不一致点已消除。

### 9.3 公开分享访问

分享 hash 对应的内容/下载入口（供分享页与直链使用）：

```
GET /api/public/share/{hash}[/可选文件名]   # 返回 FileInfo（JSON，含 content 供内联预览）
GET /api/public/dl/{hash}[/可选文件名]       # 下载（文件直传 / 目录打包）
```

- 若分享设了密码：可用 `?token=<分享token>` 免密，或请求头 `X-SHARE-PASSWORD: <密码>` 校验；
  密码错误 → `401`。
- 路径尾部可附文件名（如 `/api/public/dl/{hash}/report.pdf`），方便旧浏览器保存正确文件名。
- 分享为目录时，`/api/public/share/{hash}/子路径` 可浏览子目录。

---

## 10. CLI 客户端 ↔ API 映射

`lumen`（`github.com/socake/filebrowser-mods/client`）每个命令对应的 API 与调用要点：

| 命令 | API | 调用要点 |
|---|---|---|
| `lumen login -url -u -p` | `POST /api/login` | body `{username,password}`；**响应体即 JWT 文本**，原样保存到 `~/.lumen/config.json` 的 `token`，记录 `server/username/expires` |
| （续期） | `POST /api/renew` | 带 `X-Auth: <token>`；响应体为新 JWT；可在 token 临近过期或收到 `X-Renew-Token:true` 时触发 |
| `lumen ls <path>` | `GET /api/resources/<path>` | 带 `X-Auth`；解析返回 `FileInfo.items`，打印 `name/size/isDir` |
| `lumen put <local> <remote>` | `/api/tus/<remote>` | tus 三步：POST(`Upload-Length`) → 循环 PATCH(`Upload-Offset` + `application/offset+octet-stream`)；支持断点续传 |
| `lumen get <remote> <local>` | `GET /api/raw/<remote>` | 带 `X-Auth`；目录可加 `?algo=zip` 打包 |
| `lumen rm <remote>` | `DELETE /api/resources/<remote>` | 需 delete 权限 |
| `lumen mv <src> <dst>` | `PATCH /api/resources/<src>?action=rename&destination=<dst>` | 需 rename 权限 |
| `lumen mkdir <remote>/` | `POST /api/resources/<remote>/` | 路径以 `/` 结尾 |
| `lumen share <remote> [--password --expires]` | `POST /api/share/<remote>` | body `{password,expires,unit}`；打印返回的 `hash` 拼成 `/share/{hash}` |
| `lumen shares` | `GET /api/shares` | 列出 `share.Link[]` |
| `lumen search <kw>` | `GET /api/search/<path>?query=<kw>` | **按行解析 NDJSON 流**，忽略空心跳行 |
| `lumen drop <file> [--server]` | `POST /api/public/upload` | **免鉴权**；multipart 字段 `file`；解析返回 `download_url/browse_url` 打印 |

通用要点：
- 鉴权统一加请求头 `X-Auth: <token>`（GET 也可用 `auth` cookie，但 CLI 建议统一用头）。
- token 是纯文本 JWT，无需 `Bearer ` 前缀。
- 配置文件 `~/.lumen/config.json`：`{ server, username, token, expires }`。

---

## 11. 读源码发现的与既有描述不符之处

1. **登录响应格式**：`POST /api/login` 返回的不是 JSON，而是 **text/plain 的裸 JWT 字符串**
   （`printToken` 直接 `w.Write([]byte(signed))`）。用户信息编码在 JWT payload 里，不是顶层 JSON 字段。
   客户端应把整个响应体当 token。
2. **鉴权头**：确认是自定义头 **`X-Auth`**（非 `Authorization: Bearer`），GET 额外兼容 `auth` Cookie。
3. **分享 hash 熵已统一**（曾不一致）：早期 `POST /api/share`（`http/share.go`）用 24 字节，而投递箱
   `POST /api/public/upload` 只用 6 字节。现已把后者改为 `make([]byte, 24)`，两条路径熵一致、均抗遍历，
   此不一致点已消除。
4. **搜索是流式 NDJSON**，不是普通 JSON 数组；含 5 秒空心跳包。CLI/前端需按行解析。
5. **`/docs/download` 未接敏感过滤**（已在 ROADMAP 记录），与 list/content 的过滤行为不一致，属安全缺口。
6. **CLI 部分实现**：`internal/config`（`~/.lumen/config.json` 读写）与 `internal/api`（login / list /
   public-upload 封装，token 走 `X-Auth` 头）已落地并编译通过，`login.go/ls.go/drop.go` 三命令已接通可用；
   `put/get/rm/mv/share/shares/search` 仍为规划中。本规范其余未实现命令的映射为目标设计而非现状。
</content>
