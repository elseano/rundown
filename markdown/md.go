package markdown

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
)

func PrepareMarkdown() goldmark.Markdown {
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			CodeModifiers,
			InvisibleBlocks,
			ConsoleRenderer,
			extension.Strikethrough,
			Emoji,
			// CodeExecute,
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
	)

	return md
}
