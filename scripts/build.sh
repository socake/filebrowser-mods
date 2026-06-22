#!/usr/bin/env bash
# LumenBrowser 完整构建脚本
# 交叉编译 lumen-mcp 各平台二进制 → http/mcpbin/（供 MCP 页直接下载，不入 git），
# 再构建前端与主程序。
set -e
cd "$(dirname "$0")/.."
ROOT="$(pwd)"
export GOTOOLCHAIN=local

echo "==> [1/3] 交叉编译 lumen-mcp 各平台二进制"
mkdir -p http/mcpbin
cd mcp
for t in windows/amd64 darwin/arm64 darwin/amd64 linux/amd64; do
  os=${t%/*}; arch=${t#*/}; ext=""; [ "$os" = windows ] && ext=.exe
  GOOS=$os GOARCH=$arch GOWORK=off go build -ldflags="-s -w" \
    -o "$ROOT/http/mcpbin/lumen-mcp-$os-$arch$ext" .
  echo "    ✓ $os/$arch"
done
cd "$ROOT"

echo "==> [2/3] 构建前端"
cd frontend
[ -d node_modules ] || pnpm install
node_modules/.bin/vite build
cd "$ROOT"

echo "==> [3/3] 构建主程序"
GOWORK=off go build -o filebrowser .

echo "完成 → ./filebrowser（已内嵌 lumen-mcp 各平台二进制，MCP 设置页可直接下载）"
