// Package commands 包含 grtblog CLI 的全部子命令。
package commands

import (
	"errors"

	"github.com/grtsinry43/grtblog/cli/v2/internal/client"
	"github.com/grtsinry43/grtblog/cli/v2/internal/config"
	"github.com/grtsinry43/grtblog/cli/v2/internal/output"
)

// BuildInfo 构建信息，通过 -ldflags 注入。
type BuildInfo struct {
	Version string
	Commit  string
	Date    string
}

// App 是所有命令共享的上下文。
type App struct {
	Info     BuildInfo
	Out      *output.Printer
	Cfg      *config.Config
	Resolved config.Resolved

	flagServer  string
	flagToken   string
	flagProfile string
	flagJSON    bool
	flagNoColor bool

	client *client.Client
}

// Init 在 PersistentPreRun 阶段加载配置并构造输出器。
func (a *App) Init() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	a.Cfg = cfg
	a.Resolved = cfg.Resolve(a.flagProfile, a.flagServer, a.flagToken)
	a.Out = output.New(a.flagJSON, a.flagNoColor)
	return nil
}

// Client 惰性构造 API 客户端；缺少 server/token 时给出引导性错误。
func (a *App) Client() (*client.Client, error) {
	if a.client != nil {
		return a.client, nil
	}
	if a.Resolved.Server == "" {
		return nil, errors.New("未配置服务器地址，请先运行: grtblog auth login")
	}
	if a.Resolved.Token == "" {
		return nil, errors.New("未登录，请先运行: grtblog auth login")
	}
	version := a.Info.Version
	if version == "" {
		version = "dev"
	}
	a.client = client.New(a.Resolved.Server, a.Resolved.Token, version)
	return a.client, nil
}
