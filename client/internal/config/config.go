// Package config 负责 lumen CLI 的本地连接配置（~/.lumen/config.json）。
//
// 配置支持多个命名 profile（如 prod / dev），结构形如：
//
//	{
//	  "current": "prod",
//	  "profiles": {
//	    "prod": {"server": "...", "username": "...", "token": "...", "expires": "..."},
//	    "dev":  {"server": "...", "username": "...", "token": "..."}
//	  }
//	}
//
// 向后兼容：旧版单 profile 配置是一个扁平对象
// {"server":..,"username":..,"token":..,"expires":..}，Load 时会把它当作名为
// "default" 的唯一 profile 读入。
package config

import (
	"encoding/json"
	"errors"
	"fmt"
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

// Store 是磁盘上的完整配置：一组命名 profile + 当前选中的 profile 名。
type Store struct {
	Current  string             `json:"current"`
	Profiles map[string]*Config `json:"profiles"`
}

// DefaultProfile 是未显式命名时使用的 profile 名（也是旧配置迁移后的名字）。
const DefaultProfile = "default"

// ErrNotConfigured 表示尚未登录（配置文件不存在或为空）。
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

// LoadStore 读取整份配置，自动识别新（多 profile）/旧（单 profile）两种格式。
// 文件不存在时返回 ErrNotConfigured。
func LoadStore() (*Store, error) {
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

	// 优先按新格式解析。
	var s Store
	if err := json.Unmarshal(data, &s); err == nil && len(s.Profiles) > 0 {
		if _, ok := s.Profiles[s.Current]; !ok {
			// current 缺失或失效：退回任意一个 profile，保证可用。
			for name := range s.Profiles {
				s.Current = name
				break
			}
		}
		return &s, nil
	}

	// 回退：旧版扁平单 profile 配置。
	var legacy Config
	if err := json.Unmarshal(data, &legacy); err == nil && legacy.Server != "" {
		return &Store{
			Current:  DefaultProfile,
			Profiles: map[string]*Config{DefaultProfile: &legacy},
		}, nil
	}

	return nil, ErrNotConfigured
}

// Load 取出指定 profile 的配置；profile 为空时用当前 profile。
func Load(profile string) (*Config, error) {
	s, err := LoadStore()
	if err != nil {
		return nil, err
	}
	name := profile
	if name == "" {
		name = s.Current
	}
	c, ok := s.Profiles[name]
	if !ok {
		return nil, fmt.Errorf("profile %q 不存在（用 lumen profile 查看可用 profile）", name)
	}
	return c, nil
}

// SaveProfile 写入/更新一个命名 profile，并把它设为当前 profile。
// 文件不存在时会新建。其余 profile 保持不变。
func SaveProfile(name string, c *Config) error {
	s, err := LoadStore()
	if err != nil {
		if !errors.Is(err, ErrNotConfigured) {
			return err
		}
		s = &Store{Profiles: map[string]*Config{}}
	}
	if s.Profiles == nil {
		s.Profiles = map[string]*Config{}
	}
	s.Profiles[name] = c
	s.Current = name
	return s.save()
}

// SetCurrent 切换当前 profile（必须已存在）。
func SetCurrent(name string) error {
	s, err := LoadStore()
	if err != nil {
		return err
	}
	if _, ok := s.Profiles[name]; !ok {
		return fmt.Errorf("profile %q 不存在", name)
	}
	s.Current = name
	return s.save()
}

// save 把整份配置写到 ~/.lumen/config.json（目录 0700、文件 0600，因含 token）。
func (s *Store) save() error {
	d, err := dir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(d, 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(d, "config.json"), data, 0o600)
}
