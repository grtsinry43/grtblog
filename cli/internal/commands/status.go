package commands

import (
	"github.com/spf13/cobra"

	"github.com/grtsinry43/grtblog/cli/v2/internal/client"
	"github.com/grtsinry43/grtblog/cli/v2/internal/output"
)

func newStatusCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "查看服务器运行状态",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli, err := app.Client()
			if err != nil {
				return err
			}
			var s client.SystemStatus
			if err := cli.Get("/admin/system/status", nil, &s); err != nil {
				return err
			}
			if app.Out.JSONMode {
				return app.Out.JSON(s)
			}
			app.Out.Infof("版本:    %s (%s)    Go: %s", s.App.Version, s.App.Commit, s.App.GoVersion)
			app.Out.Infof("运行:    %s    健康模式: %s", s.App.Uptime, s.HealthMode)
			app.Out.Infof("数据库:  %s %s", healthDot(app, s.Database.Status == "connected"), s.Database.Version)
			app.Out.Infof("Redis:   %s %s %s", healthDot(app, s.Redis.Status == "connected"), s.Redis.Version, s.Redis.UsedMemory)
			app.Out.Infof("内存:    %s / %s", output.HumanBytes(s.Memory.Alloc), output.HumanBytes(s.Memory.Sys))
			app.Out.Infof("磁盘:    %s 可用 / %s 总量", output.HumanBytes(s.Disk.Free), output.HumanBytes(s.Disk.All))
			app.Out.Infof("存储:    %s", output.HumanBytes(s.Storage.Size))
			if s.Update.HasUpdate {
				app.Out.Warnf("发现新版本: %s（当前 %s），可在管理后台查看升级指引", s.Update.LatestVersion, s.Update.CurrentVersion)
			}
			if len(s.Components) > 0 {
				rows := make([][]string, 0, len(s.Components))
				for _, c := range s.Components {
					rows = append(rows, []string{c.Name, healthDot(app, c.Healthy), c.Status})
				}
				if err := app.Out.Table([]string{"组件", "健康", "状态"}, rows); err != nil {
					return err
				}
			}
			return nil
		},
	}
}

func healthDot(app *App, ok bool) string {
	if ok {
		return app.Out.Green("●")
	}
	return app.Out.Yellow("●")
}
