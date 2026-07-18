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

func TestGenerateTOCOnlyIncludesHeadingNodes(t *testing.T) {
	content := strings.Join([]string{
		"# 正常标题",
		"",
		"普通段落里的 # 井号不是标题。",
		"",
		"`# 行内代码也不是标题`",
		"",
		"```markdown",
		"# 围栏代码块里的伪标题",
		"```",
		"",
		"    # 缩进代码块里的伪标题",
		"",
		"| 列一 | 列二 |",
		"| --- | --- |",
		"| # 表格里的井号 | 内容 |",
		"",
		"<div># HTML 块里的井号</div>",
		"",
		"## 子标题",
	}, "\n")

	got := GenerateTOC(content)
	if len(got) != 1 {
		t.Fatalf("expected one root heading, got %#v", got)
	}
	if got[0].Name != "正常标题" {
		t.Fatalf("expected root heading %q, got %q", "正常标题", got[0].Name)
	}
	if len(got[0].Children) != 1 || got[0].Children[0].Name != "子标题" {
		t.Fatalf("expected only the real child heading, got %#v", got[0].Children)
	}
}
