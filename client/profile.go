package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/socake/filebrowser-mods/client/internal/config"
)

// runProfile 管理多个命名连接 profile。
//
// 用法:
//
//	lumen profile            列出所有 profile（标注当前）
//	lumen profile use <名>   切换当前 profile
func runProfile(args []string) error {
	fs := flag.NewFlagSet("profile", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Println("用法:")
		fmt.Println("  lumen profile            列出所有 profile")
		fmt.Println("  lumen profile use <名>   切换当前 profile")
	}
	fs.Parse(args)

	// 子命令 use：切换当前 profile。
	if fs.Arg(0) == "use" {
		name := fs.Arg(1)
		if name == "" {
			return fmt.Errorf("用法: lumen profile use <名>")
		}
		if err := config.SetCurrent(name); err != nil {
			return err
		}
		fmt.Printf("已切换当前 profile 为 %q\n", name)
		return nil
	}

	// 默认：列出所有 profile。
	store, err := config.LoadStore()
	if err != nil {
		return err
	}
	names := make([]string, 0, len(store.Profiles))
	for n := range store.Profiles {
		names = append(names, n)
	}
	sort.Strings(names)

	tw := tabwriter.NewWriter(os.Stdout, 0, 2, 2, ' ', 0)
	fmt.Fprintln(tw, "当前\tPROFILE\t服务器\t用户")
	for _, n := range names {
		mark := ""
		if n == store.Current {
			mark = "*"
		}
		c := store.Profiles[n]
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", mark, n, c.Server, c.Username)
	}
	return tw.Flush()
}
