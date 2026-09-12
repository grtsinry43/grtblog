package commands

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/grtsinry43/grtblog/cli/v2/internal/client"
	"github.com/grtsinry43/grtblog/cli/v2/internal/output"
)

func newCommentCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "comment",
		Aliases: []string{"comments"},
		Short:   "评论管理（ls / reply / approve / reject / block / rm / viewed）",
	}

	var page, size int
	var status string
	var all bool
	ls := &cobra.Command{
		Use:   "ls",
		Short: "列出评论（默认仅未查看）",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli, err := app.Client()
			if err != nil {
				return err
			}
			q := map[string]string{
				"page":     fmt.Sprint(page),
				"pageSize": fmt.Sprint(size),
			}
			if status != "" {
				q["status"] = status
			}
			if all {
				q["onlyUnviewed"] = "false"
			}
			var list client.CommentList
			if err := cli.Get("/admin/comments", q, &list); err != nil {
				return err
			}
			if app.Out.JSONMode {
				return app.Out.JSON(list)
			}
			rows := make([][]string, 0, len(list.Items))
			for _, c := range list.Items {
				rows = append(rows, []string{
					c.ID,
					deref(c.NickName, "-"),
					output.Trunc(deref(c.AreaTitle, fmt.Sprint(c.AreaID)), 16),
					output.Trunc(deref(c.Content, ""), 40),
					commentStatus(app, c.Status, c.IsViewed),
					output.HumanTime(c.CreatedAt),
				})
			}
			if err := app.Out.Table([]string{"ID", "作者", "评论于", "内容", "状态", "时间"}, rows); err != nil {
				return err
			}
			app.Out.Infof("%s 共 %d 条（第 %d 页）", app.Out.Dim(""), list.Total, list.Page)
			return nil
		},
	}
	ls.Flags().IntVar(&page, "page", 1, "页码")
	ls.Flags().IntVar(&size, "size", 20, "每页数量（最大 100）")
	ls.Flags().StringVar(&status, "status", "", "按状态过滤: pending/approved/rejected/blocked")
	ls.Flags().BoolVar(&all, "all", false, "显示全部（默认仅未查看）")

	reply := &cobra.Command{
		Use:   "reply <id> [内容...]",
		Short: "回复评论",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cli, err := app.Client()
			if err != nil {
				return err
			}
			content, err := readContent(args[1:])
			if err != nil {
				return err
			}
			if strings.TrimSpace(content) == "" {
				return fmt.Errorf("回复内容不能为空")
			}
			if err := cli.Post(fmt.Sprintf("/admin/comments/%s/reply", args[0]), client.CommentReplyReq{Content: content}, nil); err != nil {
				return err
			}
			app.Out.Successf("已回复评论 %s", args[0])
			return nil
		},
	}

	statusCmd := func(use, short, value string) *cobra.Command {
		return &cobra.Command{
			Use:   use + " <id>...",
			Short: short,
			Args:  cobra.MinimumNArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				cli, err := app.Client()
				if err != nil {
					return err
				}
				for _, id := range args {
					if err := cli.Put(fmt.Sprintf("/admin/comments/%s/status", id), client.CommentStatusReq{Status: value}, nil); err != nil {
						return fmt.Errorf("处理 %s 失败: %w", id, err)
					}
				}
				app.Out.Successf("%s：已处理 %d 条", short, len(args))
				return nil
			},
		}
	}

	var yes bool
	rm := &cobra.Command{
		Use:   "rm <id>...",
		Short: "删除评论",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !yes && !app.Out.Confirm("确定删除 %d 条评论？", len(args)) {
				app.Out.Warnf("已取消")
				return nil
			}
			cli, err := app.Client()
			if err != nil {
				return err
			}
			for _, id := range args {
				if err := cli.Delete("/admin/comments/"+id, nil); err != nil {
					return fmt.Errorf("删除 %s 失败: %w", id, err)
				}
			}
			app.Out.Successf("已删除 %d 条评论", len(args))
			return nil
		},
	}
	rm.Flags().BoolVarP(&yes, "yes", "y", false, "跳过确认")

	viewed := &cobra.Command{
		Use:   "viewed <id>...",
		Short: "标记评论为已查看",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cli, err := app.Client()
			if err != nil {
				return err
			}
			mark := true
			if err := cli.Put("/admin/comments/viewed", client.CommentViewedReq{IDs: args, IsViewed: &mark}, nil); err != nil {
				return err
			}
			app.Out.Successf("已标记 %d 条评论为已查看", len(args))
			return nil
		},
	}

	cmd.AddCommand(ls, reply, rm, viewed,
		statusCmd("approve", "通过评论", "approved"),
		statusCmd("reject", "拒绝评论", "rejected"),
		statusCmd("block", "屏蔽评论", "blocked"),
	)
	return cmd
}

func commentStatus(app *App, status string, viewed bool) string {
	s := status
	switch status {
	case "approved":
		s = app.Out.Green("已通过")
	case "pending":
		s = app.Out.Yellow("待审核")
	case "rejected":
		s = "已拒绝"
	case "blocked":
		s = "已屏蔽"
	}
	if !viewed {
		s += " ●"
	}
	return s
}

func deref(p *string, fallback string) string {
	if p == nil || *p == "" {
		return fallback
	}
	return *p
}
