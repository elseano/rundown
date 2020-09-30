package util

import (
	"github.com/yuin/goldmark/ast"
)

func NodeLines(v ast.Node, source []byte) string {
	if v.Type() == ast.TypeInline {
		return string(v.Text(source))
	}

	var result = ""

	for i := 0; i < v.Lines().Len(); i++ {
		line := v.Lines().At(i)
		result = result + string(line.Value(source))
	}

	return result
}
