// Package config 负责 grtblog CLI 的本地配置（多 profile）读写与解析。
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	// EnvConfig 覆盖配置文件路径。
	EnvConfig = "GRTBLOG_CONFIG"
	// EnvProfile 指定使用的 profile。
	EnvProfile = "GRTBLOG_PROFILE"
	// EnvServer 覆盖服务器地址。
	EnvServer = "GRTBLOG_SERVER"
	// EnvToken 覆盖管理员令牌。
	EnvToken = "GRTBLOG_TOKEN"
)

// Profile 表示一个站点的连接配置。
type Profile struct {
	Server string `yaml:"server"`
	Token  string `yaml:"token"`
}

// Config 是 config.yaml 的根结构。
type Config struct {
	Current  string              `yaml:"current,omitempty"`
	Profiles map[string]*Profile `yaml:"profiles,omitempty"`

	path string
}

// DefaultPath 返回配置文件路径：GRTBLOG_CONFIG > XDG_CONFIG_HOME/grtblog/config.yaml > ~/.config/grtblog/config.yaml。
func DefaultPath() string {
	if p := strings.TrimSpace(os.Getenv(EnvConfig)); p != "" {
		return p
	}
	if xdg := strings.TrimSpace(os.Getenv("XDG_CONFIG_HOME")); xdg != "" {
		return filepath.Join(xdg, "grtblog", "config.yaml")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "config.yaml"
	}
	return filepath.Join(home, ".config", "grtblog", "config.yaml")
}

// Load 从默认路径加载配置；文件不存在时返回空配置。
func Load() (*Config, error) {
	return LoadFrom(DefaultPath())
}

// LoadFrom 从指定路径加载配置。
func LoadFrom(path string) (*Config, error) {
	cfg := &Config{Profiles: map[string]*Profile{}, path: path}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return cfg, nil
	}
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("配置文件 %s 解析失败: %w", path, err)
	}
	if cfg.Profiles == nil {
		cfg.Profiles = map[string]*Profile{}
	}
	cfg.path = path
	return cfg, nil
}

// Path 返回配置文件实际路径。
func (c *Config) Path() string { return c.path }

// Save 写回配置文件（0600），目录不存在时创建（0700）。
func (c *Config) Save() error {
	if err := os.MkdirAll(filepath.Dir(c.path), 0o700); err != nil {
		return fmt.Errorf("创建配置目录失败: %w", err)
	}
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("配置序列化失败: %w", err)
	}
	if err := os.WriteFile(c.path, data, 0o600); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}
	return nil
}

// UpsertProfile 写入或更新一个 profile；若当前没有 current，则设为 current。
func (c *Config) UpsertProfile(name string, p Profile) {
	if c.Profiles == nil {
		c.Profiles = map[string]*Profile{}
	}
	c.Profiles[name] = &p
	if c.Current == "" {
		c.Current = name
	}
}

// RemoveProfile 删除 profile；若删除的是 current，则清空 current。
func (c *Config) RemoveProfile(name string) bool {
	if _, ok := c.Profiles[name]; !ok {
		return false
	}
	delete(c.Profiles, name)
	if c.Current == name {
		c.Current = ""
		for k := range c.Profiles {
			c.Current = k
			break
		}
	}
	return true
}

// Resolved 是最终的生效配置（flag > env > profile）。
type Resolved struct {
	Profile string
	Server  string
	Token   string
}

// Resolve 按优先级解析生效配置。
func (c *Config) Resolve(flagProfile, flagServer, flagToken string) Resolved {
	name := firstNonEmpty(flagProfile, os.Getenv(EnvProfile), c.Current, "default")
	r := Resolved{Profile: name}
	if p, ok := c.Profiles[name]; ok && p != nil {
		r.Server = p.Server
		r.Token = p.Token
	}
	if v := strings.TrimSpace(os.Getenv(EnvServer)); v != "" {
		r.Server = v
	}
	if v := strings.TrimSpace(os.Getenv(EnvToken)); v != "" {
		r.Token = v
	}
	if v := strings.TrimSpace(flagServer); v != "" {
		r.Server = v
	}
	if v := strings.TrimSpace(flagToken); v != "" {
		r.Token = v
	}
	r.Server = NormalizeServer(r.Server)
	return r
}

// NormalizeServer 规范化服务器地址：补全 scheme、去除末尾斜杠。
func NormalizeServer(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	if !strings.Contains(s, "://") {
		s = "https://" + s
	}
	return strings.TrimRight(s, "/")
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}
