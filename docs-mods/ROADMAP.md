# 后期开发计划 & 暂未完善的点

诚实记录当前已知的问题与未来想做的事。**当前状态：能用，但 `/docs` 与公开上传都是匿名开放，
只适合可信内网或临时场景，不要直接裸暴露公网。**

---

## 🔴 必修（安全，上公网前务必处理）

1. ✅ **已修复：`/docs/download` 绕过敏感文件过滤**
   原 `docsDownloadHandler`（`http/docs.go`）没有调用 `hiddenOrSensitive`，
   导致虽然 list/preview 看不到 `.env`、`id_rsa`、`credentials` 等文件，
   但只要知道路径就能直接 `GET /docs/download?path=/.env` 下载。
   现已在 download 前加上 `hiddenOrSensitive(filepath.Base(reqPath))` 检查，命中返回 403。
   实测：`/docs/download?path=/secrets.txt` 与 `/.env` 均返回 403，普通文件 `readme.txt` 正常 200 下载。

2. ✅ **已修复：路径越界判断的 sibling-prefix bug**
   原 `strings.HasPrefix(realPath, realRoot)` 会把 `/home/ubuntu-secret` 误判为在 `/home/ubuntu` 内。
   现抽出 `pathWithinRoot(realPath, realRoot)` 助手：`realPath == realRoot ||
   strings.HasPrefix(realPath, realRoot + string(os.PathSeparator))`，
   三个 handler（list/content/download）已全部改用。

3. **公开上传是无鉴权开放投递箱**（部分修复）
   `POST /api/public/upload` 原本无 auth、无限速、无类型校验，单文件 1GB，且生成**永久**分享。
   - ✅ **已修复：投递箱分享 hash 熵不足**——原 `make([]byte, 6)`（约 8 字符）熵偏低、易被遍历，
     现已改为 `make([]byte, 24)`，与 `http/share.go` 对齐。
   - ✅ **已修复：无限速**——加了进程内固定窗口限速（map+mutex），默认每个来源 IP 每分钟 10 次，
     超限返回 429（参数 `publicUploadRateLimit=10` / `publicUploadWindow=1min`，见 `http/public_upload.go`）。
     保留"免登录投递"卖点，默认仍无需登录。
   - ✅ **已修复：可选鉴权开关**——支持环境变量 `FB_PUBLIC_UPLOAD_TOKEN`：设置后要求请求带该口令
     （`X-Upload-Token` 头 / `Authorization: Bearer <token>` / `?token=` 查询参数任一），
     校验用常量时间比较，缺失或错误返回 401；不设置则维持免登录。
   - ⬜ **仍待处理**：文件类型白名单、单 IP 磁盘配额、分享默认带过期（`Expire:0` 永久，见第 6 点）。

4. **CSP 引入 `unsafe-eval` + 外链 CDN**
   为了 diagrams.net viewer 放开了 `unsafe-eval`，并信任 cdnjs / viewer.diagrams.net。
   → 选项：把 highlight.js / 渲染库本地化打包，去掉外链与 `unsafe-eval`，收紧 CSP。

## 🟡 健壮性

5. **硬编码 admin = user 1**
   所有匿名 handler 用 `Users.Get(root, uint(1))` 取管理员。换个部署/多管理员就不成立。
   → 改成可配置的「门户暴露用户」或显式的 portal 配置项。

6. **永久分享无清理**
   公开上传产生的分享 `Expire:0` 永不过期，`/uploads` 也只增不减。
   → 加 TTL + 定期清理任务（cron / 启动时扫）。

7. **门户 UI 内嵌成 513 行 Go 字符串**
   `docs_page.go` 难维护、无法热改。
   → 改用 `//go:embed` 把 HTML 拆成独立模板文件。

## 🟢 想做的功能

8. 门户支持搜索 / 全文检索
9. 门户支持目录级权限（哪些目录对外、哪些不对外，配置驱动而非硬编码黑名单）
10. 公开上传加可选的「上传后自动转 Markdown 预览」
11. 分享页支持更多文件类型在线预览（PDF、图片画廊、代码 diff）
12. 把这套二改整理成可向上游提 PR 的形态，或拆成独立插件

---

## 维护备忘

- 升级上游：`git fetch origin && git rebase origin/master`（在 `my-mods` 分支上 rebase）。
  冲突高发区是 `http/http.go`（路由/CSP）和 `Share.vue`。
- `master` 分支保持纯净上游，方便随时对 diff、向上游对齐。
- 历史踩坑：曾有一次构建产物坏掉（`filebrowser.broken-2026-04-27`，已清理），
  教训是构建后先本地起一遍验证再替换线上二进制。
