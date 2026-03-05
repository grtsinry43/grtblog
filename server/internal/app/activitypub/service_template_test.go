package activitypub

import (
	"strings"
	"testing"
)

func TestRenderFederatedHTMLTemplate(t *testing.T) {
	html := renderFederatedHTML(`<p>{{ .ContentType }}: {{ .Title }}</p><p>{{ .Summary }}</p><a href="{{ .URL }}">go</a>`, "标题", "摘要", "https://example.com/p/1", "article")
	if !strings.Contains(html, "文章: 标题") {
		t.Fatalf("unexpected content type render: %s", html)
	}
	if !strings.Contains(html, "https://example.com/p/1") {
		t.Fatalf("unexpected url render: %s", html)
	}
}

func TestActivityPubContentTypeLabel(t *testing.T) {
	cases := map[string]string{
		"article":  "文章",
		"moment":   "手记",
		"thinking": "思考",
		"unknown":  "",
	}
	for input, expect := range cases {
		if got := activityPubContentTypeLabel(input); got != expect {
			t.Fatalf("input %s got %s expect %s", input, got, expect)
		}
	}
}
