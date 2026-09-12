// Package editor 实现 $EDITOR 编辑工作流。
package editor

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Edit 将 initial 写入临时文件，打开编辑器，返回编辑后的内容与是否有改动。
// suffix 形如 ".md"，用于让编辑器启用对应语法高亮。
func Edit(initial []byte, suffix string) (edited []byte, changed bool, err error) {
	tmp, err := os.CreateTemp("", "grtblog-*"+suffix)
	if err != nil {
		return nil, false, fmt.Errorf("创建临时文件失败: %w", err)
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(initial); err != nil {
		tmp.Close()
		return nil, false, fmt.Errorf("写入临时文件失败: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return nil, false, err
	}

	name, args := resolveEditor()
	args = append(args, tmp.Name())
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return nil, false, fmt.Errorf("编辑器退出异常: %w", err)
	}

	edited, err = os.ReadFile(tmp.Name())
	if err != nil {
		return nil, false, fmt.Errorf("读取临时文件失败: %w", err)
	}
	return edited, !bytes.Equal(edited, initial), nil
}

// resolveEditor 按 GRTBLOG_EDITOR > VISUAL > EDITOR > vi 解析编辑器。
// 支持带参数的形式（如 "code --wait"）。
func resolveEditor() (string, []string) {
	for _, env := range []string{"GRTBLOG_EDITOR", "VISUAL", "EDITOR"} {
		if v := strings.TrimSpace(os.Getenv(env)); v != "" {
			parts := strings.Fields(v)
			return parts[0], parts[1:]
		}
	}
	return "vi", nil
}
