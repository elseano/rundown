// Package renderer renders the given AST to certain formats.
package renderer

import (
	"bufio"
	"io"
	"sync"

	"github.com/yuin/goldmark/ast"
	goldren "github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

type flushingRenderer struct {
	config               *goldren.Config
	options              map[goldren.OptionName]interface{}
	nodeRendererFuncsTmp map[ast.NodeKind]goldren.NodeRendererFunc
	maxKind              int
	nodeRendererFuncs    []goldren.NodeRendererFunc
	initSync             sync.Once
}

// NewRenderer returns a new Renderer with given options.
func NewFlushingRenderer(options ...goldren.Option) goldren.Renderer {
	config := goldren.NewConfig()
	for _, opt := range options {
		opt.SetConfig(config)
	}

	r := &flushingRenderer{
		options:              map[goldren.OptionName]interface{}{},
		config:               config,
		nodeRendererFuncsTmp: map[ast.NodeKind]goldren.NodeRendererFunc{},
	}

	return r
}

func (r *flushingRenderer) AddOptions(opts ...goldren.Option) {
	for _, opt := range opts {
		opt.SetConfig(r.config)
	}
}

func (r *flushingRenderer) Register(kind ast.NodeKind, v goldren.NodeRendererFunc) {
	r.nodeRendererFuncsTmp[kind] = v
	if int(kind) > r.maxKind {
		r.maxKind = int(kind)
	}
}

// Render renders the given AST node to the given writer with the given Renderer.
func (r *flushingRenderer) Render(w io.Writer, source []byte, n ast.Node) error {
	r.initSync.Do(func() {
		r.options = r.config.Options
		r.config.NodeRenderers.Sort()
		l := len(r.config.NodeRenderers)
		for i := l - 1; i >= 0; i-- {
			v := r.config.NodeRenderers[i]
			nr, _ := v.Value.(goldren.NodeRenderer)
			if se, ok := v.Value.(goldren.SetOptioner); ok {
				for oname, ovalue := range r.options {
					se.SetOption(oname, ovalue)
				}
			}
			nr.RegisterFuncs(r)
		}
		r.nodeRendererFuncs = make([]goldren.NodeRendererFunc, r.maxKind+1)
		for kind, nr := range r.nodeRendererFuncsTmp {
			r.nodeRendererFuncs[kind] = nr
		}
		r.config = nil
		r.nodeRendererFuncsTmp = nil
	})
	writer, ok := w.(util.BufWriter)
	if !ok {
		writer = bufio.NewWriter(w)
	}
	err := ast.Walk(n, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		s := ast.WalkStatus(ast.WalkContinue)
		var err error
		f := r.nodeRendererFuncs[n.Kind()]
		if f != nil {
			s, err = f(writer, source, n, entering)
		}
		writer.Flush() // Flush early and often
		return s, err
	})
	if err != nil {
		return err
	}
	return writer.Flush()
}
