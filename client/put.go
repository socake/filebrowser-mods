package main

import (
	"flag"
	"fmt"
	"path"
	"path/filepath"

	"github.com/socake/filebrowser-mods/client/internal/api"
	"github.com/socake/filebrowser-mods/client/internal/config"
	"github.com/socake/filebrowser-mods/client/internal/progress"
)

// runPut 上传本地文件到远程路径。
//
// 大文件走 tus 断点续传（中断重跑自动续传），小文件整文件直传；传输时显示进度。
// 若远程路径以 / 结尾，视为目标目录，自动拼上本地文件名。
func runPut(args []string) error {
	fs := flag.NewFlagSet("put", flag.ExitOnError)
	profile := fs.String("profile", "", "指定使用的 profile（缺省用当前）")
	fs.Usage = func() {
		fmt.Println("用法: lumen put [--profile 名] <本地文件> <远程路径>")
		fs.PrintDefaults()
	}
	fs.Parse(args)

	local, remote := fs.Arg(0), fs.Arg(1)
	if local == "" || remote == "" {
		fs.Usage()
		return fmt.Errorf("缺少参数")
	}
	// 远程以 / 结尾：拼接本地文件名作为目标文件。
	if remote[len(remote)-1] == '/' {
		remote = path.Join(remote, filepath.Base(local))
	}

	cfg, err := config.Load(*profile)
	if err != nil {
		return err
	}
	bar := progress.New("上传 " + filepath.Base(local))
	if err := api.New(cfg.Server, cfg.Token).Put(remote, local, bar); err != nil {
		return err
	}
	fmt.Printf("已上传: %s -> %s\n", local, remote)
	return nil
}
