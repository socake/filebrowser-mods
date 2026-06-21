package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/socake/filebrowser-mods/client/internal/api"
	"github.com/socake/filebrowser-mods/client/internal/config"
)

// runShare 为远程路径创建分享链接。背后接口：POST /api/share/<path>。
//
// 分享前端地址为 <server>/share/<hash>。可选 -password 加密码，
// -expires 配合 -unit 设置有效期（unit ∈ seconds/minutes/hours/days，缺省 hours）。
func runShare(args []string) error {
	fs := flag.NewFlagSet("share", flag.ExitOnError)
	var (
		password = fs.String("password", "", "访问密码（留空则无密码）")
		expires  = fs.String("expires", "", "有效期数值（留空表示永久）")
		unit     = fs.String("unit", "hours", "有效期单位: seconds|minutes|hours|days")
	)
	fs.Usage = func() {
		fmt.Println("用法: lumen share [-password ..] [-expires N -unit hours] <远程路径>")
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
	link, err := api.New(cfg.Server, cfg.Token).CreateShare(remote, *password, *expires, *unit)
	if err != nil {
		return err
	}

	fmt.Printf("分享链接: %s/share/%s\n", cfg.Server, link.Hash)
	if *password != "" {
		fmt.Printf("访问密码: %s\n", *password)
	}
	if link.Expire == 0 {
		fmt.Println("有效期: 永久")
	} else {
		fmt.Printf("到期时间: %s\n", time.Unix(link.Expire, 0).Format("2006-01-02 15:04:05"))
	}
	return nil
}
