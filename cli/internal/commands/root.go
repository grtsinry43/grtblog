package commands

import (
	"github.com/spf13/cobra"
)

// NewRootCmd 构建根命令并挂载全部子命令。
func NewRootCmd(info BuildInfo) *cobra.Command {
	app := &App{Info: info}

	root := &cobra.Command{
		Use:   "grtblog",
		Short: "grtblog 博客系统的命令行管理工具",
		Long: `grtblog 是 grtblog 博客系统的命令行管理工具。

使用管理员令牌（admin token，形如 gt_xxx）连接你的博客，
即可在终端里管理文章、手记、思考、评论、文件与设置项。

快速开始:
  grtblog auth login --server https://your.blog --token gt_xxx
  grtblog article ls
  grtblog thinking "今天天气不错"

配置文件位于 ~/.config/grtblog/config.yaml，支持 --profile 管理多个站点。`,
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return app.Init()
		},
	}

	flags := root.PersistentFlags()
	flags.StringVar(&app.flagServer, "server", "", "服务器地址（或环境变量 GRTBLOG_SERVER）")
	flags.StringVar(&app.flagToken, "token", "", "管理员令牌（或环境变量 GRTBLOG_TOKEN）")
	flags.StringVar(&app.flagProfile, "profile", "", "使用指定配置档案（或环境变量 GRTBLOG_PROFILE）")
	flags.BoolVar(&app.flagJSON, "json", false, "以 JSON 格式输出")
	flags.BoolVar(&app.flagNoColor, "no-color", false, "禁用颜色输出")

	root.AddCommand(
		newAuthCmd(app),
		newArticleCmd(app),
		newMomentCmd(app),
		newThinkingCmd(app),
		newUploadCmd(app),
		newFileCmd(app),
		newCommentCmd(app),
		newConfigCmd(app),
		newTaxonomyCmd(app, "category"),
		newTaxonomyCmd(app, "column"),
		newTaxonomyCmd(app, "tag"),
		newStatusCmd(app),
		newVersionCmd(app),
	)
	return root
}
