package main

import (
	"flag"
	"fmt"

	"github.com/socake/filebrowser-mods/client/internal/api"
	"github.com/socake/filebrowser-mods/client/internal/config"
)

// runSearch 在远程目录下搜索文件。背后接口：GET /api/search/<path>?query=（流式 NDJSON）。
//
// 用法: lumen search [-in 目录] <关键词>，-in 缺省为根目录 /。
// 命中的 path 相对搜索根目录。
func runSearch(args []string) error {
	fs := flag.NewFlagSet("search", flag.ExitOnError)
	root := fs.String("in", "/", "搜索根目录")
	fs.Usage = func() {
		fmt.Println("用法: lumen search [-in 目录] <关键词>")
		fs.PrintDefaults()
	}
	fs.Parse(args)

	query := fs.Arg(0)
	if query == "" {
		fs.Usage()
		return fmt.Errorf("缺少关键词")
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	results, err := api.New(cfg.Server, cfg.Token).Search(*root, query)
	if err != nil {
		return err
	}
	if len(results) == 0 {
		fmt.Println("（无匹配）")
		return nil
	}
	for _, r := range results {
		kind := "-"
		if r.Dir {
			kind = "d"
		}
		fmt.Printf("%s  %s\n", kind, r.Path)
	}
	return nil
}
