package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// profile 是一份连接配置（与 lumen CLI 的 ~/.lumen/config.json 中每个 profile 同构）。
type profile struct {
	Server   string `json:"server"`
	Username string `json:"username"`
	Token    string `json:"token"`
}

// store 是 ~/.lumen/config.json 的新版（多 profile）结构。
type store struct {
	Current  string              `json:"current"`
	Profiles map[string]*profile `json:"profiles"`
}

// resolveConn 解析出要连接的 server 与 token。
//
// 优先级：环境变量 LUMEN_SERVER / LUMEN_TOKEN > ~/.lumen/config.json 的当前 profile。
// 两个环境变量可单独覆盖（例如只设 LUMEN_TOKEN，server 仍取自配置文件）。
func resolveConn() (server, token string, err error) {
	server = os.Getenv("LUMEN_SERVER")
	token = os.Getenv("LUMEN_TOKEN")

	if server == "" || token == "" {
		p, ferr := loadCurrentProfile()
		if ferr != nil {
			// 配置缺失时，只有当环境变量也没补齐才算错误。
			if server == "" || token == "" {
				return "", "", ferr
			}
		} else {
			if server == "" {
				server = p.Server
			}
			if token == "" {
				token = p.Token
			}
		}
	}

	if server == "" {
		return "", "", fmt.Errorf("未配置服务器地址：请先运行 lumen login，或设置环境变量 LUMEN_SERVER")
	}
	if token == "" {
		return "", "", fmt.Errorf("未配置鉴权 token：请先运行 lumen login，或设置环境变量 LUMEN_TOKEN")
	}
	return strings.TrimRight(server, "/"), token, nil
}

// loadCurrentProfile 读取 ~/.lumen/config.json 的当前 profile，
// 兼容新版（{current,profiles}）与旧版扁平（{server,token,...}）两种格式。
func loadCurrentProfile() (*profile, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	path := filepath.Join(home, ".lumen", "config.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("未找到 %s：请先运行 lumen login", path)
		}
		return nil, err
	}

	// 先按新版多 profile 解析。
	var s store
	if err := json.Unmarshal(data, &s); err == nil && len(s.Profiles) > 0 {
		if p, ok := s.Profiles[s.Current]; ok && p != nil {
			return p, nil
		}
		// current 失效：退回任意一个可用 profile。
		for _, p := range s.Profiles {
			if p != nil {
				return p, nil
			}
		}
	}

	// 回退：旧版扁平单 profile。
	var legacy profile
	if err := json.Unmarshal(data, &legacy); err == nil && legacy.Server != "" {
		return &legacy, nil
	}

	return nil, fmt.Errorf("%s 内容无法识别：请重新运行 lumen login", path)
}
