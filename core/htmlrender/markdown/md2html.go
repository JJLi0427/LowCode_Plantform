package markdown

import (
	"fmt"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

func Markdown2Html(md []byte, CompletePage bool) ([]byte, error) {
	if len(md) == 0 {
		return nil, fmt.Errorf("empty input markdown text")
	}
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs
	parser := parser.NewWithExtensions(extensions)

	htmlFlags := html.CommonFlags | html.HrefTargetBlank

	if CompletePage {
		htmlFlags = htmlFlags | html.CompletePage
	}

	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	html := markdown.ToHTML(md, parser, renderer)

	return html, nil
}
