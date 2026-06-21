// Package config 负责 lumen CLI 的本地连接配置（~/.lumen/config.json）。
package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"
)

// Config 是一份连接 profile：登录后写入，后续命令读取。
type Config struct {
	Server   string    `json:"server"`
	Username string    `json:"username"`
	Token    string    `json:"token"`
	Expires  time.Time `json:"expires,omitempty"`
}

// ErrNotConfigured 表示尚未登录（配置文件不存在）。
var ErrNotConfigured = errors.New("尚未登录：请先运行 lumen login <服务器地址>")

func dir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".lumen"), nil
}

func file() (string, error) {
	d, err := dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, "config.json"), nil
}

// Load 读取配置；文件不存在时返回 ErrNotConfigured。
func Load() (*Config, error) {
	p, err := file()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotConfigured
		}
		return nil, err
	}
	var c Config
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

// Save 把配置写到 ~/.lumen/config.json（目录 0700、文件 0600，因含 token）。
func (c *Config) Save() error {
	d, err := dir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(d, 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(d, "config.json"), data, 0o600)
}
