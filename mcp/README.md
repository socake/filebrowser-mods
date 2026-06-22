# lumen-mcp — LumenBrowser MCP server

通过 [Model Context Protocol](https://modelcontextprotocol.io) 把 LumenBrowser 的文件操作暴露给 AI 客户端（Claude Desktop、qwen 等），让 AI 能用自然语言列目录、搜索、读文件、上传、创建分享。

传输方式：**stdio**。鉴权：复用 lumen CLI 登录后保存的 JWT，所有请求带 `X-Auth: <token>` 头。

## 工具

| 工具 | 参数 | 说明 |
| --- | --- | --- |
| `list_files` | `path` (必填) | 列目录，返回每项的 name / isDir / size / modified / type |
| `search_files` | `query` (必填), `path` (可选，缺省 `/`) | 按关键词搜索，返回命中项相对路径与是否目录 |
| `read_file` | `path` (必填) | 读取文件文本内容（非 UTF-8 文本会报错） |
| `upload_file` | `local_path` (必填), `remote_path` (必填) | 把本机本地文件上传到远程路径（同名覆盖） |
| `create_share` | `path` (必填), `type` (可选 `preview`/`download`，缺省 `preview`) | 创建公开分享，返回完整 URL `<server>/share/<hash>` |

## 准备

1. 构建二进制：

   ```bash
   export PATH=$HOME/go-sdk/go/bin:$PATH GOTOOLCHAIN=local
   cd mcp
   go mod tidy
   go build -o lumen-mcp .
   ```

   把产物放到一个固定路径，例如 `/usr/local/bin/lumen-mcp` 或 `~/bin/lumen-mcp`。

2. 登录 LumenBrowser（用 lumen CLI），token 会写到 `~/.lumen/config.json`：

   ```bash
   lumen login http://your-server:8099 -u admin
   ```

## 连接配置

server 与 token 的解析优先级：

1. 环境变量 `LUMEN_SERVER` / `LUMEN_TOKEN`（可单独覆盖其一）；
2. `~/.lumen/config.json` 的当前 profile（兼容新版多 profile 与旧版扁平格式）。

`~/.lumen/config.json`（多 profile）示例：

```json
{
  "current": "default",
  "profiles": {
    "default": {
      "server": "http://127.0.0.1:8099",
      "username": "admin",
      "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...."
    }
  }
}
```

## 在 Claude Desktop 中配置

编辑 `claude_desktop_config.json`（macOS: `~/Library/Application Support/Claude/claude_desktop_config.json`；Windows: `%APPDATA%\Claude\claude_desktop_config.json`），加入：

```json
{
  "mcpServers": {
    "lumen": {
      "command": "/usr/local/bin/lumen-mcp"
    }
  }
}
```

如需显式指定连接（不依赖 `~/.lumen/config.json`），用 `env` 覆盖：

```json
{
  "mcpServers": {
    "lumen": {
      "command": "/usr/local/bin/lumen-mcp",
      "env": {
        "LUMEN_SERVER": "http://127.0.0.1:8099",
        "LUMEN_TOKEN": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...."
      }
    }
  }
}
```

其他 MCP 客户端（如 qwen 的 MCP 接入）同理：command 指向 `lumen-mcp` 可执行文件，传输用 stdio，按需在 env 里设 `LUMEN_SERVER` / `LUMEN_TOKEN`。

配置后重启客户端，即可对 AI 说「列出根目录」「搜索包含 report 的文件」「把 /home/me/a.pdf 上传到 /uploads/a.pdf」「给 /uploads/a.pdf 创建一个下载分享链接」等。

## 注意

- token 过期后工具会返回 401 提示，重新 `lumen login` 或更新 `LUMEN_TOKEN` 即可。
- `upload_file` 的 `local_path` 是**运行 MCP server 的那台机器**上的路径。
- 本模块为独立 Go module（`github.com/socake/filebrowser-mods/mcp`），仅依赖 [`github.com/mark3labs/mcp-go`](https://github.com/mark3labs/mcp-go)，与服务端主模块解耦。
