// Command lumen-mcp 是 LumenBrowser 的 MCP server：通过 stdio 暴露一组文件操作工具，
// 让支持 Model Context Protocol 的 AI 客户端（Claude Desktop、qwen 等）能用自然语言
// 列目录、搜索、读文件、上传、创建分享。
//
// 连接配置（server + token）默认取自 lumen CLI 的 ~/.lumen/config.json 当前 profile，
// 可用环境变量 LUMEN_SERVER / LUMEN_TOKEN 覆盖。所有请求带 X-Auth: <token>。
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	server, token, err := resolveConn()
	if err != nil {
		fmt.Fprintln(os.Stderr, "lumen-mcp 启动失败:", err)
		os.Exit(1)
	}
	c := newClient(server, token)

	s := newServer(c)
	if err := runStdio(s); err != nil {
		fmt.Fprintln(os.Stderr, "lumen-mcp 运行错误:", err)
		os.Exit(1)
	}
}

// runStdio 用 stdio transport 运行 MCP server（单独抽出便于测试/复用）。
func runStdio(s *server.MCPServer) error {
	return server.ServeStdio(s)
}

// newServer 构建注册了全部工具的 MCP server。
func newServer(c *client) *server.MCPServer {
	s := server.NewMCPServer(
		"lumen-mcp",
		"0.1.0",
		server.WithToolCapabilities(true),
	)

	s.AddTool(
		mcp.NewTool("list_files",
			mcp.WithDescription("列出 LumenBrowser 上某个目录下的文件与子目录。返回每一项的名称、是否目录、大小（字节）、修改时间与类型。"),
			mcp.WithString("path",
				mcp.Required(),
				mcp.Description("要列出的目录路径，以 / 为根，例如 \"/\" 或 \"/documents/photos\"。"),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			path, err := req.RequireString("path")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			items, err := c.list(path)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText(jsonString(items)), nil
		},
	)

	s.AddTool(
		mcp.NewTool("search_files",
			mcp.WithDescription("在 LumenBrowser 上按关键词搜索文件与目录。返回命中项的相对路径（相对搜索根目录）及其是否为目录。"),
			mcp.WithString("query",
				mcp.Required(),
				mcp.Description("搜索关键词，按文件/目录名匹配。"),
			),
			mcp.WithString("path",
				mcp.Description("搜索的起始目录，缺省为根目录 \"/\"。例如 \"/documents\" 表示只在该目录下搜索。"),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			query, err := req.RequireString("query")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			path := req.GetString("path", "/")
			if strings.TrimSpace(path) == "" {
				path = "/"
			}
			results, err := c.search(path, query)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText(jsonString(results)), nil
		},
	)

	s.AddTool(
		mcp.NewTool("read_file",
			mcp.WithDescription("读取 LumenBrowser 上某个文件的文本内容。适用于文本文件；若内容不是有效 UTF-8 文本（如二进制文件），会返回错误提示。"),
			mcp.WithString("path",
				mcp.Required(),
				mcp.Description("要读取的文件路径，例如 \"/notes/todo.txt\"。"),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			path, err := req.RequireString("path")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			data, err := c.readFile(path)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			if !utf8.Valid(data) {
				return mcp.NewToolResultError(fmt.Sprintf("%s 不是有效的 UTF-8 文本文件（可能是二进制），无法作为文本读取", path)), nil
			}
			return mcp.NewToolResultText(string(data)), nil
		},
	)

	s.AddTool(
		mcp.NewTool("upload_file",
			mcp.WithDescription("把运行 MCP server 的本机上的一个本地文件上传到 LumenBrowser 的指定远程路径（同名会覆盖）。"),
			mcp.WithString("local_path",
				mcp.Required(),
				mcp.Description("本机上要上传的文件的绝对或相对路径，例如 \"/home/user/report.pdf\"。"),
			),
			mcp.WithString("remote_path",
				mcp.Required(),
				mcp.Description("上传到 LumenBrowser 的目标文件路径（含文件名），例如 \"/uploads/report.pdf\"。"),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			local, err := req.RequireString("local_path")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			remote, err := req.RequireString("remote_path")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			n, err := c.upload(local, remote)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText(fmt.Sprintf("已上传 %s → %s（%d 字节）", local, ensureLeadingSlash(remote), n)), nil
		},
	)

	s.AddTool(
		mcp.NewTool("create_share",
			mcp.WithDescription("为 LumenBrowser 上的某个文件或目录创建一个公开分享链接，返回可直接访问的完整 URL。"),
			mcp.WithString("path",
				mcp.Required(),
				mcp.Description("要分享的文件或目录路径，例如 \"/documents/report.pdf\"。"),
			),
			mcp.WithString("type",
				mcp.Description("分享类型：\"preview\"（在线预览页，缺省）或 \"download\"（直接触发下载）。"),
				mcp.Enum("preview", "download"),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			path, err := req.RequireString("path")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			shareType := req.GetString("type", "preview")
			if shareType != "preview" && shareType != "download" {
				return mcp.NewToolResultError("type 只能是 \"preview\" 或 \"download\""), nil
			}
			link, shareURL, err := c.createShare(path, shareType)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			out := map[string]any{
				"url":  shareURL,
				"hash": link.Hash,
				"path": link.Path,
				"type": shareType,
			}
			return mcp.NewToolResultText(jsonString(out)), nil
		},
	)

	return s
}

// jsonString 把任意值序列化为带缩进的 JSON 文本，用作工具的文本返回。
func jsonString(v any) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf("结果序列化失败: %v", err)
	}
	return string(b)
}
