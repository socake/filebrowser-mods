# 二改功能说明

基于 [filebrowser](https://github.com/filebrowser/filebrowser) v2.62.2（Apache-2.0 License）的个人定制。
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
  "share_hash": "k3Jq8vR2mNpX7wZ1aB4cD6eF9gH0iJ5L",
  "download_url": "/api/public/dl/k3Jq8vR2mNpX7wZ1aB4cD6eF9gH0iJ5L",
  "browse_url": "/share/k3Jq8vR2mNpX7wZ1aB4cD6eF9gH0iJ5L"
}
```

> 投递箱的 `share_hash` 用 24 字节随机熵（base64url 编码，约 32 字符），已与 `POST /api/share`
> 的分享接口（`http/share.go`）保持一致，不再是早期的 6 字节短 hash。

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

## 6. LumenBrowser 产品化 UI 增强

把原版偏「后台工具」的界面，重做成一套纯白 To C 体验。**所有视觉改动以独立覆盖层
`frontend/src/css/lumen-theme.css` 叠加**，不修改 filebrowser 原始 `base.css` / `styles.css`，
便于回退与跟踪上游 diff。

### 6.1 品牌化

- 产品名 **LumenBrowser**（`frontend/src/utils/constants.ts` 里 `name` 兜底为 `LumenBrowser`）。
- logo / favicon 已替换：`frontend/public/img/logo.svg`（前端引用，`logoURL = .../img/logo.svg`）、
  `frontend/public/img/icons/favicon.ico`；品牌素材源文件在仓库根 `branding/`（`logo.svg/png`、
  `icon.svg/png`、`banner.svg/png`）。
- 主色从原版蓝改为**近黑 `#141414` + 纯白**的极简风（零彩色、零渐变），按钮底色 / 当前项高亮 / 强调文本统一用 `#141414`。

### 6.2 侧栏改造（`lumen-theme.css` + `components/Sidebar.vue`）

- 顶部**用户卡片化**（`nav .sb-user`：浅灰底、圆角、用户名加粗）。
- 导航项圆角（`10px`）、舒适内距、图标文字对齐、hover 浅灰高亮；登出项 hover 用警示红。
- **当前项随路由自动高亮**：`Sidebar.vue` 用 `:class="{ 'sb-active': $route.path.startsWith('/files') }"`
  （设置项同理判断 `/settings`），高亮态 `nav .sb-active` 为近黑底白字。
- 底部存储用量**卡片化**（`nav .credits` 覆盖原 inline 样式）。

### 6.3 右侧文件树面板（`frontend/src/components/FileTree.vue`）

一个可选的右侧目录树侧栏，区别于左侧导航：

- **懒加载**：只在展开某目录时才 `api.fetch(url)` 拉取该层子项（`loadChildren`），目录在前、文件在后并按名排序。
- **当前路径自动展开一层**：导航变化时 `autoExpand()` 把当前目录（或文件的父目录）的各级祖先展开，
  恰好露出当前层的直接子项。
- **`userTouched` 干预态**：用户手动展开/折叠过的节点记入 `userTouched`，自动展开逻辑**不再覆盖**这些节点，
  避免「点开又被自动收起」的对抗。
- **每个文件夹显示项目数 badge**：`childCount()` 给目录右侧渲染数字徽标；通过 `loadCountOnly()`
  **预取当前层每个子文件夹的项目数**，使其**折叠时也能直接显示数字**（只预取一层、不深挖）。
- **顶栏 toggle 开关**：`components/header/HeaderBar.vue` 的 `#filetree-toggle` 一键开关面板；
  状态由 `stores/layout.ts` 管理并持久化到 `localStorage`（键 `lumen-filetree-open`），
  **默认关闭**（首次为 `false`）。

### 6.4 文件类型彩色图标（`frontend/public/img/file-icons/` + `components/files/ListingItem.vue`）

- `file-icons/` 下 **16 个彩色 PNG**：`pdf / word / excel / powerpoint / markdown / txt / csv /
  image / video / audio / archive / code / json / folder / folder-shared / generic`。
- `ListingItem.vue` 里 `EXT_ICON` 把扩展名映射到图标名（如 `docx→word`、`mp4→video`、`zip→archive`、
  `py/go/ts→code`），目录用 `folder.png`，未命中的扩展名兜底 `generic.png`。
- 说明：`folder-shared.png` 已随包提供，但当前 `fileIconUrl` 对目录统一用 `folder.png`，
  **`folder-shared` 暂未接入映射**（预留素材）。

### 6.5 图片缩略图预览开关（per-user，`disableThumbnails`）

- 设置页 `views/settings/Profile.vue` 提供「加载缩略图」开关，写回用户字段 `disableThumbnails`
  （`types/user.d.ts`）。
- `ListingItem.vue` 的 `isThumbsEnabled = enableThumbs && !user.disableThumbnails`：**关闭后网格
  直接用文件类型图标占位、不再请求 `/api/preview` 缩略图**，弱网或大目录更轻快。

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
