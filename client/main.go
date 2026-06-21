// Command lumen 是 LumenBrowser 的命令行客户端。
//
// 它把 LumenBrowser 服务端的 REST API 封装成终端命令，便于脚本化、CI 集成、
// 批量上传与定时备份。服务端无需任何改动——所有命令都走既有的 HTTP 接口。
package main

import (
	"fmt"
	"os"
)

const usage = `lumen — LumenBrowser CLI

用法:
  lumen <command> [arguments]

命令:
  login    登录服务器并保存连接配置 (~/.lumen/config.json)
  profile  列出/切换命名连接 profile
  ls       列出远程目录
  put      上传本地文件到远程（大文件走 tus 断点续传，带进度）
  get      下载远程文件到本地（带进度）
  rm       删除远程文件或目录
  mv       重命名/移动远程文件
  mkdir    新建远程目录
  share    为远程路径创建分享链接
  shares   列出我的分享
  search   在远程目录下搜索文件
  drop     免登录投递文件，上传即得永久分享链接

多 profile: login --profile <名> 保存多套连接，profile use <名> 切换，
其他命令可加 --profile <名> 临时指定。

用 "lumen <command> -h" 查看子命令帮助。
`

func main() {
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	}

	cmd, args := os.Args[1], os.Args[2:]

	var err error
	switch cmd {
	case "login":
		err = runLogin(args)
	case "profile":
		err = runProfile(args)
	case "ls":
		err = runLs(args)
	case "put":
		err = runPut(args)
	case "get":
		err = runGet(args)
	case "rm":
		err = runRm(args)
	case "mv":
		err = runMv(args)
	case "mkdir":
		err = runMkdir(args)
	case "share":
		err = runShare(args)
	case "shares":
		err = runShares(args)
	case "search":
		err = runSearch(args)
	case "drop":
		err = runDrop(args)
	case "-h", "--help", "help":
		fmt.Fprint(os.Stdout, usage)
		return
	default:
		fmt.Fprintf(os.Stderr, "未知命令 %q\n\n%s", cmd, usage)
		os.Exit(2)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, "lumen: 错误:", err)
		os.Exit(1)
	}
}
