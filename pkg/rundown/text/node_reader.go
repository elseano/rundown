package text

import (
	"io"

	goldast "github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

type NodeReader struct {
	io.Reader

	node           goldast.Node
	segments       *text.Segments
	source         []byte
	currentSegment int
	posInSegment   int
}

func NewNodeReaderFromSource(node goldast.Node, source []byte) *NodeReader {
	r := &NodeReader{
		node:           node,
		source:         source,
		currentSegment: 0,
		posInSegment:   0,
	}

	r.segments = r.getSegments()
	return r
}

func NewNodeReader(node goldast.Node, reader text.Reader) *NodeReader {
	r := &NodeReader{
		node:           node,
		source:         reader.Source(),
		currentSegment: 0,
		posInSegment:   0,
	}

	r.segments = r.getSegments()
	return r
}

func (r *NodeReader) getSegments() *text.Segments {
	if r.node.Type() == goldast.TypeBlock {
		return r.node.Lines()
	} else if r.node.Type() == goldast.TypeDocument {
		return r.node.Lines()
	} else {

		switch v := r.node.(type) {
		case *goldast.RawHTML:
			return v.Segments
		default:
			return text.NewSegments()
		}

	}
}

func (r *NodeReader) Read(p []byte) (n int, err error) {
	if r.currentSegment > r.segments.Len()-1 {
		return 0, io.EOF
	}

	currentSegment := r.segments.At(r.currentSegment)
	segmentValue := currentSegment.Value(r.source)

	endRead := r.posInSegment + len(p)
	if endRead > len(segmentValue) {
		endRead = len(segmentValue)
	}

	n = copy(p, segmentValue[r.posInSegment:endRead])

	// If we have reached the end of the current segment, move to the next segment.
	if r.posInSegment+n >= len(segmentValue) {
		r.currentSegment++
		r.posInSegment = 0
	} else {
		// We +1 here because we want to start reading after the last read byte.
		r.posInSegment = r.posInSegment + n
	}

	return n, nil
}
