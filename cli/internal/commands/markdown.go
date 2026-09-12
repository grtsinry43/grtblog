package commands

import (
	"fmt"

	"github.com/charmbracelet/glamour"
)

// renderMarkdown 在终端渲染 Markdown（失败时退化为原文输出）。
func renderMarkdown(app *App, content string) {
	if app.Out.NoColor {
		fmt.Fprintln(app.Out.Out, content)
		return
	}
	r, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(100),
	)
	if err != nil {
		fmt.Fprintln(app.Out.Out, content)
		return
	}
	out, err := r.Render(content)
	if err != nil {
		fmt.Fprintln(app.Out.Out, content)
		return
	}
	fmt.Fprint(app.Out.Out, out)
}
