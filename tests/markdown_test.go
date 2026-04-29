package tests

import (
	"strings"
	"testing"

	"github.com/claude/blog/internal/render"
)

func TestRenderMarkdown_Paragraph(t *testing.T) {
	out, err := render.RenderMarkdown("Hello, world.")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "<p>Hello, world.</p>") {
		t.Errorf("expected paragraph tag, got: %s", out)
	}
}

func TestRenderMarkdown_Table(t *testing.T) {
	src := "| A | B |\n|---|---|\n| 1 | 2 |\n"
	out, err := render.RenderMarkdown(src)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "<table>") {
		t.Errorf("expected <table>, got: %s", out)
	}
}

func TestRenderMarkdown_Strikethrough(t *testing.T) {
	out, err := render.RenderMarkdown("~~deleted~~")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "<del>deleted</del>") {
		t.Errorf("expected <del>, got: %s", out)
	}
}

func TestRenderMarkdown_TaskList(t *testing.T) {
	src := "- [ ] todo\n- [x] done\n"
	out, err := render.RenderMarkdown(src)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, `type="checkbox"`) {
		t.Errorf("expected checkbox input, got: %s", out)
	}
}

func TestRenderMarkdown_StripScript(t *testing.T) {
	out, err := render.RenderMarkdown("<script>alert('xss')</script>")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out, "<script>") {
		t.Errorf("safe mode should strip <script>, got: %s", out)
	}
}

func TestRenderMarkdownUnsafe_AllowsRawHTML(t *testing.T) {
	out, err := render.RenderMarkdownUnsafe("<em>raw</em>")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "<em>raw</em>") {
		t.Errorf("unsafe mode should pass raw HTML through, got: %s", out)
	}
}
