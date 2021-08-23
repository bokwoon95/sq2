package stitchdocs

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"sort"
	"testing"
	"text/template"

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
	p := parser.NewParser(parser.WithBlockParsers(parser.DefaultBlockParsers()...),
		parser.WithInlineParsers(parser.DefaultInlineParsers()...),
		parser.WithParagraphTransformers(parser.DefaultParagraphTransformers()...),
	)
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			highlighting.Highlighting,
		),
		goldmark.WithParser(p),
		goldmark.WithParserOptions(
			parser.WithAttribute(),
		),
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
	)
	b, err := fs.ReadFile(embeddedFiles, "quickstart.md")
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	var buf bytes.Buffer
	if err := md.Convert(b, &buf); err != nil {
		panic(err)
	}
	err = os.WriteFile("out.html", buf.Bytes(), 0666)
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
}

func TestTmpl(t *testing.T) {
	b, err := fs.ReadFile(embeddedFiles, "01 Quickstart.md")
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	tmpl, err := template.New("").Parse(string(b))
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	_ = tmpl.Tree.Root.Nodes
	tmpls := tmpl.Templates()
	sort.Slice(tmpls, func(i, j int) bool { return tmpls[i].Name() < tmpls[j].Name() })
	fmt.Println()
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, nil)
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	rawMarkdown := []byte(buf.String())
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM, highlighting.Highlighting),
		goldmark.WithRendererOptions(html.WithUnsafe()),
	)
	buf.Reset()
	if err := md.Convert(rawMarkdown, &buf); err != nil {
		panic(err)
	}
	err = os.WriteFile("out.html", buf.Bytes(), 0666)
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
}
