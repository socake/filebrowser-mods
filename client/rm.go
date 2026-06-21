package main

import (
	"flag"
	"fmt"

	"github.com/socake/filebrowser-mods/client/internal/api"
	"github.com/socake/filebrowser-mods/client/internal/config"
)

// runRm 删除远程文件或目录。背后接口：DELETE /api/resources/<path>。
func runRm(args []string) error {
	fs := flag.NewFlagSet("rm", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Println("用法: lumen rm <远程路径>")
		fs.PrintDefaults()
	}
	fs.Parse(args)

	remote := fs.Arg(0)
	if remote == "" {
		fs.Usage()
		return fmt.Errorf("缺少远程路径")
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if err := api.New(cfg.Server, cfg.Token).Remove(remote); err != nil {
		return err
	}
	fmt.Printf("已删除: %s\n", remote)
	return nil
}
