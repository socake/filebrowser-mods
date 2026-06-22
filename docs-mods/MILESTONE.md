# 产品化里程碑总结（2026-06-22）

从「filebrowser 个人二改」到「**LumenBrowser** 产品」的一次完整产品化收尾。
技术二改细节见 [FEATURES](./FEATURES.md) / [IMPLEMENTATION](./IMPLEMENTATION.md) / [PRODUCT-DESIGN](./PRODUCT-DESIGN.md)，
技术与安全待办见 [ROADMAP](./ROADMAP.md)。本文件记录**产品化与运营层面**的成果。

---

## 一、品牌与开源

- 仓库改名 `filebrowser-mods` → **`socake/LumenBrowser`**（GitHub public，旧名自动重定向）
- Slogan：**把你的文件带到光下 · Bring files to light**；主色 `#141414`
- 纯白 To C 视觉重做：登录页、公开宣传页 `/about`、设置页 To C 重构、侧栏 Powered by
- `master` = 纯净上游，`my-mods` = 二改主分支

## 二、功能产品化（本轮已落地）

- 多类型在线预览（图片/视频/Office/PDF/Markdown，16+ 类型图标）
- AI 助手接入（MCP server，设置页一键生成令牌 + 各平台 `lumen-mcp` 二进制下载，免构建）
- RBAC 角色权限（4 预设角色 + 新用户默认角色 + 个人覆盖，权限直接作用到界面）
- 分享管理 / 角色管理 / 用户管理提到一级导航；分享管理多选 + 筛选
- 新建合并入口、登录页与宣传页打磨

## 三、演示环境（已上线）

- 在线演示：**https://file.vishine.top**（注册已关闭，演示数据内置镜像）
- 部署链路：
  1. 本地 `scripts/build.sh`（⚠️ **`CGO_ENABLED=0` 静态编译**，否则 alpine/musl 运行报 `exec ... no such file`）
  2. `docker build -f Dockerfile.demo` → 推**阿里云 ACR**
  3. 轻量 ECS `docker pull && docker run`（容器 `127.0.0.1:8091`，db 卷，`--restart unless-stopped`）
  4. 宿主 Nginx 反代 → certbot 签 HTTPS
- 弱配置服务器策略：**本地构建、远程只拉取**，不在服务器上 build

## 四、文档体系

| 文档 | 用途 |
| --- | --- |
| [README.md](../README.md) | 产品介绍 + 特性 + 截图 + 快速开始 + 赞助 + Star History |
| [docs/DEPLOY.md](../docs/DEPLOY.md) | 私有化部署详解（构建 / Docker / ACR / Nginx / certbot / 备份 / FAQ）|
| [docs-mods/](./README.md) | 二改技术文档（功能 / 实现 / 接口 / 设计 / roadmap / 本里程碑）|
| `docs/img/` | 功能截图 + 微信&支付宝赞助二维码 |

## 五、官网展示

- 已加入个人官网 [www.vishine.top](https://www.vishine.top)（3w.vshy.top）的产品板块
- 位置：生物命题宝之后、LoopyPet 之前；项目图标 = 黑框 + 黑白同心圆 glyph
- 跳转：打开 → `file.vishine.top`，了解 → GitHub（与其他项目一致）

## 六、开源与隐私

- 仓库（含 git 历史）**无任何真实密码 / 服务器 IP / ACR 凭据 / token**
- 演示站去掉 `admin/admin` 账号提示（防他人登录滥用服务器），保留自部署默认密码的改密提醒
- 赞助区：微信 + 支付宝收款二维码 + GitHub

---

## 当前状态

✅ **产品化规划完成**：开源发布 + 演示站上线 + 文档齐全 + 官网展示。

后续技术演进与安全加固详见 [ROADMAP.md](./ROADMAP.md)（如公开上传鉴权、永久分享清理、门户 UI embed 化等）。
