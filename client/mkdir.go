package main

import (
	"flag"
	"fmt"

	"github.com/socake/filebrowser-mods/client/internal/api"
	"github.com/socake/filebrowser-mods/client/internal/config"
)

// runMkdir 新建远程目录。背后接口：POST /api/resources/<path>/（结尾斜杠表示目录）。
func runMkdir(args []string) error {
	fs := flag.NewFlagSet("mkdir", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Println("用法: lumen mkdir <远程目录>")
		fs.PrintDefaults()
	}
	fs.Parse(args)

	remote := fs.Arg(0)
	if remote == "" {
		fs.Usage()
		return fmt.Errorf("缺少远程目录")
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if err := api.New(cfg.Server, cfg.Token).Mkdir(remote); err != nil {
		return err
	}
	fmt.Printf("已创建目录: %s\n", remote)
	return nil
}
