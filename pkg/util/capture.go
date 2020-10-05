package util

import (
	"bytes"
	"io"
	"os"

	"github.com/yuin/goldmark/ast"
)

func CaptureStdout(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func DumpNode(node ast.Node, source []byte) string {
	return CaptureStdout(func() { node.Dump(source, 1) })
}
