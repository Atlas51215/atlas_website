package render

import (
	"bytes"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
)

var mdSafe = goldmark.New(
	goldmark.WithExtensions(extension.GFM),
)

var mdUnsafe = goldmark.New(
	goldmark.WithExtensions(extension.GFM),
	goldmark.WithRendererOptions(html.WithUnsafe()),
)

// RenderMarkdown converts Markdown to HTML with GFM extensions.
// Raw HTML in the source is stripped to prevent XSS.
func RenderMarkdown(src string) (string, error) {
	var buf bytes.Buffer
	if err := mdSafe.Convert([]byte(src), &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// RenderMarkdownUnsafe converts Markdown to HTML with GFM extensions,
// allowing raw HTML passthrough. Use only for trusted curator content.
func RenderMarkdownUnsafe(src string) (string, error) {
	var buf bytes.Buffer
	if err := mdUnsafe.Convert([]byte(src), &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}
