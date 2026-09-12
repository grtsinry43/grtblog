package commands

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/grtsinry43/grtblog/cli/v2/internal/client"
	"github.com/grtsinry43/grtblog/cli/v2/internal/config"
)

func newAuthCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "认证管理（login / logout / status）",
	}

	var profileName string

	login := &cobra.Command{
		Use:   "login",
		Short: "使用管理员令牌（gt_xxx）登录",
		Example: `  grtblog auth login --server https://your.blog --token gt_xxx
  grtblog auth login   # 交互式输入`,
		RunE: func(cmd *cobra.Command, args []string) error {
			server := config.NormalizeServer(app.Resolved.Server)
			if server == "" {
				server = prompt("服务器地址 (如 https://your.blog): ", false)
				server = config.NormalizeServer(server)
			}
			if server == "" {
				return errors.New("服务器地址不能为空")
			}

			token := strings.TrimSpace(app.Resolved.Token)
			if token == "" {
				token = prompt("管理员令牌 (gt_xxx): ", true)
			}
			if token == "" {
				return errors.New("令牌不能为空")
			}

			// 验证令牌有效性
			version := app.Info.Version
			if version == "" {
				version = "dev"
			}
			probe := client.New(server, token, version)
			var user client.User
			if err := probe.Get("/auth/profile", nil, &user); err != nil {
				var apiErr *client.APIError
				if errors.As(err, &apiErr) && apiErr.IsAuth() {
					return errors.New("令牌无效或已过期，请检查后重试")
				}
				return fmt.Errorf("无法连接 %s: %w", server, err)
			}
			if !user.IsAdmin {
				return fmt.Errorf("令牌对应用户 %s 不是管理员", user.Username)
			}

			app.Cfg.UpsertProfile(profileName, config.Profile{Server: server, Token: token})
			if err := app.Cfg.Save(); err != nil {
				return err
			}
			name := user.Nickname
			if name == "" {
				name = user.Username
			}
			app.Out.Successf("已登录为 %s（管理员），站点 %s，profile %q 已保存", name, server, profileName)
			return nil
		},
	}
	login.Flags().StringVar(&profileName, "as", "default", "保存到的 profile 名称")

	logout := &cobra.Command{
		Use:   "logout",
		Short: "删除本地保存的登录凭据",
		RunE: func(cmd *cobra.Command, args []string) error {
			name := app.Resolved.Profile
			if !app.Cfg.RemoveProfile(name) {
				return fmt.Errorf("profile %q 不存在", name)
			}
			if err := app.Cfg.Save(); err != nil {
				return err
			}
			app.Out.Successf("已退出登录（profile %q 已删除）", name)
			app.Out.Warnf("注意：服务端令牌不会被吊销，可在管理后台「设置 → 令牌」中删除")
			return nil
		},
	}

	status := &cobra.Command{
		Use:   "status",
		Short: "查看当前登录状态",
		RunE: func(cmd *cobra.Command, args []string) error {
			if app.Resolved.Server == "" || app.Resolved.Token == "" {
				return errors.New("未登录，请先运行: grtblog auth login")
			}
			cli, err := app.Client()
			if err != nil {
				return err
			}
			var user client.User
			if err := cli.Get("/auth/profile", nil, &user); err != nil {
				return err
			}
			if app.Out.JSONMode {
				return app.Out.JSON(user)
			}
			name := user.Nickname
			if name == "" {
				name = user.Username
			}
			app.Out.Infof("站点:    %s", app.Resolved.Server)
			app.Out.Infof("Profile: %s", app.Resolved.Profile)
			app.Out.Infof("用户:    %s (id=%d, 管理员=%v)", name, user.ID, user.IsAdmin)
			app.Out.Infof("配置:    %s", app.Cfg.Path())
			return nil
		},
	}

	cmd.AddCommand(login, logout, status)
	return cmd
}

// prompt 从终端读取一行输入；secret 时不回显。
func prompt(label string, secret bool) string {
	fmt.Fprint(os.Stderr, label)
	if secret && term.IsTerminal(int(syscall.Stdin)) {
		b, err := term.ReadPassword(int(syscall.Stdin))
		fmt.Fprintln(os.Stderr)
		if err != nil {
			return ""
		}
		return strings.TrimSpace(string(b))
	}
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return ""
	}
	return strings.TrimSpace(scanner.Text())
}
