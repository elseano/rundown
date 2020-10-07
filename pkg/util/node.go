package util

import (
	"github.com/yuin/goldmark/ast"
)

func NodeLines(v ast.Node, source []byte) string {
	text := string(v.Text(source))
	if text != "" {
		return text
	}

	// We can't walk the lines for Inline elements, so just return empty string.
	if v.Type() == ast.TypeInline {
		return ""
	}

	var result = ""

	for i := 0; i < v.Lines().Len(); i++ {
		line := v.Lines().At(i)
		result = result + string(line.Value(source))
	}

	return result
}
