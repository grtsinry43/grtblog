// Package frontmatter 处理 Markdown 文件的 YAML front-matter 拆分与合成。
package frontmatter

import (
	"bytes"
)

var delimiter = []byte("---")

// Split 拆分 front-matter 与正文。
// 输入需以 "---\n" 开头，否则认为没有 front-matter。
func Split(data []byte) (meta []byte, body []byte, hasMeta bool) {
	normalized := bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n"))
	if !bytes.HasPrefix(normalized, append(delimiter, '\n')) {
		return nil, data, false
	}
	rest := normalized[len(delimiter)+1:]
	idx := bytes.Index(rest, append(append([]byte{'\n'}, delimiter...), '\n'))
	if idx < 0 {
		return nil, data, false
	}
	meta = rest[:idx]
	body = rest[idx+1+len(delimiter)+1:]
	// 去除 front-matter 与正文之间约定的一行空行，使 Compose/Split 可精确往返。
	if len(body) > 0 && body[0] == '\n' {
		body = body[1:]
	}
	return meta, body, true
}

// Compose 合成带 front-matter 的 Markdown 文档。
func Compose(meta []byte, body string) []byte {
	var buf bytes.Buffer
	buf.Write(delimiter)
	buf.WriteByte('\n')
	buf.Write(bytes.TrimRight(meta, "\n"))
	buf.WriteByte('\n')
	buf.Write(delimiter)
	buf.WriteByte('\n')
	buf.WriteByte('\n')
	buf.WriteString(body)
	if len(body) > 0 && body[len(body)-1] != '\n' {
		buf.WriteByte('\n')
	}
	return buf.Bytes()
}
