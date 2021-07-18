// Package renderer renders the given AST to certain formats.
package renderer

import (
	"bytes"
	"io"

	goldast "github.com/yuin/goldmark/ast"
	goldrenderer "github.com/yuin/goldmark/renderer"
)

type Context struct {
	Env map[string]string
}

func NewContext() *Context {
	return &Context{
		Env: map[string]string{},
	}
}

func (c *Context) ImportEnv(env map[string]string) {
	for k, v := range env {
		c.Env[k] = v
	}
}

type RundownRenderer struct {
	actualRenderer goldrenderer.Renderer
	Section        string
	Context        *Context
}

// NewRenderer returns a new Renderer with given options.
func NewRundownRenderer(actualRenderer goldrenderer.Renderer, context *Context) *RundownRenderer {

	r := &RundownRenderer{
		actualRenderer: actualRenderer,
		Context:        context,
	}

	return r
}

func (r *RundownRenderer) AddOptions(opts ...goldrenderer.Option) {
	// for _, opt := range opts {
	// 	opt.SetConfig(r.config)
	// }
}

// WalkJump indicates that the walker should jump to the next node.
const WalkJump goldast.WalkStatus = 100

type JumpError struct {
	ToNode       goldast.Node
	ReturnToNode goldast.Node
}

func (e JumpError) Error() string {
	return "Jump compatible NodeWalker required"
}

type GlamourFlushNode struct{ goldast.BaseBlock }

func (n *GlamourFlushNode) Kind() goldast.NodeKind { return goldast.KindDocument }

// Render will individually render the child (block) nodes.
// This is required as Glamour buffers block node renders until the end of the document, which
// means execution blocks are run before any output is seen.
func (r *RundownRenderer) Render(w io.Writer, source []byte, n goldast.Node) error {
	if doc, ok := n.(*goldast.Document); ok {

		startingNode := doc.FirstChild()

		for child := startingNode; child != nil; child = child.NextSibling() {
			outputBuffer := &bytes.Buffer{}

			err := r.actualRenderer.Render(outputBuffer, source, child)
			if err != nil {
				return err
			}

			// Glamour pads lines so they're the full width of the terminal. We don't want that, so trim it all.
			for _, line := range bytes.Split(outputBuffer.Bytes(), []byte("\n")) {
				w.Write(bytes.TrimRight(line, "  "))
				w.Write([]byte("\n"))
			}

		}
	}
	return nil
}
