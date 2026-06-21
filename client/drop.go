package main

import (
	"flag"
	"fmt"

	"github.com/socake/filebrowser-mods/client/internal/api"
	"github.com/socake/filebrowser-mods/client/internal/config"
)

// runDrop 免登录投递一个文件，上传成功后返回永久分享链接。
// 背后接口：POST /api/public/upload（无需鉴权，本分支二改新增）。
//
// 用法约定：可选 flag 需写在文件名之前，如 lumen drop -server https://x file.pdf。
func runDrop(args []string) error {
	fs := flag.NewFlagSet("drop", flag.ExitOnError)
	server := fs.String("server", "", "服务器地址（未登录时必填）")
	profile := fs.String("profile", "", "从指定 profile 取服务器地址（缺省用当前）")
	fs.Parse(args)

	file := fs.Arg(0)
	if file == "" {
		return fmt.Errorf("用法: lumen drop [-server 地址] [--profile 名] <文件路径>")
	}

	target := *server
	if target == "" {
		if cfg, err := config.Load(*profile); err == nil {
			target = cfg.Server
		}
	}
	if target == "" {
		return fmt.Errorf("未指定服务器：用 -server 指定，或先 lumen login")
	}

	res, err := api.PublicUpload(target, file)
	if err != nil {
		return err
	}
	fmt.Printf("已投递: %s (%d 字节)\n", res.Filename, res.Size)
	fmt.Printf("下载链接: %s%s\n", target, res.DownloadURL)
	fmt.Printf("浏览链接: %s%s\n", target, res.BrowseURL)
	return nil
}
