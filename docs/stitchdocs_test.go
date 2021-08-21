package stitchdocs

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"testing"

	"github.com/bokwoon95/sq/internal/testutil"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

//go:embed *.md
var embeddedFiles embed.FS

func TestMD(t *testing.T) {
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			highlighting.Highlighting,
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
		),
	)
	b, err := fs.ReadFile(embeddedFiles, "docs.md")
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	var buf bytes.Buffer
	if err := md.Convert(b, &buf); err != nil {
		panic(err)
	}
	fmt.Println(buf.String())
}
