package main

import (
	"flag"
	"fmt"

	"github.com/socake/filebrowser-mods/client/internal/api"
	"github.com/socake/filebrowser-mods/client/internal/config"
)

// runMv 重命名/移动远程文件。背后接口：
// PATCH /api/resources/<src>?action=rename&destination=<dst>。
func runMv(args []string) error {
	fs := flag.NewFlagSet("mv", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Println("用法: lumen mv <源路径> <目标路径>")
		fs.PrintDefaults()
	}
	fs.Parse(args)

	src, dst := fs.Arg(0), fs.Arg(1)
	if src == "" || dst == "" {
		fs.Usage()
		return fmt.Errorf("缺少参数")
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if err := api.New(cfg.Server, cfg.Token).Move(src, dst); err != nil {
		return err
	}
	fmt.Printf("已移动: %s -> %s\n", src, dst)
	return nil
}
