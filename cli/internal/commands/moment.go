package commands

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/grtsinry43/grtblog/cli/v2/internal/client"
	"github.com/grtsinry43/grtblog/cli/v2/internal/editor"
	"github.com/grtsinry43/grtblog/cli/v2/internal/frontmatter"
	"github.com/grtsinry43/grtblog/cli/v2/internal/output"
)

// momentFrontMatter 是手记编辑模板的 YAML 头部。
type momentFrontMatter struct {
	Title        string   `yaml:"title"`
	Summary      string   `yaml:"summary"`
	Column       string   `yaml:"column"` // 专栏名称或 ID
	Topics       []string `yaml:"topics"` // 话题（标签）名称或 ID
	Images       []string `yaml:"images"` // 图片 URL 列表
	ShortURL     string   `yaml:"shortUrl"`
	Published    bool     `yaml:"published"`
	Top          bool     `yaml:"top"`
	Original     bool     `yaml:"original"`
	AllowComment bool     `yaml:"allowComment"`
}

func newMomentCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "moment",
		Aliases: []string{"moments"},
		Short:   "手记管理（ls / view / new / edit / rm / publish / top）",
	}

	var page, size int
	var search, column, topic string
	var draft, published bool

	ls := &cobra.Command{
		Use:   "ls",
		Short: "列出手记",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli, err := app.Client()
			if err != nil {
				return err
			}
			q := map[string]string{
				"page":     fmt.Sprint(page),
				"pageSize": fmt.Sprint(size),
			}
			if search != "" {
				q["search"] = search
			}
			if draft {
				q["published"] = "false"
			} else if published {
				q["published"] = "true"
			}
			if column != "" || topic != "" {
				idx, err := loadTaxonomy(cli)
				if err != nil {
					return err
				}
				if column != "" {
					id, err := resolveID(column, idx.columnNames(), "专栏")
					if err != nil {
						return err
					}
					q["columnId"] = fmt.Sprint(id)
				}
				if topic != "" {
					id, err := resolveID(topic, idx.tagNames(), "话题")
					if err != nil {
						return err
					}
					q["topicId"] = fmt.Sprint(id)
				}
			}
			var list client.MomentList
			if err := cli.Get("/admin/moments", q, &list); err != nil {
				return err
			}
			if app.Out.JSONMode {
				return app.Out.JSON(list)
			}
			rows := make([][]string, 0, len(list.Items))
			for _, m := range list.Items {
				rows = append(rows, []string{
					fmt.Sprint(m.ID),
					output.Trunc(m.Title, 30),
					m.ColumnName,
					strings.Join(m.Topics, ","),
					fmt.Sprint(len(m.Image)),
					fmt.Sprint(m.Views),
					articleStatus(app, m.IsPublished, m.IsTop),
					output.HumanTime(m.UpdatedAt),
				})
			}
			if err := app.Out.Table([]string{"ID", "标题", "专栏", "话题", "图", "阅读", "状态", "更新于"}, rows); err != nil {
				return err
			}
			app.Out.Infof("%s 共 %d 篇（第 %d 页）", app.Out.Dim(""), list.Total, list.Page)
			return nil
		},
	}
	ls.Flags().IntVar(&page, "page", 1, "页码")
	ls.Flags().IntVar(&size, "size", 20, "每页数量（最大 100）")
	ls.Flags().StringVar(&search, "search", "", "搜索关键词")
	ls.Flags().StringVar(&column, "column", "", "按专栏过滤（名称或 ID）")
	ls.Flags().StringVar(&topic, "topic", "", "按话题过滤（名称或 ID）")
	ls.Flags().BoolVar(&draft, "draft", false, "仅看草稿")
	ls.Flags().BoolVar(&published, "published", false, "仅看已发布")

	var raw bool
	view := &cobra.Command{
		Use:   "view <id>",
		Short: "查看手记详情与正文",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cli, err := app.Client()
			if err != nil {
				return err
			}
			var m client.Moment
			if err := cli.Get("/admin/moments/"+args[0], nil, &m); err != nil {
				return err
			}
			if app.Out.JSONMode {
				return app.Out.JSON(m)
			}
			if raw {
				app.Out.Infof("%s", m.Content)
				return nil
			}
			printMomentMeta(app, &m)
			renderMarkdown(app, m.Content)
			return nil
		},
	}
	view.Flags().BoolVar(&raw, "raw", false, "输出原始 Markdown")

	var title, file string
	var publish, top bool
	newCmd := &cobra.Command{
		Use:   "new",
		Short: "创建手记（默认打开编辑器，默认保存为草稿）",
		Example: `  grtblog moment new                  # 打开编辑器撰写
  grtblog moment new -f note.md       # 从 Markdown 文件创建
  grtblog moment new -f - < note.md   # 从标准输入创建
  grtblog moment new --publish        # 撰写后直接发布`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cli, err := app.Client()
			if err != nil {
				return err
			}
			fm := momentFrontMatter{Original: true, AllowComment: true}
			var body string
			switch {
			case file == "-":
				data, err := io.ReadAll(os.Stdin)
				if err != nil {
					return err
				}
				if err := parseMomentDoc(data, &fm, &body); err != nil {
					return err
				}
			case file != "":
				data, err := os.ReadFile(file)
				if err != nil {
					return fmt.Errorf("读取文件失败: %w", err)
				}
				if err := parseMomentDoc(data, &fm, &body); err != nil {
					return err
				}
			default:
				fm.Title = title
				if fm.Title == "" {
					fm.Title = "未命名手记"
				}
				tpl := composeMomentDoc(&fm, "\n在这里输入正文（Markdown）…\n")
				edited, changed, err := editor.Edit(tpl, ".md")
				if err != nil {
					return err
				}
				if !changed {
					app.Out.Warnf("内容未修改，已取消")
					return nil
				}
				if err := parseMomentDoc(edited, &fm, &body); err != nil {
					return err
				}
			}
			if title != "" {
				fm.Title = title
			}
			if cmd.Flags().Changed("publish") && publish {
				fm.Published = true
			}
			if top {
				fm.Top = true
			}
			return submitMoment(app, cli, 0, &fm, body)
		},
	}
	newCmd.Flags().StringVar(&title, "title", "", "标题（覆盖 front-matter）")
	newCmd.Flags().StringVarP(&file, "file", "f", "", "从 Markdown 文件创建（- 表示标准输入）")
	newCmd.Flags().BoolVar(&publish, "publish", false, "创建后直接发布")
	newCmd.Flags().BoolVar(&top, "top", false, "置顶")

	edit := &cobra.Command{
		Use:   "edit <id>",
		Short: "编辑手记（打开编辑器，保存后提交）",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cli, err := app.Client()
			if err != nil {
				return err
			}
			var m client.Moment
			if err := cli.Get("/admin/moments/"+args[0], nil, &m); err != nil {
				return err
			}
			fm := momentFrontMatter{
				Title:        m.Title,
				Summary:      m.Summary,
				Column:       m.ColumnName,
				Images:       m.Image,
				ShortURL:     m.ShortURL,
				Published:    m.IsPublished,
				Top:          m.IsTop,
				Original:     m.IsOriginal,
				AllowComment: m.AllowComment,
			}
			for _, t := range m.Topics {
				fm.Topics = append(fm.Topics, t.Name)
			}
			edited, changed, err := editor.Edit(composeMomentDoc(&fm, m.Content), ".md")
			if err != nil {
				return err
			}
			if !changed {
				app.Out.Warnf("内容未修改")
				return nil
			}
			var body string
			if err := parseMomentDoc(edited, &fm, &body); err != nil {
				return err
			}
			if fm.ShortURL == "" {
				fm.ShortURL = m.ShortURL
			}
			return submitMoment(app, cli, m.ID, &fm, body)
		},
	}

	rm := newBatchCmd(app, batchSpec{
		use:     "rm <id>...",
		short:   "删除手记",
		confirm: "确定删除 %d 篇手记？",
		run: func(cli *client.Client, ids []int64) error {
			if len(ids) == 1 {
				return cli.Delete(fmt.Sprintf("/moments/%d", ids[0]), nil)
			}
			return cli.Post("/admin/moments/batch-delete", client.BatchIDsReq{IDs: ids}, nil)
		},
	})

	publishCmd := newBatchCmd(app, batchSpec{
		use:   "publish <id>...",
		short: "发布手记",
		run: func(cli *client.Client, ids []int64) error {
			return cli.Put("/admin/moments/published", client.BatchPublishedReq{IDs: ids, IsPublished: true}, nil)
		},
	})

	unpublish := newBatchCmd(app, batchSpec{
		use:   "unpublish <id>...",
		short: "转为草稿",
		run: func(cli *client.Client, ids []int64) error {
			return cli.Put("/admin/moments/published", client.BatchPublishedReq{IDs: ids, IsPublished: false}, nil)
		},
	})

	topCmd := newBatchCmd(app, batchSpec{
		use:   "top <id>...",
		short: "置顶手记",
		run: func(cli *client.Client, ids []int64) error {
			return cli.Put("/admin/moments/top", client.BatchTopReq{IDs: ids, IsTop: true}, nil)
		},
	})

	untop := newBatchCmd(app, batchSpec{
		use:   "untop <id>...",
		short: "取消置顶",
		run: func(cli *client.Client, ids []int64) error {
			return cli.Put("/admin/moments/top", client.BatchTopReq{IDs: ids, IsTop: false}, nil)
		},
	})

	cmd.AddCommand(ls, view, newCmd, edit, rm, publishCmd, unpublish, topCmd, untop)
	return cmd
}

func parseMomentDoc(data []byte, fm *momentFrontMatter, body *string) error {
	meta, content, has := frontmatter.Split(data)
	if has {
		if err := yaml.Unmarshal(meta, fm); err != nil {
			return fmt.Errorf("front-matter 解析失败: %w", err)
		}
	} else {
		content = data
	}
	*body = strings.TrimSpace(string(content))
	if fm.Title == "" {
		return fmt.Errorf("标题不能为空（请在 front-matter 中设置 title 或使用 --title）")
	}
	if *body == "" {
		return fmt.Errorf("正文不能为空")
	}
	return nil
}

func composeMomentDoc(fm *momentFrontMatter, body string) []byte {
	meta, _ := yaml.Marshal(fm)
	return frontmatter.Compose(meta, body)
}

func submitMoment(app *App, cli *client.Client, id int64, fm *momentFrontMatter, body string) error {
	idx, err := loadTaxonomy(cli)
	if err != nil {
		return err
	}
	req := client.MomentUpsertReq{
		Title:        fm.Title,
		Summary:      fm.Summary,
		Content:      body,
		Image:        fm.Images,
		IsPublished:  fm.Published,
		IsTop:        fm.Top,
		IsOriginal:   fm.Original,
		AllowComment: &fm.AllowComment,
	}
	if fm.Column != "" {
		colID, err := resolveID(fm.Column, idx.columnNames(), "专栏")
		if err != nil {
			return err
		}
		req.ColumnID = &colID
	}
	if len(fm.Topics) > 0 {
		req.TopicIDs, err = resolveIDs(fm.Topics, idx.tagNames(), "话题")
		if err != nil {
			return err
		}
	}
	if fm.ShortURL != "" {
		req.ShortURL = &fm.ShortURL
	}
	if id == 0 {
		var created client.Moment
		if err := cli.Post("/moments", req, &created); err != nil {
			return err
		}
		app.Out.Successf("手记已创建：id=%d 标题=%q 状态=%s", created.ID, created.Title, articleStatus(app, created.IsPublished, created.IsTop))
		return nil
	}
	var updated client.Moment
	if err := cli.Put(fmt.Sprintf("/moments/%d", id), req, &updated); err != nil {
		return err
	}
	app.Out.Successf("手记已更新：id=%d 标题=%q", updated.ID, updated.Title)
	return nil
}

func printMomentMeta(app *App, m *client.Moment) {
	app.Out.Infof("标题:   %s", m.Title)
	app.Out.Infof("ID:     %d    短链: %s    状态: %s", m.ID, m.ShortURL, articleStatus(app, m.IsPublished, m.IsTop))
	topics := make([]string, 0, len(m.Topics))
	for _, t := range m.Topics {
		topics = append(topics, t.Name)
	}
	app.Out.Infof("专栏:   %s    话题: %s", m.ColumnName, strings.Join(topics, ", "))
	if len(m.Image) > 0 {
		app.Out.Infof("图片:   %s", strings.Join(m.Image, ", "))
	}
	app.Out.Infof("更新于: %s", output.HumanTime(m.UpdatedAt))
	fmt.Fprintln(app.Out.Out, strings.Repeat("─", 40))
}
