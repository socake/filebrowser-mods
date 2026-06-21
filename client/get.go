package main

import (
	"flag"
	"fmt"
	"path"
	"path/filepath"

	"github.com/socake/filebrowser-mods/client/internal/api"
	"github.com/socake/filebrowser-mods/client/internal/config"
)

// runGet 下载远程文件到本地。背后接口：GET /api/raw/<path>。
//
// 本地参数省略或以 / 结尾时，用远程文件名落地到该目录。
func runGet(args []string) error {
	fs := flag.NewFlagSet("get", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Println("用法: lumen get <远程路径> [本地路径]")
		fs.PrintDefaults()
	}
	fs.Parse(args)

	remote, local := fs.Arg(0), fs.Arg(1)
	if remote == "" {
		fs.Usage()
		return fmt.Errorf("缺少远程路径")
	}
	if local == "" {
		local = path.Base(remote)
	} else if local[len(local)-1] == '/' {
		local = filepath.Join(local, path.Base(remote))
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	n, err := api.New(cfg.Server, cfg.Token).Get(remote, local)
	if err != nil {
		return err
	}
	fmt.Printf("已下载: %s -> %s (%d 字节)\n", remote, local, n)
	return nil
}
