package main

import (
	"errors"
	"fmt"
	"os"
	"runtime/debug"

	"github.com/grtsinry43/grtblog/cli/v2/internal/client"
	"github.com/grtsinry43/grtblog/cli/v2/internal/commands"
)

var (
	version = ""
	commit  = ""
	date    = ""
)

func main() {
	root := commands.NewRootCmd(resolveBuildInfo())
	if err := root.Execute(); err != nil {
		var apiErr *client.APIError
		switch {
		case errors.As(err, &apiErr) && apiErr.IsAuth():
			fmt.Fprintf(os.Stderr, "错误: %s（登录已过期或令牌无效，请重新运行 grtblog auth login）\n", apiErr.Msg)
		default:
			fmt.Fprintf(os.Stderr, "错误: %s\n", err)
		}
		os.Exit(1)
	}
}

// resolveBuildInfo 优先使用 ldflags 注入的版本；否则回退到 Go 构建信息
// （go install 安装时，这里会给出模块版本，即项目 tag 或伪版本）。
func resolveBuildInfo() commands.BuildInfo {
	info := commands.BuildInfo{Version: version, Commit: commit, Date: date}

	bi, ok := debug.ReadBuildInfo()
	if !ok {
		if info.Version == "" {
			info.Version = "dev"
		}
		return info
	}

	if info.Version == "" || info.Version == "dev" {
		if v := bi.Main.Version; v != "" && v != "(devel)" {
			info.Version = v
		}
	}
	for _, s := range bi.Settings {
		switch s.Key {
		case "vcs.revision":
			if info.Commit == "" {
				info.Commit = s.Value
				if len(info.Commit) > 12 {
					info.Commit = info.Commit[:12]
				}
			}
		case "vcs.time":
			if info.Date == "" {
				info.Date = s.Value
			}
		}
	}
	if info.Version == "" {
		info.Version = "dev"
	}
	return info
}
