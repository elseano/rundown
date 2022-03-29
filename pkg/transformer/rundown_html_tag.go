package transformer

import (
	"strings"

	"github.com/elseano/rundown/pkg/text"
	goldast "github.com/yuin/goldmark/ast"
	goldtext "github.com/yuin/goldmark/text"
	"golang.org/x/net/html"
)

type RundownHtmlTag struct {
	tag      string
	attrs    []html.Attribute
	contents string
	closed   bool
	closer   bool
}

func ExtractRundownElement(node goldast.Node, reader goldtext.Reader, currentTag string) *RundownHtmlTag {
	z := html.NewTokenizerFragment(text.NewNodeReader(node, reader), currentTag)
	// source := string(node.Text(reader.Source()))
	// z := html.NewTokenizer(strings.NewReader(source))

	var currentRundownTag *RundownHtmlTag

	for {
		ttype := z.Next()
		token := z.Token()
		switch ttype {
		case html.StartTagToken, html.SelfClosingTagToken:
			if token.Data == "r" || token.Data == "rundown" {
				currentRundownTag = &RundownHtmlTag{
					tag:    token.Data,
					closed: false,
				}
			}

			if currentRundownTag != nil {
				if token.Attr != nil {
					currentRundownTag.attrs = make([]html.Attribute, len(token.Attr))
					copy(currentRundownTag.attrs, token.Attr)
				}

				if ttype == html.SelfClosingTagToken {
					currentRundownTag.closed = true
					return currentRundownTag
				}
			}
		case html.TextToken:
			if currentRundownTag != nil {
				currentRundownTag.contents = currentRundownTag.contents + token.Data
			}
		case html.EndTagToken:
			if currentRundownTag != nil {
				currentRundownTag.closed = true
				return currentRundownTag
			}

			// Otherwise, we might have the closing tag of an earlier opened tag.
			if token.Data == "r" || strings.HasPrefix(token.Data, "r-") {
				return &RundownHtmlTag{
					tag:    token.Data,
					closer: true,
				}
			}

			return nil

		// ErrorToken is expected for inline RawHTML nodes, as they don't contain the entire HTML element,
		// instead there's a RawHTML for the opening, and a RawHTML for the closing tag.
		case html.ErrorToken:
			return currentRundownTag
		}

	}
}
