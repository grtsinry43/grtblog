package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/grtsinry43/grtblog/cli/v2/internal/client"
	"github.com/grtsinry43/grtblog/cli/v2/internal/output"
)

func newUploadCmd(app *App) *cobra.Command {
	var markdown bool
	cmd := &cobra.Command{
		Use:   "upload <file>...",
		Short: "上传文件到媒体库，输出访问链接",
		Example: `  grtblog upload ./cover.png            # 输出文件 → URL
  grtblog upload a.png b.jpg --markdown # 输出 Markdown 图片语法`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cli, err := app.Client()
			if err != nil {
				return err
			}
			type result struct {
				File string `json:"file"`
				ID   int64  `json:"id"`
				URL  string `json:"url"`
			}
			results := make([]result, 0, len(args))
			for _, path := range args {
				if _, err := os.Stat(path); err != nil {
					return fmt.Errorf("文件不存在: %s", path)
				}
				var up client.UploadFile
				if err := cli.Upload("/upload", "file", path, &up); err != nil {
					return fmt.Errorf("上传 %s 失败: %w", path, err)
				}
				results = append(results, result{File: path, ID: up.ID, URL: up.PublicURL})
			}
			if app.Out.JSONMode {
				return app.Out.JSON(results)
			}
			for _, r := range results {
				if markdown {
					fmt.Fprintf(app.Out.Out, "![%s](%s)\n", strings.TrimSuffix(filepath.Base(r.File), filepath.Ext(r.File)), r.URL)
				} else {
					app.Out.Successf("%s → %s", r.File, r.URL)
				}
			}
			return nil
		},
	}
	cmd.Flags().BoolVarP(&markdown, "markdown", "m", false, "以 Markdown 图片语法输出")
	return cmd
}

func newFileCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "file",
		Aliases: []string{"files", "uploads"},
		Short:   "媒体库文件管理（ls / rename / rm / download / sync）",
	}

	var page, size int
	ls := &cobra.Command{
		Use:   "ls",
		Short: "列出媒体库文件",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli, err := app.Client()
			if err != nil {
				return err
			}
			var list client.UploadFileList
			if err := cli.Get("/uploads", map[string]string{
				"page":     fmt.Sprint(page),
				"pageSize": fmt.Sprint(size),
			}, &list); err != nil {
				return err
			}
			if app.Out.JSONMode {
				return app.Out.JSON(list)
			}
			rows := make([][]string, 0, len(list.Items))
			for _, f := range list.Items {
				rows = append(rows, []string{
					fmt.Sprint(f.ID),
					output.Trunc(f.Name, 36),
					f.Type,
					output.HumanBytes(uint64(f.Size)),
					output.HumanTime(f.CreatedAt),
					f.PublicURL,
				})
			}
			if err := app.Out.Table([]string{"ID", "文件名", "类型", "大小", "上传于", "URL"}, rows); err != nil {
				return err
			}
			app.Out.Infof("%s 共 %d 个文件（第 %d 页）", app.Out.Dim(""), list.Total, list.Page)
			return nil
		},
	}
	ls.Flags().IntVar(&page, "page", 1, "页码")
	ls.Flags().IntVar(&size, "size", 20, "每页数量（最大 100）")

	rename := &cobra.Command{
		Use:   "rename <id> <新文件名>",
		Short: "重命名文件",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cli, err := app.Client()
			if err != nil {
				return err
			}
			if err := cli.Put("/upload/"+args[0], client.UploadRenameReq{Name: args[1]}, nil); err != nil {
				return err
			}
			app.Out.Successf("文件 %s 已重命名为 %q", args[0], args[1])
			return nil
		},
	}

	var yes bool
	rm := &cobra.Command{
		Use:   "rm <id>...",
		Short: "删除文件",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !yes && !app.Out.Confirm("确定删除 %d 个文件？", len(args)) {
				app.Out.Warnf("已取消")
				return nil
			}
			cli, err := app.Client()
			if err != nil {
				return err
			}
			for _, id := range args {
				if err := cli.Delete("/upload/"+id, nil); err != nil {
					return fmt.Errorf("删除 %s 失败: %w", id, err)
				}
			}
			app.Out.Successf("已删除 %d 个文件", len(args))
			return nil
		},
	}
	rm.Flags().BoolVarP(&yes, "yes", "y", false, "跳过确认")

	var outPath string
	download := &cobra.Command{
		Use:   "download <id>",
		Short: "下载文件",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cli, err := app.Client()
			if err != nil {
				return err
			}
			dest := outPath
			if dest == "" {
				dest = fmt.Sprintf("upload-%s", args[0])
			}
			// dest 是已存在目录时，先下载为临时名，再按服务端文件名落位
			if info, err := os.Stat(dest); err == nil && info.IsDir() {
				tmp := filepath.Join(dest, fmt.Sprintf(".grtblog-dl-%s", args[0]))
				name, err := cli.Download("/upload/"+args[0]+"/download", tmp)
				if err != nil {
					os.Remove(tmp)
					return err
				}
				if name == "" {
					name = fmt.Sprintf("upload-%s", args[0])
				}
				final := filepath.Join(dest, filepath.Base(name))
				if err := os.Rename(tmp, final); err != nil {
					return err
				}
				dest = final
			} else {
				if _, err := cli.Download("/upload/"+args[0]+"/download", dest); err != nil {
					return err
				}
			}
			app.Out.Successf("已下载到 %s", dest)
			return nil
		},
	}
	download.Flags().StringVarP(&outPath, "out", "o", "", "保存路径（文件或目录）")

	sync := &cobra.Command{
		Use:   "sync",
		Short: "同步磁盘文件到上传索引",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli, err := app.Client()
			if err != nil {
				return err
			}
			var res client.UploadSyncResult
			if err := cli.Post("/uploads/sync", nil, &res); err != nil {
				return err
			}
			if app.Out.JSONMode {
				return app.Out.JSON(res)
			}
			app.Out.Successf("同步完成：扫描 %d，新增 %d，更新 %d，清理 %d，跳过重复 %d",
				res.Scanned, res.Created, res.Updated, res.Deleted, res.SkippedDuplicates)
			return nil
		},
	}

	cmd.AddCommand(ls, rename, rm, download, sync)
	return cmd
}
