package commands

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/grtsinry43/grtblog/cli/v2/internal/client"
	"github.com/grtsinry43/grtblog/cli/v2/internal/editor"
	"github.com/grtsinry43/grtblog/cli/v2/internal/output"
)

func newThinkingCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "thinking [内容...]",
		Aliases: []string{"thinkings"},
		Short:   "思考管理（ls / new / edit / rm）",
		Long: `思考管理。

不带子命令时，直接跟内容即可快速发布一条思考：
  grtblog thinking "今天天气不错"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}
			cli, err := app.Client()
			if err != nil {
				return err
			}
			return postThinking(app, cli, strings.Join(args, " "), true)
		},
	}

	var page, size int
	ls := &cobra.Command{
		Use:   "ls",
		Short: "列出思考",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli, err := app.Client()
			if err != nil {
				return err
			}
			var list client.ThinkingList
			if err := cli.Get("/thinkings", map[string]string{
				"page":     fmt.Sprint(page),
				"pageSize": fmt.Sprint(size),
			}, &list); err != nil {
				return err
			}
			if app.Out.JSONMode {
				return app.Out.JSON(list)
			}
			rows := make([][]string, 0, len(list.Items))
			for _, t := range list.Items {
				rows = append(rows, []string{
					fmt.Sprint(t.ID),
					output.Trunc(t.Content, 50),
					fmt.Sprint(t.Views),
					fmt.Sprint(t.Likes),
					fmt.Sprint(t.Comments),
					output.HumanTime(t.CreatedAt),
				})
			}
			if err := app.Out.Table([]string{"ID", "内容", "阅读", "赞", "评论", "发布于"}, rows); err != nil {
				return err
			}
			app.Out.Infof("%s 共 %d 条（第 %d 页）", app.Out.Dim(""), list.Total, page)
			return nil
		},
	}
	ls.Flags().IntVar(&page, "page", 1, "页码")
	ls.Flags().IntVar(&size, "size", 20, "每页数量（最大 100）")

	var noComment bool
	newCmd := &cobra.Command{
		Use:   "new [内容...]",
		Short: "发布一条思考",
		Example: `  grtblog thinking "今天天气不错"
  echo "多行内容" | grtblog thinking
  grtblog thinking          # 打开编辑器`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cli, err := app.Client()
			if err != nil {
				return err
			}
			content, err := readContent(args)
			if err != nil {
				return err
			}
			if strings.TrimSpace(content) == "" {
				return fmt.Errorf("内容不能为空")
			}
			return postThinking(app, cli, content, !noComment)
		},
	}
	newCmd.Flags().BoolVar(&noComment, "no-comment", false, "关闭评论")

	edit := &cobra.Command{
		Use:   "edit <id>",
		Short: "编辑思考内容（打开编辑器）",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cli, err := app.Client()
			if err != nil {
				return err
			}
			var t client.Thinking
			if err := cli.Get("/thinkings/"+args[0], nil, &t); err != nil {
				return err
			}
			edited, changed, err := editor.Edit([]byte(t.Content), ".md")
			if err != nil {
				return err
			}
			if !changed {
				app.Out.Warnf("内容未修改")
				return nil
			}
			content := strings.TrimSpace(string(edited))
			if content == "" {
				return fmt.Errorf("内容不能为空")
			}
			if err := cli.Put("/thinkings/"+args[0], client.ThinkingUpsertReq{Content: content}, nil); err != nil {
				return err
			}
			app.Out.Successf("思考已更新：id=%s", args[0])
			return nil
		},
	}

	rm := newBatchCmd(app, batchSpec{
		use:     "rm <id>...",
		short:   "删除思考",
		confirm: "确定删除 %d 条思考？",
		run: func(cli *client.Client, ids []int64) error {
			if len(ids) == 1 {
				return cli.Delete(fmt.Sprintf("/thinkings/%d", ids[0]), nil)
			}
			return cli.Post("/admin/thinkings/batch-delete", client.BatchIDsReq{IDs: ids}, nil)
		},
	})

	cmd.AddCommand(ls, newCmd, edit, rm)
	return cmd
}

// postThinking 发布一条思考。
func postThinking(app *App, cli *client.Client, content string, allowComment bool) error {
	req := client.ThinkingUpsertReq{Content: content, AllowComment: &allowComment}
	var created client.Thinking
	if err := cli.Post("/thinkings", req, &created); err != nil {
		return err
	}
	app.Out.Successf("思考已发布：id=%d", created.ID)
	return nil
}

// readContent 依次尝试：命令行参数 > 管道输入 > 编辑器。
func readContent(args []string) (string, error) {
	if len(args) > 0 {
		return strings.Join(args, " "), nil
	}
	if stat, _ := os.Stdin.Stat(); stat.Mode()&os.ModeCharDevice == 0 {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(data)), nil
	}
	edited, changed, err := editor.Edit(nil, ".md")
	if err != nil {
		return "", err
	}
	if !changed {
		return "", fmt.Errorf("内容为空，已取消")
	}
	return strings.TrimSpace(string(edited)), nil
}
