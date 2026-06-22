# mcpbin — lumen-mcp 预编译二进制

本目录存放交叉编译好的 `lumen-mcp` 各平台二进制，由主程序 `go:embed`
内嵌，供「设置 → MCP 接入」页面直接下载，用户无需自行构建。

二进制（`lumen-mcp-<os>-<arch>`）**不入 git**（见 .gitignore），
由构建脚本生成：

```bash
./scripts/build.sh
```

该脚本会交叉编译 windows/amd64、darwin/arm64、darwin/amd64、linux/amd64
四个平台到此目录，再构建前端与主程序。

clone 后若未运行构建脚本，此目录仅有本说明文件，下载接口返回 404、
MCP 页平台列表为空——运行 `scripts/build.sh` 即可。
