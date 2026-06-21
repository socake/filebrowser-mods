package main

import (
	"flag"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/socake/filebrowser-mods/client/internal/api"
	"github.com/socake/filebrowser-mods/client/internal/config"
)

// runLs 列出远程目录。背后接口：GET /api/resources/<path>。
func runLs(args []string) error {
	fs := flag.NewFlagSet("ls", flag.ExitOnError)
	profile := fs.String("profile", "", "指定使用的 profile（缺省用当前）")
	fs.Parse(args)

	path := fs.Arg(0)
	if path == "" {
		path = "/"
	}

	cfg, err := config.Load(*profile)
	if err != nil {
		return err
	}

	items, err := api.New(cfg.Server, cfg.Token).List(path)
	if err != nil {
		return err
	}

	tw := tabwriter.NewWriter(os.Stdout, 0, 2, 2, ' ', 0)
	for _, it := range items {
		kind, name := "-", it.Name
		if it.IsDir {
			kind, name = "d", it.Name+"/"
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\n", kind, humanSize(it.Size), name)
	}
	return tw.Flush()
}

func humanSize(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%dB", n)
	}
	div, exp := int64(unit), 0
	for m := n / unit; m >= unit; m /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%cB", float64(n)/float64(div), "KMGTPE"[exp])
}
