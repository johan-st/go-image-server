package components

import (
	"context"
	"errors"
	"io"
	"strings"

	"github.com/a-h/templ"
)

type Theme struct {
	// colors
	ColorPrimary    string
	ColorSecondary  string
	ColorBackground string
	ColorText       string
	ColorBorder     string

	// border
	BorderRadius string
}

func StyleTag(t Theme, style string) (templ.Component, error) {
	if t.ColorPrimary == "" {
		return nil, errors.New("ColorPrimary is required")
	}
	if t.ColorSecondary == "" {
		return nil, errors.New("ColorSecondary is required")
	}
	if t.ColorBackground == "" {
		return nil, errors.New("ColorBackground is required")
	}
	if t.ColorText == "" {
		return nil, errors.New("ColorText is required")
	}
	if t.ColorBorder == "" {
		return nil, errors.New("ColorBorder is required")
	}
	if t.BorderRadius == "" {
		return nil, errors.New("BorderRadius is required")
	}
	

	return templ.ComponentFunc(
		func(ctx context.Context, w io.Writer) (err error) {
			str := strings.Builder{}

			str.WriteString("<style type=\"text/css\">")
			str.WriteString(":root {")
			str.WriteString("--clr-primary:" + t.ColorPrimary + ";")
			str.WriteString("--clr-secondary: " + t.ColorSecondary + ";")
			str.WriteString("--clr-background: " + t.ColorBackground + ";")
			str.WriteString("--clr-text: " + t.ColorText + ";")
			str.WriteString("--clr-border: " + t.ColorBorder + ";")
			str.WriteString("--border-radius: " + t.BorderRadius + ";")
			str.WriteString("}</style>")

			str.WriteString("<style type=\"text/css\">")
			str.WriteString(style)
			str.WriteString("}</style>")

			io.WriteString(w, str.String())

			return nil
		}),nil
}
