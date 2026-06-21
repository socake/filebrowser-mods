package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/socake/filebrowser-mods/client/internal/api"
	"github.com/socake/filebrowser-mods/client/internal/config"
)

// runLogin 登录服务器，拿到 JWT 并把连接配置写到 ~/.lumen/config.json。
// 背后接口：POST /api/login。
func runLogin(args []string) error {
	fs := flag.NewFlagSet("login", flag.ExitOnError)
	var (
		username = fs.String("u", "", "用户名")
		password = fs.String("p", "", "密码（留空则提示输入）")
	)
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "用法: lumen login <服务器地址> [-u 用户名] [-p 密码]")
		fs.PrintDefaults()
	}
	fs.Parse(args)

	server := fs.Arg(0)
	if server == "" {
		fs.Usage()
		return fmt.Errorf("缺少服务器地址")
	}

	user := *username
	if user == "" {
		fmt.Print("用户名: ")
		user = readLine()
	}
	pass := *password
	if pass == "" {
		fmt.Print("密码: ")
		pass = readLine()
	}

	token, err := api.Login(server, user, pass)
	if err != nil {
		return err
	}

	cfg := &config.Config{
		Server:   strings.TrimRight(server, "/"),
		Username: user,
		Token:    token,
	}
	if err := cfg.Save(); err != nil {
		return err
	}
	fmt.Printf("已登录 %s，配置已保存到 ~/.lumen/config.json\n", cfg.Server)
	return nil
}

func readLine() string {
	line, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	return strings.TrimSpace(line)
}
