package commands

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

func newVersionCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "显示版本信息",
		Run: func(cmd *cobra.Command, args []string) {
			v := app.Info.Version
			if v == "" {
				v = "dev"
			}
			fmt.Printf("grtblog %s\n", v)
			if app.Info.Commit != "" {
				fmt.Printf("commit: %s\n", app.Info.Commit)
			}
			if app.Info.Date != "" {
				fmt.Printf("built:  %s\n", app.Info.Date)
			}
			fmt.Printf("go:     %s\n", runtime.Version())
		},
	}
}
