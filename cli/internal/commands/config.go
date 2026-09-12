package commands

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/grtsinry43/grtblog/cli/v2/internal/client"
	"github.com/grtsinry43/grtblog/cli/v2/internal/output"
)

func newConfigCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "站点设置项管理（ls / get / set）",
		Long: `查看与修改站点的简单设置项。

复杂配置（联邦、邮件、AI、OAuth 等）请仍在管理后台修改。`,
	}

	var group string
	ls := &cobra.Command{
		Use:   "ls",
		Short: "列出全部设置项",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli, err := app.Client()
			if err != nil {
				return err
			}
			var tree client.SysConfigTree
			if err := cli.Get("/admin/sysconfig", nil, &tree); err != nil {
				return err
			}
			items := flattenConfig(tree)
			if app.Out.JSONMode {
				return app.Out.JSON(items)
			}
			rows := make([][]string, 0, len(items))
			for _, it := range items {
				if group != "" && !strings.HasPrefix(it.GroupPath, group) {
					continue
				}
				rows = append(rows, []string{
					it.Key,
					it.Label,
					it.GroupPath,
					it.ValueType,
					configValue(it, 40),
				})
			}
			return app.Out.Table([]string{"KEY", "名称", "分组", "类型", "值"}, rows)
		},
	}
	ls.Flags().StringVar(&group, "group", "", "按分组前缀过滤")

	get := &cobra.Command{
		Use:   "get <key>",
		Short: "查看单个设置项",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cli, err := app.Client()
			if err != nil {
				return err
			}
			var tree client.SysConfigTree
			if err := cli.Get("/admin/sysconfig", map[string]string{"keys": args[0]}, &tree); err != nil {
				return err
			}
			items := flattenConfig(tree)
			for _, it := range items {
				if it.Key == args[0] {
					if app.Out.JSONMode {
						return app.Out.JSON(it)
					}
					app.Out.Infof("Key:   %s", it.Key)
					app.Out.Infof("名称:  %s", it.Label)
					if it.Description != "" {
						app.Out.Infof("说明:  %s", it.Description)
					}
					app.Out.Infof("类型:  %s", it.ValueType)
					app.Out.Infof("值:    %s", configValue(it, 0))
					return nil
				}
			}
			return fmt.Errorf("设置项 %q 不存在", args[0])
		},
	}

	set := &cobra.Command{
		Use:   "set <key> <value>",
		Short: "修改设置项（value 支持 JSON，无法解析时按字符串处理）",
		Example: `  grtblog config set site.title "我的博客"
  grtblog config set site.footer.icp '{"text":"京ICP备xxx","url":"https://beian.miit.gov.cn"}'`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cli, err := app.Client()
			if err != nil {
				return err
			}
			var value any
			if err := json.Unmarshal([]byte(args[1]), &value); err != nil {
				value = args[1] // 非 JSON 按字符串处理
			}
			req := client.SysConfigBatchUpdateReq{
				Items: []client.SysConfigUpdateItem{{Key: args[0], Value: value}},
			}
			if err := cli.Put("/admin/sysconfig", req, nil); err != nil {
				return err
			}
			app.Out.Successf("已更新 %s = %s", args[0], args[1])
			return nil
		},
	}

	cmd.AddCommand(ls, get, set)
	return cmd
}

// flattenConfig 将配置树拍平为配置项列表。
func flattenConfig(tree client.SysConfigTree) []client.SysConfigItem {
	var out []client.SysConfigItem
	out = append(out, tree.Items...)
	var walk func(groups []client.SysConfigGroup)
	walk = func(groups []client.SysConfigGroup) {
		for _, g := range groups {
			out = append(out, g.Items...)
			walk(g.Children)
		}
	}
	walk(tree.Groups)
	return out
}

// configValue 格式化配置值；敏感值打码，limit>0 时截断。
func configValue(it client.SysConfigItem, limit int) string {
	if it.IsSensitive {
		return "******"
	}
	if it.Value == nil {
		return "-"
	}
	var s string
	switch v := it.Value.(type) {
	case string:
		s = v
	default:
		b, err := json.Marshal(v)
		if err != nil {
			s = fmt.Sprint(v)
		} else {
			s = string(b)
		}
	}
	if limit > 0 {
		return output.Trunc(s, limit)
	}
	return s
}
