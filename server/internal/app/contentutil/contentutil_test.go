package contentutil

import (
	"strings"
	"testing"
)

func TestBuildSummaryUsesExplicitSummary(t *testing.T) {
	got := BuildSummary("  自定义摘要  ", "# 标题\n\n正文")
	if got != "自定义摘要" {
		t.Fatalf("expected explicit summary to win, got %q", got)
	}
}

func TestBuildSummaryExtractsFirstParagraphFromMarkdown(t *testing.T) {
	content := "# 标题\n\n这是第一段，有 **强调**、[链接](https://example.com) 和 `inline code`。\n\n第二段不该进入摘要。"
	got := BuildSummary("", content)
	want := "这是第一段，有 强调、链接 和 inline code。"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestBuildSummarySkipsCodeBlocksAndImages(t *testing.T) {
	content := "![cover](cover.jpg)\n\n```ts\nconst hidden = true\n```\n\n真正应该被提取的正文。"
	got := BuildSummary("", content)
	want := "真正应该被提取的正文。"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestBuildSummaryFallsBackToListText(t *testing.T) {
	content := "- 第一条要点\n- 第二条要点"
	got := BuildSummary("", content)
	want := "第一条要点"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestBuildSummaryTruncatesExtractedText(t *testing.T) {
	content := strings.Repeat("摘要", 120)
	got := BuildSummary("", content)
	if len([]rune(got)) != defaultSummaryRuneLimit {
		t.Fatalf("expected %d runes, got %d", defaultSummaryRuneLimit, len([]rune(got)))
	}
}
