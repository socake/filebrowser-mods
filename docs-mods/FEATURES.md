# 二改功能说明

基于 [filebrowser](https://github.com/filebrowser/filebrowser) v2.62.2（MIT License）的个人定制。
原版是一个带登录的私有文件管理器，本分支在其之上增加了**对外的文档门户**、**公开投递箱**，并增强了分享页与 Markdown 预览。

> 分支约定：`master` = 纯净上游 v2.62.2；`my-mods` = 本人二改。两者 diff 即全部改动。

---

## 1. 公开文档门户 `/docs`

一个**无需登录**的只读门户，把服务器 root 目录当成一份在线文档库对外展示。

- 访问 `http://<host>/docs` 打开单页门户（内嵌 HTML，零额外依赖）
- 左侧目录树 + 右侧内容预览，支持代码高亮（highlight.js）、Markdown 渲染、drawio 图（diagrams.net viewer）
- 文件可在线预览或下载

**接口**
| 路由 | 方法 | 作用 |
|------|------|------|
| `/docs` | GET | 返回门户页面（HTML） |
| `/docs/list?path=/` | GET | 列目录（JSON） |
| `/docs/content?path=/a.md` | GET | 取文件原始内容（预览用，≤10MB） |
| `/docs/download?path=/a.md` | GET | 触发下载 |

**敏感文件过滤**：列表与预览会跳过点开头文件，以及名字含 `credentials / secrets / password / .env / id_rsa / id_ed25519 / token` 的文件。

> ⚠️ 已知缺口：下载接口 `/docs/download` 当前**未**接同一套过滤，详见 [ROADMAP.md](./ROADMAP.md)。

---

## 2. 公开上传投递箱 `POST /api/public/upload`

一个**无需登录**的文件投递接口，任何人都能上传文件并立即拿到一个永久分享链接。

- `multipart/form-data`，字段名 `file`，单文件上限 1GB
- 文件落到 root 下的 `/uploads/`，文件名加时间戳前缀防冲突
- 自动生成一条**永久**分享记录（不过期），返回下载/浏览链接

**返回示例**
```json
{
  "filename": "report.pdf",
  "size": 102400,
  "path": "/uploads/20260612-103000_report.pdf",
  "share_hash": "Ab3xYz",
  "download_url": "/api/public/dl/Ab3xYz",
  "browse_url": "/share/Ab3xYz"
}
```

**用法**
```bash
curl -F "file=@./report.pdf" http://<host>/api/public/upload
```

> ⚠️ 这是开放投递箱（无鉴权、无限速、永久分享），仅适合可信内网/临时场景，详见 [ROADMAP.md](./ROADMAP.md)。

---

## 3. 分享页增强（Share.vue）

- 分享页支持 Markdown / 文本在线预览，带语法高亮，不再只能下载
- 密码分享的交互优化（错误密码提示、token 流程）
- 后端 `http/public.go` 在分享取文件时返回文件内容（`Content: true`），前端才能内联预览

## 4. Markdown 预览样式（mdPreview.css）+ 编辑器（Editor.vue）

- 一套完整的 Markdown 渲染样式（标题/表格/代码块/引用等约 285 行）
- 编辑器视图增强，编辑时可预览

## 5. 安全增量

- 分享 hash 熵从 6 字节提升到 **24 字节**（`http/share.go`），显著降低被遍历/猜测的概率
- CSP 策略放开到允许 highlight.js（cdnjs）与 diagrams.net viewer（`http/http.go`），代价见 ROADMAP

---

## 构建

```bash
# 前端
cd frontend && pnpm install && pnpm build
# 后端（产物会内嵌前端 dist）
cd .. && go build -o filebrowser
# 运行
./filebrowser --address 0.0.0.0 --port 8080 --root <要暴露的目录> --database ./filebrowser.db
```
