# 技术实现笔记

记录二改的实现思路、关键代码位置与设计取舍，方便日后自己接手。

---

## 路由注册（http/http.go）

所有自定义路由都挂在**根 router `r`** 上，而不是 `/api` 子路由：

```go
r.Handle("/docs", monkey(docsPageHandler, "")).Methods("GET")
r.Handle("/docs/list", monkey(docsListHandler, "")).Methods("GET")
r.Handle("/docs/content", monkey(docsContentHandler, "")).Methods("GET")
r.Handle("/docs/download", monkey(docsDownloadHandler, "")).Methods("GET")
...
public.Handle("/upload", monkey(publicUploadHandler, "")).Methods("POST")
```

`monkey(fn, prefix)` 只是 `handle(fn, prefix, store, server)` 的包装，**不做鉴权**——
鉴权在 filebrowser 里是通过 `/api` 子路由上的 auth 中间件实现的。因为这些路由不在 auth 链上，
所以它们是**匿名可访问**的。这是 `/docs` 门户和公开上传的设计前提，但也是最大的安全面（见 ROADMAP）。

`/docs` 路由必须注册在 `r.NotFoundHandler = index` **之前**，否则会被前端 SPA 的 fallback 吞掉。

> 注意：因为没有登录态，处理函数里都用 `d.store.Users.Get(d.server.Root, uint(1))` **硬取 ID=1 的管理员**
> 来拿文件系统句柄。隐含假设：admin 永远是 user 1。

## CSP 放开（http/http.go）

原版 CSP 是 `default-src 'self'; style-src 'unsafe-inline';`，门户要用到外链的
highlight.js 和 diagrams.net 嵌入式 viewer，所以放开为：

```
default-src 'self';
style-src 'unsafe-inline' https://cdnjs.cloudflare.com;
script-src 'self' 'unsafe-inline' 'unsafe-eval' https://cdnjs.cloudflare.com https://viewer.diagrams.net;
connect-src 'self'; img-src 'self' data:;
```

`unsafe-eval` 是 diagrams.net viewer 需要的，引入了一定 XSS 风险，属于取舍。

---

## 文档门户（http/docs.go + docs_page.go）

- **docs_page.go**：一个 513 行的 Go 常量字符串 `docsHTML`，内嵌完整单页应用（HTML+CSS+JS，zh-CN）。
  好处是零构建、随二进制走；坏处是不好维护（见 ROADMAP，可考虑改 embed 模板文件）。
- **docs.go**：四个 handler + 安全工具函数。
  - `hiddenOrSensitive(name)`：点开头文件 + 敏感名子串黑名单，**list / content 都调用了**。
  - 路径越界防护：`filepath.Join(root, reqPath)` 后 `filepath.Abs`，再
    `strings.HasPrefix(realPath, realRoot)`。
    > 这里有个经典 sibling-prefix bug：`/home/ubuntu` 会匹配到 `/home/ubuntu-secret`。
    > 应改成 `realRoot + string(os.PathSeparator)` 前缀比较。已记进 ROADMAP。
  - content 预览限制 10MB，超过返回 413。
  - **download handler 没调 `hiddenOrSensitive`** —— 已知漏洞，见 ROADMAP。

## 公开上传（http/public_upload.go）

流程：取 admin FS → `MkdirAll(/uploads)` → `ParseMultipartForm(1GB)` → 取 `file` 字段 →
`filepath.Base` 清洗文件名 → 时间戳前缀（碰撞再加 `_1` 后缀）→ 写盘 →
`crypto/rand` 生成 hash → `share.Save` 存一条 `Expire:0`（永久）分享 → 返回 JSON。

设计取舍：
- 用 admin 的 `user.Fs` 而非裸 `os`，复用 filebrowser 的虚拟文件系统/作用域。
- 分享永久不过期，方便但会无限堆积（见 ROADMAP）。

---

## 分享页与预览（前端）

- **public.go** 加 `Content: true`：让分享接口在返回文件元信息的同时带上内容，前端才能内联预览，
  不必再发一次下载请求。
- **share.go** 把分享 hash 从 `make([]byte, 6)` 改成 `make([]byte, 24)`：
  6 字节 base64 约 8 字符、熵偏低，对永久分享不安全；24 字节足够抗遍历。
- **Share.vue / Editor.vue / mdPreview.css**：前端渲染逻辑与样式，预览走 highlight.js + markdown 渲染。

---

## 关键文件速查

| 关注点 | 文件:符号 |
|--------|----------|
| 路由注册 / CSP | `http/http.go` `NewHandler` |
| 门户后端 | `http/docs.go` `docsListHandler` / `docsContentHandler` / `docsDownloadHandler` |
| 门户前端 | `http/docs_page.go` `docsHTML` |
| 敏感文件过滤 | `http/docs.go` `hiddenOrSensitive` |
| 公开上传 | `http/public_upload.go` `publicUploadHandler` |
| 分享内容回传 | `http/public.go` `withHashFile` |
| 分享 hash 熵 | `http/share.go` `sharePostHandler` |
| 分享页 / 预览 | `frontend/src/views/Share.vue` `frontend/src/views/files/Editor.vue` |
