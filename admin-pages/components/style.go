package components

import (
	"context"
	"io"
	"strings"

	"github.com/a-h/templ"
)

type Theme struct {
	PrimaryColor    string
	SecondaryColor  string
	BackgroundColor string
	TextColor       string
}

func StyleTag(t Theme, style string) templ.Component {
	return templ.ComponentFunc(
		func(ctx context.Context, w io.Writer) (err error) {
			str := strings.Builder{}

			str.WriteString("<style type=\"text/css\">")
			str.WriteString(":root {")
			str.WriteString("--clr-primary:" + t.PrimaryColor + ";")
			str.WriteString("--clr-secondary: " + t.SecondaryColor + ";")
			str.WriteString("--clr-background: " + t.BackgroundColor + ";")
			str.WriteString("--clr-text: " + t.TextColor + ";")
			str.WriteString("}</style>")
			
			str.WriteString("<style type=\"text/css\">")
			str.WriteString(style)
			str.WriteString("}</style>")

			io.WriteString(w, str.String())

			return nil
		})
}
