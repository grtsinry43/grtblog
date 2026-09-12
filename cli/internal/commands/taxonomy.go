package commands

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/grtsinry43/grtblog/cli/v2/internal/client"
)

// newTaxonomyCmd 生成 category / column / tag 三个结构一致的命令。
// kind 取值: "category" | "column" | "tag"
func newTaxonomyCmd(app *App, kind string) *cobra.Command {
	var label, plural, base string
	switch kind {
	case "category":
		label, plural, base = "分类", "categories", "/categories"
	case "column":
		label, plural, base = "专栏", "columns", "/columns"
	default:
		label, plural, base = "标签", "tags", "/tags"
	}
	adminBase := "/admin/" + plural

	cmd := &cobra.Command{
		Use:     kind,
		Aliases: []string{plural},
		Short:   fmt.Sprintf("%s管理（ls / new / rm）", label),
	}

	ls := &cobra.Command{
		Use:   "ls",
		Short: "列出" + label,
		RunE: func(cmd *cobra.Command, args []string) error {
			cli, err := app.Client()
			if err != nil {
				return err
			}
			if kind == "tag" {
				var tags []client.Tag
				if err := cli.Get(base, nil, &tags); err != nil {
					return err
				}
				if app.Out.JSONMode {
					return app.Out.JSON(tags)
				}
				rows := make([][]string, 0, len(tags))
				for _, t := range tags {
					rows = append(rows, []string{fmt.Sprint(t.ID), t.Name, fmt.Sprint(t.ArticleCount)})
				}
				return app.Out.Table([]string{"ID", "名称", "文章数"}, rows)
			}
			if kind == "category" {
				var items []client.Category
				if err := cli.Get(base, nil, &items); err != nil {
					return err
				}
				if app.Out.JSONMode {
					return app.Out.JSON(items)
				}
				rows := make([][]string, 0, len(items))
				for _, c := range items {
					rows = append(rows, []string{fmt.Sprint(c.ID), c.Name, c.ShortURL})
				}
				return app.Out.Table([]string{"ID", "名称", "短链"}, rows)
			}
			var items []client.Column
			if err := cli.Get(base, nil, &items); err != nil {
				return err
			}
			if app.Out.JSONMode {
				return app.Out.JSON(items)
			}
			rows := make([][]string, 0, len(items))
			for _, c := range items {
				rows = append(rows, []string{fmt.Sprint(c.ID), c.Name, c.ShortURL})
			}
			return app.Out.Table([]string{"ID", "名称", "短链"}, rows)
		},
	}

	var shortURL string
	newCmd := &cobra.Command{
		Use:   "new <名称>",
		Short: "创建" + label,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cli, err := app.Client()
			if err != nil {
				return err
			}
			name := strings.TrimSpace(args[0])
			if name == "" {
				return fmt.Errorf("名称不能为空")
			}
			if kind == "tag" {
				if err := cli.Post(adminBase, client.TagReq{Name: name}, nil); err != nil {
					return err
				}
			} else {
				req := client.TaxonomyReq{Name: name}
				if shortURL != "" {
					req.ShortURL = &shortURL
				}
				if err := cli.Post(adminBase, req, nil); err != nil {
					return err
				}
			}
			app.Out.Successf("%s %q 已创建", label, name)
			return nil
		},
	}
	if kind != "tag" {
		newCmd.Flags().StringVar(&shortURL, "short-url", "", "自定义短链")
	}

	rm := &cobra.Command{
		Use:   "rm <id>...",
		Short: "删除" + label,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			yes, _ := cmd.Flags().GetBool("yes")
			if !yes && !app.Out.Confirm("确定删除 %d 个%s？", len(args), label) {
				app.Out.Warnf("已取消")
				return nil
			}
			cli, err := app.Client()
			if err != nil {
				return err
			}
			for _, id := range args {
				if err := cli.Delete(adminBase+"/"+id, nil); err != nil {
					return fmt.Errorf("删除 %s 失败: %w", id, err)
				}
			}
			app.Out.Successf("已删除 %d 个%s", len(args), label)
			return nil
		},
	}
	rm.Flags().BoolP("yes", "y", false, "跳过确认")

	cmd.AddCommand(ls, newCmd, rm)
	return cmd
}
