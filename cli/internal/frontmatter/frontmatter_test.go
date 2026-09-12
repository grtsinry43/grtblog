package frontmatter

import (
	"strings"
	"testing"
)

func TestSplit(t *testing.T) {
	doc := "---\ntitle: 你好\npublished: false\n---\n\n正文内容\n第二行\n"
	meta, body, has := Split([]byte(doc))
	if !has {
		t.Fatal("expected front-matter")
	}
	if !strings.Contains(string(meta), "title: 你好") {
		t.Errorf("meta = %q", meta)
	}
	if !strings.Contains(string(body), "正文内容") {
		t.Errorf("body = %q", body)
	}
}

func TestSplitNoFrontMatter(t *testing.T) {
	doc := "# 标题\n\n正文"
	meta, body, has := Split([]byte(doc))
	if has {
		t.Fatal("did not expect front-matter")
	}
	if meta != nil {
		t.Errorf("meta = %q, want nil", meta)
	}
	if string(body) != doc {
		t.Errorf("body = %q", body)
	}
}

func TestSplitCRLF(t *testing.T) {
	doc := "---\r\ntitle: hi\r\n---\r\nbody\r\n"
	_, body, has := Split([]byte(doc))
	if !has {
		t.Fatal("expected front-matter with CRLF")
	}
	if !strings.Contains(string(body), "body") {
		t.Errorf("body = %q", body)
	}
}

func TestComposeRoundTrip(t *testing.T) {
	meta := []byte("title: 测试\npublished: true")
	body := "正文\n"
	doc := Compose(meta, body)
	gotMeta, gotBody, has := Split(doc)
	if !has {
		t.Fatal("composed doc has no front-matter")
	}
	if strings.TrimSpace(string(gotMeta)) != string(meta) {
		t.Errorf("meta round trip = %q", gotMeta)
	}
	if string(gotBody) != body {
		t.Errorf("body round trip = %q", gotBody)
	}
}
