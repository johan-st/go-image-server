package components

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/a-h/templ"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
)

var (
	md   goldmark.Markdown
	once sync.Once
)

func markdownFile(file []byte) templ.Component {
	// setup markdown parser
	once.Do(func() {
		md = goldmark.New(
			goldmark.WithExtensions(
				highlighting.NewHighlighting(
					highlighting.WithStyle("native"),
				),
				extension.GFM,
				// md_meta.Meta,
			),
			goldmark.WithRendererOptions(
				html.WithUnsafe(),
			),
		)
	})

	return templ.ComponentFunc(
		func(ctx context.Context, w io.Writer) (err error) {
			fmt.Println("markdownFile rendering... TODO: only call this once")
			return md.Convert(file, w)
		})
}
