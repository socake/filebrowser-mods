# 后期开发计划 & 暂未完善的点

诚实记录当前已知的问题与未来想做的事。**当前状态：能用，但 `/docs` 与公开上传都是匿名开放，
只适合可信内网或临时场景，不要直接裸暴露公网。**

---

## 🔴 必修（安全，上公网前务必处理）

1. **`/docs/download` 绕过敏感文件过滤**
   `docsDownloadHandler`（`http/docs.go`）没有调用 `hiddenOrSensitive`，
   导致虽然 list/preview 看不到 `.env`、`id_rsa`、`credentials` 等文件，
   但只要知道路径就能直接 `GET /docs/download?path=/.env` 下载。
   → 修法：download 前加同样的 `hiddenOrSensitive(filepath.Base(reqPath))` 检查。

2. **路径越界判断的 sibling-prefix bug**
   `strings.HasPrefix(realPath, realRoot)` 会把 `/home/ubuntu-secret` 误判为在 `/home/ubuntu` 内。
   → 修法：比较 `realPath == realRoot || strings.HasPrefix(realPath, realRoot + string(os.PathSeparator))`。
   三个 handler（list/content/download）都要改。

3. **公开上传是无鉴权开放投递箱**
   `POST /api/public/upload` 无 auth、无限速、无类型校验，单文件 1GB，且生成**永久**分享。
   任何人可填满磁盘、上传恶意文件、刷爆分享表。
   → 选项：加上传 token / 一次性凭证、限速、单 IP 配额、文件类型白名单、分享默认带过期。

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
