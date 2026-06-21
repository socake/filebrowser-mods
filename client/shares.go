package main

import (
	"flag"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/socake/filebrowser-mods/client/internal/api"
	"github.com/socake/filebrowser-mods/client/internal/config"
)

// runShares 列出我的分享。背后接口：GET /api/shares。
func runShares(args []string) error {
	fs := flag.NewFlagSet("shares", flag.ExitOnError)
	fs.Parse(args)

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	links, err := api.New(cfg.Server, cfg.Token).ListShares()
	if err != nil {
		return err
	}
	if len(links) == 0 {
		fmt.Println("（暂无分享）")
		return nil
	}

	tw := tabwriter.NewWriter(os.Stdout, 0, 2, 2, ' ', 0)
	fmt.Fprintln(tw, "HASH\t到期\t路径")
	for _, l := range links {
		exp := "永久"
		if l.Expire != 0 {
			exp = time.Unix(l.Expire, 0).Format("2006-01-02 15:04")
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\n", l.Hash, exp, l.Path)
	}
	return tw.Flush()
}
