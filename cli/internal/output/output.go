// Package output 负责 CLI 的终端输出：表格、JSON、颜色与交互确认。
package output

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"
	"golang.org/x/term"
)

// Printer 封装输出约定。
type Printer struct {
	JSONMode bool
	NoColor  bool
	Out      io.Writer
	ErrOut   io.Writer
	in       io.Reader
}

// New 创建 Printer。JSONMode 时结构化数据以 JSON 输出。
func New(jsonMode, noColor bool) *Printer {
	if os.Getenv("NO_COLOR") != "" {
		noColor = true
	}
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		noColor = true
	}
	return &Printer{
		JSONMode: jsonMode,
		NoColor:  noColor,
		Out:      os.Stdout,
		ErrOut:   os.Stderr,
		in:       os.Stdin,
	}
}

// JSON 以缩进 JSON 输出数据。
func (p *Printer) JSON(v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("JSON 序列化失败: %w", err)
	}
	_, err = fmt.Fprintln(p.Out, string(data))
	return err
}

// Table 渲染对齐表格。
func (p *Printer) Table(headers []string, rows [][]string) error {
	table := tablewriter.NewTable(p.Out)
	h := make([]any, len(headers))
	for i, v := range headers {
		h[i] = v
	}
	table.Header(h...)
	for _, row := range rows {
		r := make([]any, len(row))
		for i, v := range row {
			r[i] = v
		}
		if err := table.Append(r...); err != nil {
			return err
		}
	}
	return table.Render()
}

// Infof 输出普通信息。
func (p *Printer) Infof(format string, args ...any) {
	fmt.Fprintf(p.Out, format+"\n", args...)
}

// Successf 输出成功信息（绿色 ✓）。
func (p *Printer) Successf(format string, args ...any) {
	if p.NoColor {
		fmt.Fprintf(p.Out, "✓ "+format+"\n", args...)
		return
	}
	fmt.Fprintf(p.Out, "\033[32m✓\033[0m "+format+"\n", args...)
}

// Warnf 输出警告信息（黄色 !）。
func (p *Printer) Warnf(format string, args ...any) {
	if p.NoColor {
		fmt.Fprintf(p.ErrOut, "! "+format+"\n", args...)
		return
	}
	fmt.Fprintf(p.ErrOut, "\033[33m!\033[0m "+format+"\n", args...)
}

// Dim 返回灰色文本（仅终端内）。
func (p *Printer) Dim(s string) string {
	if p.NoColor {
		return s
	}
	return "\033[2m" + s + "\033[0m"
}

// Green 返回绿色文本。
func (p *Printer) Green(s string) string {
	if p.NoColor {
		return s
	}
	return "\033[32m" + s + "\033[0m"
}

// Yellow 返回黄色文本。
func (p *Printer) Yellow(s string) string {
	if p.NoColor {
		return s
	}
	return "\033[33m" + s + "\033[0m"
}

// Confirm 询问用户确认；返回 true 表示继续。
func (p *Printer) Confirm(format string, args ...any) bool {
	fmt.Fprintf(p.ErrOut, "? "+format+" [y/N] ", args...)
	scanner := bufio.NewScanner(p.in)
	if !scanner.Scan() {
		return false
	}
	ans := strings.ToLower(strings.TrimSpace(scanner.Text()))
	return ans == "y" || ans == "yes"
}

// HumanBytes 人性化字节数。
func HumanBytes(n uint64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := uint64(unit), 0
	for v := n / unit; v >= unit; v /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(n)/float64(div), "KMGTPE"[exp])
}

// HumanTime 本地时间格式。
func HumanTime(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	return t.Local().Format("2006-01-02 15:04")
}

// Trunc 按 rune 截断字符串并加省略号。
func Trunc(s string, n int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n-1]) + "…"
}
