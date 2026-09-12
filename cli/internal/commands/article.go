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

// articleFrontMatter 是文章编辑模板的 YAML 头部。
type articleFrontMatter struct {
	Title        string   `yaml:"title"`
	Summary      string   `yaml:"summary"`
	Category     string   `yaml:"category"` // 名称或 ID
	Tags         []string `yaml:"tags"`     // 名称或 ID
	ShortURL     string   `yaml:"shortUrl"`
	Published    bool     `yaml:"published"`
	Top          bool     `yaml:"top"`
	Original     bool     `yaml:"original"`
	AllowComment bool     `yaml:"allowComment"`
}

func newArticleCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "article",
		Aliases: []string{"articles"},
		Short:   "文章管理（ls / view / new / edit / rm / publish / top）",
	}

	var page, size int
	var search, category, tag string
	var draft, published bool

	ls := &cobra.Command{
		Use:   "ls",
		Short: "列出文章",
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
			if category != "" || tag != "" {
				idx, err := loadTaxonomy(cli)
				if err != nil {
					return err
				}
				if category != "" {
					id, err := resolveID(category, idx.categoryNames(), "分类")
					if err != nil {
						return err
					}
					q["categoryId"] = fmt.Sprint(id)
				}
				if tag != "" {
					id, err := resolveID(tag, idx.tagNames(), "标签")
					if err != nil {
						return err
					}
					q["tagId"] = fmt.Sprint(id)
				}
			}
			var list client.ArticleList
			if err := cli.Get("/admin/articles", q, &list); err != nil {
				return err
			}
			if app.Out.JSONMode {
				return app.Out.JSON(list)
			}
			rows := make([][]string, 0, len(list.Items))
			for _, a := range list.Items {
				rows = append(rows, []string{
					fmt.Sprint(a.ID),
					output.Trunc(a.Title, 30),
					a.CategoryName,
					strings.Join(a.Tags, ","),
					fmt.Sprint(a.Views),
					articleStatus(app, a.IsPublished, a.IsTop),
					output.HumanTime(a.UpdatedAt),
				})
			}
			if err := app.Out.Table([]string{"ID", "标题", "分类", "标签", "阅读", "状态", "更新于"}, rows); err != nil {
				return err
			}
			app.Out.Infof("%s 共 %d 篇（第 %d 页）", app.Out.Dim(""), list.Total, list.Page)
			return nil
		},
	}
	ls.Flags().IntVar(&page, "page", 1, "页码")
	ls.Flags().IntVar(&size, "size", 20, "每页数量（最大 100）")
	ls.Flags().StringVar(&search, "search", "", "搜索关键词")
	ls.Flags().StringVar(&category, "category", "", "按分类过滤（名称或 ID）")
	ls.Flags().StringVar(&tag, "tag", "", "按标签过滤（名称或 ID）")
	ls.Flags().BoolVar(&draft, "draft", false, "仅看草稿")
	ls.Flags().BoolVar(&published, "published", false, "仅看已发布")

	var raw bool
	view := &cobra.Command{
		Use:   "view <id>",
		Short: "查看文章详情与正文",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cli, err := app.Client()
			if err != nil {
				return err
			}
			var a client.Article
			if err := cli.Get("/admin/articles/"+args[0], nil, &a); err != nil {
				return err
			}
			if app.Out.JSONMode {
				return app.Out.JSON(a)
			}
			if raw {
				app.Out.Infof("%s", a.Content)
				return nil
			}
			printArticleMeta(app, &a)
			renderMarkdown(app, a.Content)
			return nil
		},
	}
	view.Flags().BoolVar(&raw, "raw", false, "输出原始 Markdown")

	var title, file string
	var publish, top bool
	newCmd := &cobra.Command{
		Use:   "new",
		Short: "创建文章（默认打开编辑器，默认保存为草稿）",
		Example: `  grtblog article new                  # 打开编辑器撰写
  grtblog article new -f post.md       # 从 Markdown 文件创建（支持 front-matter）
  grtblog article new -f - < post.md   # 从标准输入创建
  grtblog article new --publish        # 撰写后直接发布`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cli, err := app.Client()
			if err != nil {
				return err
			}
			fm := articleFrontMatter{Original: true, AllowComment: true}
			var body string
			switch {
			case file == "-":
				data, err := io.ReadAll(os.Stdin)
				if err != nil {
					return err
				}
				if err := parseArticleDoc(data, &fm, &body); err != nil {
					return err
				}
			case file != "":
				data, err := os.ReadFile(file)
				if err != nil {
					return fmt.Errorf("读取文件失败: %w", err)
				}
				if err := parseArticleDoc(data, &fm, &body); err != nil {
					return err
				}
			default:
				fm.Title = title
				if fm.Title == "" {
					fm.Title = "未命名文章"
				}
				tpl := composeArticleDoc(&fm, "\n在这里输入正文（Markdown）…\n")
				edited, changed, err := editor.Edit(tpl, ".md")
				if err != nil {
					return err
				}
				if !changed {
					app.Out.Warnf("内容未修改，已取消")
					return nil
				}
				if err := parseArticleDoc(edited, &fm, &body); err != nil {
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
			return submitArticle(app, cli, 0, &fm, body)
		},
	}
	newCmd.Flags().StringVar(&title, "title", "", "标题（覆盖 front-matter）")
	newCmd.Flags().StringVarP(&file, "file", "f", "", "从 Markdown 文件创建（- 表示标准输入）")
	newCmd.Flags().BoolVar(&publish, "publish", false, "创建后直接发布")
	newCmd.Flags().BoolVar(&top, "top", false, "置顶")

	edit := &cobra.Command{
		Use:   "edit <id>",
		Short: "编辑文章（打开编辑器，保存后提交）",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cli, err := app.Client()
			if err != nil {
				return err
			}
			var a client.Article
			if err := cli.Get("/admin/articles/"+args[0], nil, &a); err != nil {
				return err
			}
			fm := articleFrontMatter{
				Title:        a.Title,
				Summary:      a.Summary,
				Category:     a.CategoryName,
				ShortURL:     a.ShortURL,
				Published:    a.IsPublished,
				Top:          a.IsTop,
				Original:     a.IsOriginal,
				AllowComment: a.AllowComment,
			}
			for _, t := range a.Tags {
				fm.Tags = append(fm.Tags, t.Name)
			}
			edited, changed, err := editor.Edit(composeArticleDoc(&fm, a.Content), ".md")
			if err != nil {
				return err
			}
			if !changed {
				app.Out.Warnf("内容未修改")
				return nil
			}
			var body string
			if err := parseArticleDoc(edited, &fm, &body); err != nil {
				return err
			}
			if fm.ShortURL == "" {
				fm.ShortURL = a.ShortURL // 更新接口要求 shortUrl 必填
			}
			return submitArticle(app, cli, a.ID, &fm, body)
		},
	}

	rm := newBatchCmd(app, batchSpec{
		use:     "rm <id>...",
		short:   "删除文章",
		confirm: "确定删除 %d 篇文章？",
		run: func(cli *client.Client, ids []int64) error {
			if len(ids) == 1 {
				return cli.Delete(fmt.Sprintf("/articles/%d", ids[0]), nil)
			}
			return cli.Post("/admin/articles/batch-delete", client.BatchIDsReq{IDs: ids}, nil)
		},
	})

	publishCmd := newBatchCmd(app, batchSpec{
		use:   "publish <id>...",
		short: "发布文章",
		run: func(cli *client.Client, ids []int64) error {
			return cli.Put("/admin/articles/published", client.BatchPublishedReq{IDs: ids, IsPublished: true}, nil)
		},
	})

	unpublish := newBatchCmd(app, batchSpec{
		use:   "unpublish <id>...",
		short: "转为草稿",
		run: func(cli *client.Client, ids []int64) error {
			return cli.Put("/admin/articles/published", client.BatchPublishedReq{IDs: ids, IsPublished: false}, nil)
		},
	})

	topCmd := newBatchCmd(app, batchSpec{
		use:   "top <id>...",
		short: "置顶文章",
		run: func(cli *client.Client, ids []int64) error {
			return cli.Put("/admin/articles/top", client.BatchTopReq{IDs: ids, IsTop: true}, nil)
		},
	})

	untop := newBatchCmd(app, batchSpec{
		use:   "untop <id>...",
		short: "取消置顶",
		run: func(cli *client.Client, ids []int64) error {
			return cli.Put("/admin/articles/top", client.BatchTopReq{IDs: ids, IsTop: false}, nil)
		},
	})

	cmd.AddCommand(ls, view, newCmd, edit, rm, publishCmd, unpublish, topCmd, untop)
	return cmd
}

// batchSpec 描述一个批量 ID 操作命令。
type batchSpec struct {
	use     string
	short   string
	confirm string // 非空时需确认（可用 -y 跳过）
	run     func(cli *client.Client, ids []int64) error
}

func newBatchCmd(app *App, spec batchSpec) *cobra.Command {
	var yes bool
	cmd := &cobra.Command{
		Use:   spec.use,
		Short: spec.short,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ids, err := parseIDs(args)
			if err != nil {
				return err
			}
			if spec.confirm != "" && !yes {
				if !app.Out.Confirm(spec.confirm, len(ids)) {
					app.Out.Warnf("已取消")
					return nil
				}
			}
			cli, err := app.Client()
			if err != nil {
				return err
			}
			if err := spec.run(cli, ids); err != nil {
				return err
			}
			app.Out.Successf("%s：已处理 %d 项", spec.short, len(ids))
			return nil
		},
	}
	if spec.confirm != "" {
		cmd.Flags().BoolVarP(&yes, "yes", "y", false, "跳过确认")
	}
	return cmd
}

// parseArticleDoc 解析 Markdown 文档（可含 front-matter）。
func parseArticleDoc(data []byte, fm *articleFrontMatter, body *string) error {
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

// composeArticleDoc 由 front-matter + 正文生成编辑模板。
func composeArticleDoc(fm *articleFrontMatter, body string) []byte {
	meta, _ := yaml.Marshal(fm)
	return frontmatter.Compose(meta, body)
}

// submitArticle 创建（id=0）或更新文章。
func submitArticle(app *App, cli *client.Client, id int64, fm *articleFrontMatter, body string) error {
	idx, err := loadTaxonomy(cli)
	if err != nil {
		return err
	}
	req := client.ArticleUpsertReq{
		Title:        fm.Title,
		Summary:      fm.Summary,
		Content:      body,
		IsPublished:  fm.Published,
		IsTop:        fm.Top,
		IsOriginal:   fm.Original,
		AllowComment: &fm.AllowComment,
	}
	if fm.Category != "" {
		catID, err := resolveID(fm.Category, idx.categoryNames(), "分类")
		if err != nil {
			return err
		}
		req.CategoryID = &catID
	}
	if len(fm.Tags) > 0 {
		req.TagIDs, err = resolveIDs(fm.Tags, idx.tagNames(), "标签")
		if err != nil {
			return err
		}
	}
	if fm.ShortURL != "" {
		req.ShortURL = &fm.ShortURL
	}
	if id == 0 {
		var created client.Article
		if err := cli.Post("/articles", req, &created); err != nil {
			return err
		}
		app.Out.Successf("文章已创建：id=%d 标题=%q 状态=%s", created.ID, created.Title, articleStatus(app, created.IsPublished, created.IsTop))
		return nil
	}
	var updated client.Article
	if err := cli.Put(fmt.Sprintf("/articles/%d", id), req, &updated); err != nil {
		return err
	}
	app.Out.Successf("文章已更新：id=%d 标题=%q", updated.ID, updated.Title)
	return nil
}

func articleStatus(app *App, published, top bool) string {
	var s string
	if published {
		s = app.Out.Green("已发布")
	} else {
		s = app.Out.Yellow("草稿")
	}
	if top {
		s += " 📌"
	}
	return s
}

func printArticleMeta(app *App, a *client.Article) {
	app.Out.Infof("标题:   %s", a.Title)
	app.Out.Infof("ID:     %d    短链: %s    状态: %s", a.ID, a.ShortURL, articleStatus(app, a.IsPublished, a.IsTop))
	tags := make([]string, 0, len(a.Tags))
	for _, t := range a.Tags {
		tags = append(tags, t.Name)
	}
	app.Out.Infof("分类:   %s    标签: %s", a.CategoryName, strings.Join(tags, ", "))
	app.Out.Infof("更新于: %s", output.HumanTime(a.UpdatedAt))
	fmt.Fprintln(app.Out.Out, strings.Repeat("─", 40))
}
